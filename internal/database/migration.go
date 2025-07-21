package database

import (
	"bufio"
	"encoding/json"
	"evolution-postgres-backup/internal/config"
	"evolution-postgres-backup/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type MigrationService struct {
	db      *DB
	dataDir string
}

func NewMigrationService(db *DB, dataDir string) *MigrationService {
	return &MigrationService{
		db:      db,
		dataDir: dataDir,
	}
}

// MigrateAll performs complete migration from JSON to SQLite
func (m *MigrationService) MigrateAll() error {
	fmt.Println("🔄 Starting migration from JSON to SQLite...")

	// 1. Migrate config.json (PostgreSQL instances)
	if err := m.migrateConfig(); err != nil {
		return fmt.Errorf("failed to migrate config: %w", err)
	}

	// 2. Migrate backups.json
	if err := m.migrateBackups(); err != nil {
		return fmt.Errorf("failed to migrate backups: %w", err)
	}

	// 3. Migrate log files
	if err := m.migrateLogs(); err != nil {
		return fmt.Errorf("failed to migrate logs: %w", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	return nil
}

// migrateConfig migrates PostgreSQL instances from config.json
func (m *MigrationService) migrateConfig() error {
	fmt.Print("📋 Migrating PostgreSQL instances from config.json... ")

	configPath := filepath.Join(m.dataDir, "../config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.json" // Try current directory
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("⚠️  config.json not found, skipping")
		return nil
	}

	// Load existing config
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Migrate PostgreSQL instances
	postgresRepo := NewPostgreSQLRepository(m.db)
	migratedCount := 0

	for _, instance := range cfg.PostgreSQLInstances {
		// Check if already exists
		exists, err := postgresRepo.Exists(instance.ID)
		if err != nil {
			return err
		}

		if !exists {
			if err := postgresRepo.Create(&instance); err != nil {
				return fmt.Errorf("failed to create PostgreSQL instance %s: %w", instance.ID, err)
			}
			migratedCount++
		}
	}

	fmt.Printf("✅ (%d instances)\n", migratedCount)
	return nil
}

// migrateBackups migrates backup data from backups.json
func (m *MigrationService) migrateBackups() error {
	fmt.Print("💾 Migrating backups from backups.json... ")

	backupsPath := filepath.Join(m.dataDir, "backups.json")
	if _, err := os.Stat(backupsPath); os.IsNotExist(err) {
		fmt.Println("⚠️  backups.json not found, skipping")
		return nil
	}

	// Read and parse backups.json
	data, err := os.ReadFile(backupsPath)
	if err != nil {
		return err
	}

	var backups []*models.BackupInfo
	if err := json.Unmarshal(data, &backups); err != nil {
		return err
	}

	// Migrate backups
	backupRepo := NewBackupRepository(m.db)
	migratedCount := 0

	for _, backup := range backups {
		// Check if already exists
		existing, err := backupRepo.GetByID(backup.ID)
		if err == nil && existing != nil {
			continue // Already exists
		}

		if err := backupRepo.Create(backup); err != nil {
			return fmt.Errorf("failed to create backup %s: %w", backup.ID, err)
		}
		migratedCount++
	}

	fmt.Printf("✅ (%d backups)\n", migratedCount)
	return nil
}

// migrateLogs migrates log files to structured database logs
func (m *MigrationService) migrateLogs() error {
	fmt.Print("📝 Migrating log files... ")

	logsDir := os.Getenv("LOG_DIR")
	if logsDir == "" {
		logsDir = "logs"
	}

	// Check if logs directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		fmt.Println("⚠️  logs directory not found, skipping")
		return nil
	}

	// Find all log files
	pattern := filepath.Join(logsDir, "backup_*.log")
	logFiles, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(logFiles) == 0 {
		fmt.Println("⚠️  no log files found, skipping")
		return nil
	}

	logRepo := NewLogRepository(m.db)
	totalMigrated := 0

	for _, logFile := range logFiles {
		count, err := m.migrateLogFile(logFile, logRepo)
		if err != nil {
			fmt.Printf("⚠️  failed to migrate %s: %v\n", logFile, err)
			continue
		}
		totalMigrated += count
	}

	fmt.Printf("✅ (%d log entries from %d files)\n", totalMigrated, len(logFiles))
	return nil
}

// migrateLogFile migrates a single log file
func (m *MigrationService) migrateLogFile(logFilePath string, logRepo *LogRepository) (int, error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)

	// Extract date from filename (backup_2025-07-18.log)
	filename := filepath.Base(logFilePath)
	dateStr := strings.TrimPrefix(filename, "backup_")
	dateStr = strings.TrimSuffix(dateStr, ".log")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry := m.parseLogLine(line, dateStr)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	// Batch insert for efficiency
	if len(entries) > 0 {
		if err := logRepo.CreateBatch(entries); err != nil {
			return 0, err
		}
	}

	return len(entries), nil
}

