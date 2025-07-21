package service

import (
	"evolution-postgres-backup/internal/config"
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/models"
	"fmt"
	"os"
	"time"
)

// DatabaseService integrates all PostgreSQL repositories and provides high-level operations
type DatabaseService struct {
	db           *database.DB
	backupRepo   *database.BackupRepository
	postgresRepo *database.PostgreSQLRepository
	logRepo      *database.LogRepository
	migrationSvc *database.MigrationService
}

// NewDatabaseService creates a new integrated database service
func NewDatabaseService() (*DatabaseService, error) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "backup-data"
	}

	db, err := database.NewDB(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	backupRepo := database.NewBackupRepository(db)
	postgresRepo := database.NewPostgreSQLRepository(db)
	logRepo := database.NewLogRepository(db)
	migrationSvc := database.NewMigrationService(db, dataDir)

	return &DatabaseService{
		db:           db,
		backupRepo:   backupRepo,
		postgresRepo: postgresRepo,
		logRepo:      logRepo,
		migrationSvc: migrationSvc,
	}, nil
}

// Close closes the database connection
func (s *DatabaseService) Close() error {
	return s.db.Close()
}

// ==================== PostgreSQL Instance Management ====================

// GetPostgreSQLInstances returns all PostgreSQL instances
func (s *DatabaseService) GetPostgreSQLInstances() ([]*config.PostgreSQLConfig, error) {
	return s.postgresRepo.GetAll()
}

// GetEnabledPostgreSQLInstances returns only enabled PostgreSQL instances
func (s *DatabaseService) GetEnabledPostgreSQLInstances() ([]*config.PostgreSQLConfig, error) {
	return s.postgresRepo.GetEnabled()
}

// GetPostgreSQLInstance returns a specific PostgreSQL instance
func (s *DatabaseService) GetPostgreSQLInstance(id string) (*config.PostgreSQLConfig, error) {
	return s.postgresRepo.GetByID(id)
}

// CreatePostgreSQLInstance creates a new PostgreSQL instance
func (s *DatabaseService) CreatePostgreSQLInstance(instance *config.PostgreSQLConfig) error {
	// Generate ID if not provided
	if instance.ID == "" {
		instance.ID = generateID()
	}

	return s.postgresRepo.Create(instance)
}

// UpdatePostgreSQLInstance updates an existing PostgreSQL instance
func (s *DatabaseService) UpdatePostgreSQLInstance(instance *config.PostgreSQLConfig) error {
	return s.postgresRepo.Update(instance)
}

// DeletePostgreSQLInstance deletes a PostgreSQL instance
func (s *DatabaseService) DeletePostgreSQLInstance(id string) error {
	return s.postgresRepo.Delete(id)
}

// ==================== Backup Management ====================

// GetBackups returns backups with optional filters
func (s *DatabaseService) GetBackups(filters ...database.BackupFilter) ([]*models.BackupInfo, error) {
	return s.backupRepo.GetAll(filters...)
}

// GetBackup returns a specific backup
func (s *DatabaseService) GetBackup(id string) (*models.BackupInfo, error) {
	return s.backupRepo.GetByID(id)
}

// CreateBackup creates a new backup record
func (s *DatabaseService) CreateBackup(backup *models.BackupInfo) error {
	return s.backupRepo.Create(backup)
}

// UpdateBackup updates an existing backup record
func (s *DatabaseService) UpdateBackup(backup *models.BackupInfo) error {
	return s.backupRepo.Update(backup)
}

// GetBackupsByInstance returns backups for a specific PostgreSQL instance
func (s *DatabaseService) GetBackupsByInstance(postgresID string) ([]*models.BackupInfo, error) {
	return s.backupRepo.GetByPostgreSQLID(postgresID)
}

// GetBackupsByStatus returns backups with a specific status
func (s *DatabaseService) GetBackupsByStatus(status models.BackupStatus) ([]*models.BackupInfo, error) {
	return s.backupRepo.GetByStatus(status)
}

