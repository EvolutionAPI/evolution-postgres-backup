package service

import (
	"encoding/json"
	"evolution-postgres-backup/internal/models"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type BackupPersistence struct {
	dataFile string
	mutex    sync.RWMutex
}

type PersistedBackup struct {
	*models.BackupInfo
	// Add serialization helpers for time fields if needed
}

func NewBackupPersistence(dataDir string) *BackupPersistence {
	// Ensure data directory exists
	os.MkdirAll(dataDir, 0755)

	return &BackupPersistence{
		dataFile: filepath.Join(dataDir, "backups.json"),
	}
}

// SaveBackups saves all backups to JSON file
func (bp *BackupPersistence) SaveBackups(backups map[string]*models.BackupInfo) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	// Convert map to slice for JSON serialization
	backupList := make([]*models.BackupInfo, 0, len(backups))
	for _, backup := range backups {
		backupList = append(backupList, backup)
	}

	// Create backup of existing file
	if _, err := os.Stat(bp.dataFile); err == nil {
		backupFile := bp.dataFile + ".backup." + time.Now().Format("20060102-150405")
		if copyErr := copyFile(bp.dataFile, backupFile); copyErr != nil {
			// Log error but don't fail the save operation
			os.Stderr.WriteString("Warning: failed to create backup of data file: " + copyErr.Error() + "\n")
		}
	}

	// Marshal to JSON with pretty formatting
	data, err := json.MarshalIndent(backupList, "", "  ")
	if err != nil {
		return err
	}

	// Write to temporary file first, then rename (atomic operation)
	tempFile := bp.dataFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tempFile, bp.dataFile)
}

// LoadBackups loads all backups from JSON file
func (bp *BackupPersistence) LoadBackups() (map[string]*models.BackupInfo, error) {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	backups := make(map[string]*models.BackupInfo)

	// Check if file exists
	if _, err := os.Stat(bp.dataFile); os.IsNotExist(err) {
		// File doesn't exist, return empty map (first run)
		return backups, nil
	}

	// Read file
	data, err := os.ReadFile(bp.dataFile)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var backupList []*models.BackupInfo
	if err := json.Unmarshal(data, &backupList); err != nil {
		return nil, err
	}

	// Convert slice back to map
	for _, backup := range backupList {
		backups[backup.ID] = backup
	}

	return backups, nil
}

// SaveSingleBackup saves a single backup (for real-time updates)
func (bp *BackupPersistence) SaveSingleBackup(allBackups map[string]*models.BackupInfo) error {
	// For simplicity, save all backups (could be optimized later)
	return bp.SaveBackups(allBackups)
}

// CleanupOldBackupFiles removes old backup files (keep last 5)
func (bp *BackupPersistence) CleanupOldBackupFiles() {
	pattern := bp.dataFile + ".backup.*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// Sort by modification time and keep only the 5 most recent
	if len(matches) > 5 {
		// Simple cleanup - could be more sophisticated
		for i := 0; i < len(matches)-5; i++ {
			os.Remove(matches[i])
		}
	}
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
