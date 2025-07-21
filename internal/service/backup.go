package service

import (
	"evolution-postgres-backup/internal/config"
	"evolution-postgres-backup/internal/logger"
	"evolution-postgres-backup/internal/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BackupService struct {
	s3Client    *S3Client
	config      *config.Config
	backups     map[string]*models.BackupInfo
	tempDir     string
	persistence *BackupPersistence
}

func NewBackupService(s3Client *S3Client, cfg *config.Config) *BackupService {
	tempDir := os.Getenv("BACKUP_TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/postgres-backups"
	}

	// Create temp directory if it doesn't exist
	os.MkdirAll(tempDir, 0755)

	// Initialize persistence
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/app/data"
	}
	persistence := NewBackupPersistence(dataDir)

	// Load existing backups from disk
	backups, err := persistence.LoadBackups()
	if err != nil {
		fmt.Printf("Warning: failed to load existing backups: %v\n", err)
		backups = make(map[string]*models.BackupInfo)
	} else {
		fmt.Printf("Loaded %d existing backups from disk\n", len(backups))
	}

	service := &BackupService{
		s3Client:    s3Client,
		config:      cfg,
		backups:     backups,
		tempDir:     tempDir,
		persistence: persistence,
	}

	// Initialize S3 client
	if err := s3Client.Initialize(cfg); err != nil {
		panic(fmt.Sprintf("Failed to initialize S3 client: %v", err))
	}

	// Clean up old backup files periodically
	go func() {
		time.Sleep(5 * time.Minute) // Wait 5 minutes after startup
		persistence.CleanupOldBackupFiles()
	}()

	return service
}

func (bs *BackupService) CreateBackup(postgresID string, backupType models.BackupType, databaseName string) (*models.BackupInfo, error) {
	pgConfig, exists := bs.config.GetPostgreSQLByID(postgresID)
	if !exists {
		return nil, fmt.Errorf("PostgreSQL instance %s not found", postgresID)
	}

	if !pgConfig.Enabled {
		return nil, fmt.Errorf("PostgreSQL instance %s is disabled", postgresID)
	}

	// Use provided database name or default from config
	if databaseName == "" {
		databaseName = pgConfig.GetDefaultDatabase()
	}

	backupID := uuid.New().String()
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s_%s_%s_%s.sql", pgConfig.Name, databaseName, string(backupType), timestamp)
	filename = strings.ReplaceAll(filename, " ", "_")

	backupInfo := &models.BackupInfo{
		ID:           backupID,
		PostgreSQLID: postgresID,
		DatabaseName: databaseName,
		BackupType:   backupType,
		Status:       models.BackupStatusPending,
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
	}

	bs.backups[backupID] = backupInfo

	// Save to disk
	if err := bs.persistence.SaveSingleBackup(bs.backups); err != nil {
		fmt.Printf("Warning: failed to save backup to disk: %v\n", err)
	}

	// Start backup in goroutine
	go bs.performBackup(backupInfo, pgConfig, filename, timestamp)

	return backupInfo, nil
}

func (bs *BackupService) performBackup(backupInfo *models.BackupInfo, pgConfig *config.PostgreSQLConfig, filename, timestamp string) {
	log := logger.GetLogger()
	jobID := backupInfo.ID[:8] // Short ID for logs

	log.LogJobStart(jobID, "BACKUP", fmt.Sprintf("Database: %s/%s", pgConfig.Name, backupInfo.DatabaseName))

	backupInfo.Status = models.BackupStatusInProgress
	log.LogJobProgress(jobID, "Status: IN_PROGRESS")

	// Save status update
	bs.persistence.SaveSingleBackup(bs.backups)

	// Create local backup file
	localPath := filepath.Join(bs.tempDir, filename)
	log.LogJobProgress(jobID, "Local file: %s", localPath)

	// Build pg_dump command
	cmd := exec.Command("pg_dump",
		"-h", pgConfig.Host,
		"-p", fmt.Sprintf("%d", pgConfig.Port),
		"-U", pgConfig.Username,
		"-d", backupInfo.DatabaseName,
		"-f", localPath,
		"--verbose",
		"--no-password",
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", pgConfig.Password))

	log.LogJobProgress(jobID, "Executing pg_dump: %s@%s:%d/%s", pgConfig.Username, pgConfig.Host, pgConfig.Port, backupInfo.DatabaseName)

	// Execute backup
	if err := cmd.Run(); err != nil {
		backupInfo.Status = models.BackupStatusFailed
		backupInfo.ErrorMessage = fmt.Sprintf("pg_dump failed: %v", err)
		endTime := time.Now()
		backupInfo.EndTime = &endTime
		log.LogJobError(jobID, "pg_dump failed: %v", err)

		// Save error status
		bs.persistence.SaveSingleBackup(bs.backups)
		return
	}

	log.LogJobProgress(jobID, "pg_dump completed successfully")

	// Get file size
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		backupInfo.Status = models.BackupStatusFailed
		backupInfo.ErrorMessage = fmt.Sprintf("failed to get file info: %v", err)
		endTime := time.Now()
		backupInfo.EndTime = &endTime
		log.LogJobError(jobID, "Failed to get file info: %v", err)

		// Save error status
		bs.persistence.SaveSingleBackup(bs.backups)
		return
	}

	backupInfo.FileSize = fileInfo.Size()
	backupInfo.FilePath = localPath
	log.LogJobProgress(jobID, "Backup file size: %d bytes (%.2f MB)", fileInfo.Size(), float64(fileInfo.Size())/1024/1024)

	// Generate S3 key
	s3Key := GenerateS3Key(backupInfo.PostgreSQLID, string(backupInfo.BackupType), timestamp, filename)
	backupInfo.S3Key = s3Key
	log.LogJobProgress(jobID, "S3 key: %s", s3Key)

	// Upload to S3
	log.LogJobProgress(jobID, "Starting S3 upload...")
	if err := bs.s3Client.UploadFile(localPath, s3Key); err != nil {
		backupInfo.Status = models.BackupStatusFailed
		backupInfo.ErrorMessage = fmt.Sprintf("S3 upload failed: %v", err)
		endTime := time.Now()
		backupInfo.EndTime = &endTime
		log.LogJobError(jobID, "S3 upload failed: %v", err)

		// Save error status
		bs.persistence.SaveSingleBackup(bs.backups)
		return
	}
	log.LogJobProgress(jobID, "S3 upload completed successfully")

	// Clean up local file
	os.Remove(localPath)
	log.LogJobProgress(jobID, "Local file cleaned up")

	// Mark as completed
	backupInfo.Status = models.BackupStatusCompleted
	endTime := time.Now()
	backupInfo.EndTime = &endTime

	duration := endTime.Sub(backupInfo.StartTime)
	log.LogJobSuccess(jobID, "Backup completed in %v (%.2f MB)", duration, float64(backupInfo.FileSize)/1024/1024)

	// Save completion status
	if err := bs.persistence.SaveSingleBackup(bs.backups); err != nil {
		log.LogJobProgress(jobID, "Warning: failed to save backup completion to disk: %v", err)
	}

	// Clean up old backups based on retention policy
	go bs.cleanupOldBackups(backupInfo.PostgreSQLID, backupInfo.BackupType)
}