// parseLogLine parses a log line into a structured LogEntry
func (m *MigrationService) parseLogLine(line, dateContext string) *LogEntry {
	// Log format examples:
	// 2025/07/18 08:41:03 [2025-07-18 08:41:03] [INFO] [MAIN] 🚀 PostgreSQL Backup Service starting...
	// 2025/07/18 08:41:03 [2025-07-18 08:41:03] [INFO] [JOB] [68196c55] Starting pg_dump: postgres@manager.chatpolos.com.br:5432/evolution_lb

	// Regex to parse different log formats
	patterns := []struct {
		name    string
		regex   *regexp.Regexp
		handler func(matches []string, dateContext string) *LogEntry
	}{
		{
			name:  "job_log",
			regex: regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+\[([^\]]+)\]\s+\[([^\]]+)\]\s+\[JOB\]\s+\[([^\]]+)\]\s+(.*)$`),
			handler: func(matches []string, dateContext string) *LogEntry {
				timestamp := m.parseTimestamp(matches[2], dateContext)
				return &LogEntry{
					Timestamp: timestamp,
					Level:     matches[3],
					Component: "JOB",
					JobID:     matches[4],
					Message:   matches[5],
				}
			},
		},
		{
			name:  "standard_log",
			regex: regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+\[([^\]]+)\]\s+\[([^\]]+)\]\s+\[([^\]]+)\]\s+(.*)$`),
			handler: func(matches []string, dateContext string) *LogEntry {
				timestamp := m.parseTimestamp(matches[2], dateContext)
				return &LogEntry{
					Timestamp: timestamp,
					Level:     matches[3],
					Component: matches[4],
					Message:   matches[5],
				}
			},
		},
		{
			name:  "simple_log",
			regex: regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+(.*)$`),
			handler: func(matches []string, dateContext string) *LogEntry {
				timestamp := m.parseTimestamp(matches[1], dateContext)
				return &LogEntry{
					Timestamp: timestamp,
					Level:     "INFO",
					Component: "SYSTEM",
					Message:   matches[2],
				}
			},
		},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(line)
		if len(matches) > 0 {
			return pattern.handler(matches, dateContext)
		}
	}

	// Fallback: treat entire line as message
	return &LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "UNKNOWN",
		Message:   line,
	}
}

// parseTimestamp parses timestamp from log format
func (m *MigrationService) parseTimestamp(timestampStr, dateContext string) time.Time {
	// Try to parse the bracketed timestamp format [2025-07-18 08:41:03]
	if t, err := time.Parse("2006-01-02 15:04:05", timestampStr); err == nil {
		return t
	}

	// Try to parse the simple timestamp format 2025/07/18 08:41:03
	if t, err := time.Parse("2006/01/02 15:04:05", timestampStr); err == nil {
		return t
	}

	// Fallback to current time
	return time.Now()
}

// CreateBackupDatabase creates a backup of the existing JSON files before migration
func (m *MigrationService) CreateBackupDatabase() error {
	backupDir := filepath.Join(m.dataDir, "json_backup_"+time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Files to backup
	filesToBackup := []string{
		"config.json",
		filepath.Join(m.dataDir, "backups.json"),
	}

	for _, file := range filesToBackup {
		if _, err := os.Stat(file); err == nil {
			// Copy file to backup directory
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			backupFile := filepath.Join(backupDir, filepath.Base(file))
			os.WriteFile(backupFile, data, 0644)
		}
	}

	fmt.Printf("📂 JSON backup created at: %s\n", backupDir)
	return nil
}

// GetMigrationStatus checks what data needs to be migrated
func (m *MigrationService) GetMigrationStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Check PostgreSQL instances
	postgresRepo := NewPostgreSQLRepository(m.db)
	instances, err := postgresRepo.GetAll()
	if err != nil {
		return nil, err
	}
	status["postgresql_instances"] = len(instances)

	// Check backups
	backupRepo := NewBackupRepository(m.db)
	backups, err := backupRepo.GetAll()
	if err != nil {
		return nil, err
	}
	status["backups"] = len(backups)

	// Check logs
	logRepo := NewLogRepository(m.db)
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		return nil, err
	}
	status["has_logs"] = len(logs) > 0

	// Check if JSON files exist
	status["json_files_exist"] = map[string]bool{
		"config.json":  m.fileExists("config.json") || m.fileExists(filepath.Join(m.dataDir, "../config.json")),
		"backups.json": m.fileExists(filepath.Join(m.dataDir, "backups.json")),
		"logs":         m.dirExists("logs"),
	}

	return status, nil
}

// Helper functions
func (m *MigrationService) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *MigrationService) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
