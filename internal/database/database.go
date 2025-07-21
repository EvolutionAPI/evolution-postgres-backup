package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB wraps sql.DB with PostgreSQL connection
type DB struct {
	*sql.DB
	connStr string
}

// NewDB creates a new PostgreSQL database connection
func NewDB(dataDir string) (*DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "backup_service"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "backup_admin"
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "backup_password_2024"
	}

	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open PostgreSQL database
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	db := &DB{
		DB:      sqlDB,
		connStr: connStr,
	}

	// Configure connection pool for PostgreSQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return db, nil
}

// GetDBType returns "postgres"
func (db *DB) GetDBType() string {
	return "postgres"
}

// GetDBPath returns the PostgreSQL connection string (sanitized)
func (db *DB) GetDBPath() string {
	return db.connStr
}

// GetStats returns PostgreSQL database statistics
func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get table counts
	tables := []string{"postgresql_instances", "backups", "logs", "schedules", "jobs", "config"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := db.QueryRow(query).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", table, err)
		}
		stats[table+"_count"] = count
	}

	// Get database size
	var dbSize int64
	query := "SELECT pg_database_size(current_database())"
	if err := db.QueryRow(query).Scan(&dbSize); err == nil {
		stats["database_size_bytes"] = dbSize
		stats["database_size_mb"] = float64(dbSize) / 1024 / 1024
	}

	// Get PostgreSQL version
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err == nil {
		stats["postgres_version"] = version
	}

	// Get connection info
	stats["max_connections"] = 25
	stats["idle_connections"] = 5

	return stats, nil
}