func (bs *BackupService) RestoreBackup(backupID, postgresID, databaseName string) error {
	backupInfo, exists := bs.backups[backupID]
	if !exists {
		return fmt.Errorf("backup %s not found", backupID)
	}

	if backupInfo.Status != models.BackupStatusCompleted {
		return fmt.Errorf("backup %s is not completed", backupID)
	}

	pgConfig, exists := bs.config.GetPostgreSQLByID(postgresID)
	if !exists {
		return fmt.Errorf("PostgreSQL instance %s not found", postgresID)
	}

	// Download backup file from S3
	localPath := filepath.Join(bs.tempDir, fmt.Sprintf("restore_%s_%d.sql", backupID, time.Now().Unix()))

	if err := bs.s3Client.DownloadFile(backupInfo.S3Key, localPath); err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}
	defer os.Remove(localPath)

	// Build psql command for restore
	cmd := exec.Command("psql",
		"-h", pgConfig.Host,
		"-p", fmt.Sprintf("%d", pgConfig.Port),
		"-U", pgConfig.Username,
		"-d", databaseName,
		"-f", localPath,
		"--quiet",
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", pgConfig.Password))

	// Execute restore
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("psql restore failed: %w", err)
	}

	return nil
}

func (bs *BackupService) GetBackups() []*models.BackupInfo {
	backups := make([]*models.BackupInfo, 0, len(bs.backups))
	for _, backup := range bs.backups {
		backups = append(backups, backup)
	}
	return backups
}

func (bs *BackupService) GetBackup(backupID string) (*models.BackupInfo, bool) {
	backup, exists := bs.backups[backupID]
	return backup, exists
}

func (bs *BackupService) cleanupOldBackups(postgresID string, backupType models.BackupType) {
	var retentionCount int

	switch backupType {
	case models.BackupTypeHourly:
		retentionCount = bs.config.RetentionPolicy.Hourly
	case models.BackupTypeDaily:
		retentionCount = bs.config.RetentionPolicy.Daily
	case models.BackupTypeWeekly:
		retentionCount = bs.config.RetentionPolicy.Weekly
	case models.BackupTypeMonthly:
		retentionCount = bs.config.RetentionPolicy.Monthly
	default:
		return // Don't clean up manual backups
	}

	prefix := fmt.Sprintf("backups/%s/%s/", postgresID, string(backupType))
	if err := bs.s3Client.CleanupOldBackups(prefix, retentionCount); err != nil {
		fmt.Printf("Failed to cleanup old backups: %v\n", err)
	}
}

func (bs *BackupService) CreateBackupForAllEnabledInstances(backupType models.BackupType) []*models.BackupInfo {
	var results []*models.BackupInfo

	for _, pgConfig := range bs.config.PostgreSQLInstances {
		if pgConfig.Enabled {
			// Create backup for each database in this PostgreSQL instance
			databases := pgConfig.GetDatabases()
			for _, dbName := range databases {
				backup, err := bs.CreateBackup(pgConfig.ID, backupType, dbName)
				if err != nil {
					fmt.Printf("Failed to create backup for %s/%s: %v\n", pgConfig.ID, dbName, err)
					continue
				}
				results = append(results, backup)
			}
		}
	}

	return results
}
