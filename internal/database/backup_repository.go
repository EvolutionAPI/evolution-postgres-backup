package database

import (
	"database/sql"
	"evolution-postgres-backup/internal/models"
	"time"
)

type BackupRepository struct {
	db *DB
}

func NewBackupRepository(db *DB) *BackupRepository {
	return &BackupRepository{db: db}
}

// Create inserts a new backup record
func (r *BackupRepository) Create(backup *models.BackupInfo) error {
	query := `
		INSERT INTO backups (
			id, postgresql_id, database_name, backup_type, status,
			start_time, end_time, file_path, file_size, s3_key,
			error_message, created_at, job_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(
		query,
		backup.ID,
		backup.PostgreSQLID,
		backup.DatabaseName,
		string(backup.BackupType),
		string(backup.Status),
		backup.StartTime,
		backup.EndTime,
		backup.FilePath,
		backup.FileSize,
		backup.S3Key,
		backup.ErrorMessage,
		backup.CreatedAt,
		backup.JobID,
	)

	return err
}

// Update updates an existing backup record
func (r *BackupRepository) Update(backup *models.BackupInfo) error {
	query := `
		UPDATE backups SET
			status = $1,
			end_time = $2,
			file_path = $3,
			file_size = $4,
			s3_key = $5,
			error_message = $6
		WHERE id = $7`

	_, err := r.db.Exec(
		query,
		string(backup.Status),
		backup.EndTime,
		backup.FilePath,
		backup.FileSize,
		backup.S3Key,
		backup.ErrorMessage,
		backup.ID,
	)

	return err
}

// GetByID retrieves a backup by ID
func (r *BackupRepository) GetByID(id string) (*models.BackupInfo, error) {
	query := `
		SELECT id, postgresql_id, database_name, backup_type, status,
			   start_time, end_time, file_path, file_size, s3_key,
			   error_message, created_at
		FROM backups WHERE id = $1`

	row := r.db.QueryRow(query, id)
	return r.scanBackup(row)
}

// GetAll retrieves all backups with optional filters
func (r *BackupRepository) GetAll(filters ...BackupFilter) ([]*models.BackupInfo, error) {
	query := `
		SELECT id, postgresql_id, database_name, backup_type, status,
			   start_time, end_time, file_path, file_size, s3_key,
			   error_message, created_at
		FROM backups`

	args := []interface{}{}
	whereClauses := []string{}

	// Apply filters
	for _, filter := range filters {
		clause, arg := filter.Apply()
		if clause != "" {
			whereClauses = append(whereClauses, clause)
			if arg != nil {
				args = append(args, arg)
			}
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*models.BackupInfo
	for rows.Next() {
		backup, err := r.scanBackup(rows)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}

	return backups, rows.Err()
}

// GetByPostgreSQLID retrieves backups for a specific PostgreSQL instance
func (r *BackupRepository) GetByPostgreSQLID(postgresID string) ([]*models.BackupInfo, error) {
	return r.GetAll(FilterByPostgreSQLID(postgresID))
}

// GetByStatus retrieves backups with a specific status
func (r *BackupRepository) GetByStatus(status models.BackupStatus) ([]*models.BackupInfo, error) {
	return r.GetAll(FilterByStatus(status))
}

// GetByType retrieves backups of a specific type
func (r *BackupRepository) GetByType(backupType models.BackupType) ([]*models.BackupInfo, error) {
	return r.GetAll(FilterByType(backupType))
}

// GetOldBackups retrieves backups older than the specified time for cleanup
func (r *BackupRepository) GetOldBackups(postgresID string, backupType models.BackupType, olderThan time.Time) ([]*models.BackupInfo, error) {
	query := `
		SELECT id, postgresql_id, database_name, backup_type, status,
			   start_time, end_time, file_path, file_size, s3_key,
			   error_message, created_at
		FROM backups 
		WHERE postgresql_id = $1 AND backup_type = $2 AND created_at < $3
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, postgresID, string(backupType), olderThan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*models.BackupInfo
	for rows.Next() {
		backup, err := r.scanBackup(rows)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}

	return backups, rows.Err()
}

// Delete removes a backup record
func (r *BackupRepository) Delete(id string) error {
	query := "DELETE FROM backups WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteOldBackups removes backups older than the specified time
func (r *BackupRepository) DeleteOldBackups(postgresID string, backupType models.BackupType, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM backups 
		WHERE postgresql_id = $1 AND backup_type = $2 AND created_at < $3`

	result, err := r.db.Exec(query, postgresID, string(backupType), olderThan)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns backup statistics
func (r *BackupRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total backups by status
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM backups 
		GROUP BY status`

	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}
	stats["by_status"] = statusCounts

	// Total backups by type
	typeQuery := `
		SELECT backup_type, COUNT(*) 
		FROM backups 
		GROUP BY backup_type`

	rows, err = r.db.Query(typeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	typeCounts := make(map[string]int)
	for rows.Next() {
		var backupType string
		var count int
		if err := rows.Scan(&backupType, &count); err != nil {
			return nil, err
		}
		typeCounts[backupType] = count
	}
	stats["by_type"] = typeCounts

	// Total storage used
	var totalSize sql.NullInt64
	sizeQuery := "SELECT SUM(file_size) FROM backups WHERE status = 'completed'"
	if err := r.db.QueryRow(sizeQuery).Scan(&totalSize); err != nil {
		return nil, err
	}
	if totalSize.Valid {
		stats["total_size_bytes"] = totalSize.Int64
		stats["total_size_mb"] = float64(totalSize.Int64) / 1024 / 1024
	}

	return stats, nil
}

// scanBackup scans a database row into a BackupInfo struct
func (r *BackupRepository) scanBackup(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.BackupInfo, error) {
	backup := &models.BackupInfo{}
	var backupType, status string
	var endTime sql.NullTime

	err := scanner.Scan(
		&backup.ID,
		&backup.PostgreSQLID,
		&backup.DatabaseName,
		&backupType,
		&status,
		&backup.StartTime,
		&endTime,
		&backup.FilePath,
		&backup.FileSize,
		&backup.S3Key,
		&backup.ErrorMessage,
		&backup.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	backup.BackupType = models.BackupType(backupType)
	backup.Status = models.BackupStatus(status)

	if endTime.Valid {
		backup.EndTime = &endTime.Time
	}

	return backup, nil
}

// BackupFilter interface for filtering backups
type BackupFilter interface {
	Apply() (string, interface{})
}

// Concrete filter implementations
type backupPostgreSQLIDFilter struct {
	postgresID string
}

func FilterByPostgreSQLID(postgresID string) BackupFilter {
	return &backupPostgreSQLIDFilter{postgresID: postgresID}
}

func (f *backupPostgreSQLIDFilter) Apply() (string, interface{}) {
	return "postgresql_id = $1", f.postgresID
}

type backupStatusFilter struct {
	status models.BackupStatus
}

func FilterByStatus(status models.BackupStatus) BackupFilter {
	return &backupStatusFilter{status: status}
}

func (f *backupStatusFilter) Apply() (string, interface{}) {
	return "status = $1", string(f.status)
}

type backupTypeFilter struct {
	backupType models.BackupType
}

func FilterByType(backupType models.BackupType) BackupFilter {
	return &backupTypeFilter{backupType: backupType}
}

func (f *backupTypeFilter) Apply() (string, interface{}) {
	return "backup_type = $1", string(f.backupType)
}
