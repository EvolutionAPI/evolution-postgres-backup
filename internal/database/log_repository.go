package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// LogEntry represents a log entry in the system
type LogEntry struct {
	ID        int64     `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Level     string    `json:"level" db:"level"`         // DEBUG, INFO, WARN, ERROR
	Component string    `json:"component" db:"component"` // api, worker, backup, etc.
	JobID     string    `json:"job_id,omitempty" db:"job_id"`
	BackupID  string    `json:"backup_id,omitempty" db:"backup_id"`
	Message   string    `json:"message" db:"message"`
	Details   string    `json:"details,omitempty" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LogRepository struct {
	db *DB
}

func NewLogRepository(db *DB) *LogRepository {
	return &LogRepository{db: db}
}

// Create inserts a new log entry
func (r *LogRepository) Create(entry *LogEntry) error {
	query := `
		INSERT INTO logs (
			timestamp, level, component, job_id, backup_id, 
			message, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	result, err := r.db.Exec(
		query,
		entry.Timestamp,
		entry.Level,
		entry.Component,
		nullString(entry.JobID),
		nullString(entry.BackupID),
		entry.Message,
		nullString(entry.Details),
		time.Now(),
	)

	if err != nil {
		return err
	}

	// Get the inserted ID (PostgreSQL way)
	if entry.ID == 0 {
		// For PostgreSQL, we need to use RETURNING clause or get the last inserted id differently
		// Since the table uses SERIAL, the ID is auto-generated
	}

	_ = result // Suppress unused variable warning
	return nil
}

// CreateBatch inserts multiple log entries efficiently
func (r *LogRepository) CreateBatch(entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	query := `
		INSERT INTO logs (
			timestamp, level, component, job_id, backup_id, 
			message, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.Exec(
			entry.Timestamp,
			entry.Level,
			entry.Component,
			nullString(entry.JobID),
			nullString(entry.BackupID),
			entry.Message,
			nullString(entry.Details),
			time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFiltered retrieves logs with filtering options
func (r *LogRepository) GetFiltered(filters LogFilters) ([]*LogEntry, error) {
	query := `
		SELECT id, timestamp, level, component, job_id, backup_id, message, details, created_at
		FROM logs`

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Apply date filter
	if !filters.StartDate.IsZero() && !filters.EndDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp BETWEEN $%d AND $%d", argIndex, argIndex+1))
		args = append(args, filters.StartDate, filters.EndDate)
		argIndex += 2
	} else if !filters.StartDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, filters.StartDate)
		argIndex++
	} else if !filters.EndDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, filters.EndDate)
		argIndex++
	}

	// Apply level filter
	if filters.Level != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("level = $%d", argIndex))
		args = append(args, filters.Level)
		argIndex++
	}

	// Apply component filter
	if filters.Component != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("component = $%d", argIndex))
		args = append(args, filters.Component)
		argIndex++
	}

	// Apply job ID filter
	if filters.JobID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("job_id = $%d", argIndex))
		args = append(args, filters.JobID)
		argIndex++
	}

	// Apply backup ID filter
	if filters.BackupID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("backup_id = $%d", argIndex))
		args = append(args, filters.BackupID)
		argIndex++
	}

	// Add WHERE clause if we have filters
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add ordering and limit
	query += " ORDER BY timestamp DESC"
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogEntry
	for rows.Next() {
		log, err := r.scanLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetByJobID retrieves all logs for a specific job
func (r *LogRepository) GetByJobID(jobID string) ([]*LogEntry, error) {
	return r.GetFiltered(LogFilters{
		JobID: jobID,
	})
}

// GetByBackupID retrieves all logs for a specific backup
func (r *LogRepository) GetByBackupID(backupID string) ([]*LogEntry, error) {
	return r.GetFiltered(LogFilters{
		BackupID: backupID,
	})
}

// GetRecent retrieves the most recent logs
func (r *LogRepository) GetRecent(limit int) ([]*LogEntry, error) {
	return r.GetFiltered(LogFilters{
		Limit: limit,
	})
}

// GetByDate retrieves logs for a specific date
func (r *LogRepository) GetByDate(date time.Time) ([]*LogEntry, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.Add(24 * time.Hour)

	return r.GetFiltered(LogFilters{
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// DeleteOldLogs removes logs older than the specified duration
func (r *LogRepository) DeleteOldLogs(olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	query := "DELETE FROM logs WHERE timestamp < $1"

	result, err := r.db.Exec(query, cutoffTime)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns log statistics
func (r *LogRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total logs
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&total); err != nil {
		return nil, err
	}
	stats["total_logs"] = total

	// Logs by level
	levelQuery := `
		SELECT level, COUNT(*) 
		FROM logs 
		GROUP BY level`

	rows, err := r.db.Query(levelQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	levelCounts := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		levelCounts[level] = count
	}
	stats["by_level"] = levelCounts

	// Logs by component
	componentQuery := `
		SELECT component, COUNT(*) 
		FROM logs 
		GROUP BY component`

	rows, err = r.db.Query(componentQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	componentCounts := make(map[string]int)
	for rows.Next() {
		var component string
		var count int
		if err := rows.Scan(&component, &count); err != nil {
			return nil, err
		}
		componentCounts[component] = count
	}
	stats["by_component"] = componentCounts

	return stats, nil
}

// scanLog scans a database row into a LogEntry struct
func (r *LogRepository) scanLog(scanner interface {
	Scan(dest ...interface{}) error
}) (*LogEntry, error) {
	entry := &LogEntry{}
	var jobID, backupID, details sql.NullString

	err := scanner.Scan(
		&entry.ID,
		&entry.Timestamp,
		&entry.Level,
		&entry.Component,
		&jobID,
		&backupID,
		&entry.Message,
		&details,
		&entry.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if jobID.Valid {
		entry.JobID = jobID.String
	}
	if backupID.Valid {
		entry.BackupID = backupID.String
	}
	if details.Valid {
		entry.Details = details.String
	}

	return entry, nil
}

// LogFilters represents filters for querying logs
type LogFilters struct {
	StartDate time.Time
	EndDate   time.Time
	Level     string
	Component string
	JobID     string
	BackupID  string
	Limit     int
}

// Helper function to convert string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