// GetOldBackupsForCleanup returns old backups for cleanup
func (s *DatabaseService) GetOldBackupsForCleanup(postgresID string, backupType models.BackupType, olderThan time.Time) ([]*models.BackupInfo, error) {
	return s.backupRepo.GetOldBackups(postgresID, backupType, olderThan)
}

// DeleteOldBackups removes old backups based on retention policy
func (s *DatabaseService) DeleteOldBackups(postgresID string, backupType models.BackupType, olderThan time.Time) (int64, error) {
	return s.backupRepo.DeleteOldBackups(postgresID, backupType, olderThan)
}

// ==================== Log Management ====================

// GetLogs returns logs with filters
func (s *DatabaseService) GetLogs(filters database.LogFilters) ([]*database.LogEntry, error) {
	return s.logRepo.GetFiltered(filters)
}

// GetLogsByJobID returns logs for a specific job
func (s *DatabaseService) GetLogsByJobID(jobID string) ([]*database.LogEntry, error) {
	return s.logRepo.GetByJobID(jobID)
}

// GetLogsByBackupID returns logs for a specific backup
func (s *DatabaseService) GetLogsByBackupID(backupID string) ([]*database.LogEntry, error) {
	return s.logRepo.GetByBackupID(backupID)
}

// GetRecentLogs returns the most recent logs
func (s *DatabaseService) GetRecentLogs(limit int) ([]*database.LogEntry, error) {
	return s.logRepo.GetRecent(limit)
}

// CreateLog creates a new log entry
func (s *DatabaseService) CreateLog(entry *database.LogEntry) error {
	return s.logRepo.Create(entry)
}

// CreateLogBatch creates multiple log entries efficiently
func (s *DatabaseService) CreateLogBatch(entries []*database.LogEntry) error {
	return s.logRepo.CreateBatch(entries)
}

// ==================== Statistics & Reports ====================

// GetDashboardStats returns comprehensive dashboard statistics
func (s *DatabaseService) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// PostgreSQL instances stats
	pgStats, err := s.postgresRepo.GetStats()
	if err != nil {
		return nil, err
	}
	stats["postgresql"] = pgStats

	// Backup stats
	backupStats, err := s.backupRepo.GetStats()
	if err != nil {
		return nil, err
	}
	stats["backups"] = backupStats

	// Log stats
	logStats, err := s.logRepo.GetStats()
	if err != nil {
		return nil, err
	}
	stats["logs"] = logStats

	// Database stats
	dbStats, err := s.db.GetStats()
	if err != nil {
		return nil, err
	}
	stats["database"] = dbStats

	return stats, nil
}

// GetBackupTrends returns backup trends over time
func (s *DatabaseService) GetBackupTrends(days int) (map[string]interface{}, error) {
	// This would be implemented with more complex SQL queries
	// For now, return basic stats
	return s.backupRepo.GetStats()
}

// ==================== Health Checks ====================

// HealthCheck performs a comprehensive health check
func (s *DatabaseService) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})

	// Database connectivity
	if err := s.db.Ping(); err != nil {
		health["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		health["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// PostgreSQL instances
	instances, err := s.postgresRepo.GetEnabled()
	if err != nil {
		health["postgresql_instances"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["postgresql_instances"] = map[string]interface{}{
			"status": "healthy",
			"count":  len(instances),
		}
	}

	// Recent backups
	recentBackups, err := s.backupRepo.GetAll()
	if err != nil {
		health["backups"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["backups"] = map[string]interface{}{
			"status": "healthy",
			"total":  len(recentBackups),
		}
	}

	return health
}

// ==================== Migration Support ====================

// GetMigrationStatus returns migration status
func (s *DatabaseService) GetMigrationStatus() (map[string]interface{}, error) {
	return s.migrationSvc.GetMigrationStatus()
}

// PerformMigration performs migration from JSON to SQLite
func (s *DatabaseService) PerformMigration() error {
	return s.migrationSvc.MigrateAll()
}

// ==================== Utility Functions ====================

// GetDB returns the underlying database connection (for worker integration)
func (s *DatabaseService) GetDB() *database.DB {
	return s.db
}

func generateID() string {
	return fmt.Sprintf("pg-%d", time.Now().Unix())
}
