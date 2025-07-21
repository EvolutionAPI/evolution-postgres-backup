package database

import (
	"database/sql"
	"encoding/json"
	"evolution-postgres-backup/internal/config"
	"time"
)

type PostgreSQLRepository struct {
	db *DB
}

func NewPostgreSQLRepository(db *DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// Create inserts a new PostgreSQL instance
func (r *PostgreSQLRepository) Create(instance *config.PostgreSQLConfig) error {
	// Convert databases array to JSON
	databasesJSON, err := json.Marshal(instance.GetDatabases())
	if err != nil {
		return err
	}

	query := `
		INSERT INTO postgresql_instances (
			id, name, host, port, username, password, databases, enabled, ssl_mode, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now()
	_, err = r.db.Exec(
		query,
		instance.ID,
		instance.Name,
		instance.Host,
		instance.Port,
		instance.Username,
		instance.Password,
		string(databasesJSON),
		instance.Enabled,
		instance.GetSSLMode(),
		now,
		now,
	)

	return err
}

// Update updates an existing PostgreSQL instance
func (r *PostgreSQLRepository) Update(instance *config.PostgreSQLConfig) error {
	// Convert databases array to JSON
	databasesJSON, err := json.Marshal(instance.GetDatabases())
	if err != nil {
		return err
	}

	query := `
		UPDATE postgresql_instances SET
			name = $1,
			host = $2,
			port = $3,
			username = $4,
			password = $5,
			databases = $6,
			enabled = $7,
			ssl_mode = $8,
			updated_at = $9
		WHERE id = $10`

	_, err = r.db.Exec(
		query,
		instance.Name,
		instance.Host,
		instance.Port,
		instance.Username,
		instance.Password,
		string(databasesJSON),
		instance.Enabled,
		instance.GetSSLMode(),
		time.Now(),
		instance.ID,
	)

	return err
}

// GetByID retrieves a PostgreSQL instance by ID
func (r *PostgreSQLRepository) GetByID(id string) (*config.PostgreSQLConfig, error) {
	query := `
		SELECT id, name, host, port, username, password, databases, enabled, ssl_mode, created_at, updated_at
		FROM postgresql_instances WHERE id = $1`

	row := r.db.QueryRow(query, id)
	return r.scanPostgreSQL(row)
}

// GetAll retrieves all PostgreSQL instances
func (r *PostgreSQLRepository) GetAll() ([]*config.PostgreSQLConfig, error) {
	query := `
		SELECT id, name, host, port, username, password, databases, enabled, ssl_mode, created_at, updated_at
		FROM postgresql_instances 
		ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*config.PostgreSQLConfig
	for rows.Next() {
		instance, err := r.scanPostgreSQL(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, rows.Err()
}

// GetEnabled retrieves all enabled PostgreSQL instances
func (r *PostgreSQLRepository) GetEnabled() ([]*config.PostgreSQLConfig, error) {
	query := `
		SELECT id, name, host, port, username, password, databases, enabled, ssl_mode, created_at, updated_at
		FROM postgresql_instances 
		WHERE enabled = true
		ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*config.PostgreSQLConfig
	for rows.Next() {
		instance, err := r.scanPostgreSQL(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, rows.Err()
}

// Delete removes a PostgreSQL instance
func (r *PostgreSQLRepository) Delete(id string) error {
	query := `DELETE FROM postgresql_instances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Exists checks if a PostgreSQL instance exists
func (r *PostgreSQLRepository) Exists(id string) (bool, error) {
	query := "SELECT 1 FROM postgresql_instances WHERE id = $1 LIMIT 1"
	var exists int
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetStats returns PostgreSQL instances statistics
func (r *PostgreSQLRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total instances
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM postgresql_instances").Scan(&total); err != nil {
		return nil, err
	}
	stats["total_instances"] = total

	// Enabled instances
	var enabled int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM postgresql_instances WHERE enabled = true").Scan(&enabled); err != nil {
		return nil, err
	}
	stats["enabled_instances"] = enabled

	// Disabled instances
	stats["disabled_instances"] = total - enabled

	return stats, nil
}

// scanPostgreSQL scans a row into a PostgreSQLConfig
func (r *PostgreSQLRepository) scanPostgreSQL(scanner interface {
	Scan(dest ...interface{}) error
}) (*config.PostgreSQLConfig, error) {
	var instance config.PostgreSQLConfig
	var databasesJSON string
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&instance.ID,
		&instance.Name,
		&instance.Host,
		&instance.Port,
		&instance.Username,
		&instance.Password,
		&databasesJSON,
		&instance.Enabled,
		&instance.SSLMode,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse databases JSON
	if databasesJSON != "" {
		var databases []string
		if err := json.Unmarshal([]byte(databasesJSON), &databases); err == nil {
			instance.Databases = databases
		}
	}

	// Set default if no databases
	if len(instance.Databases) == 0 {
		instance.Databases = []string{"postgres"}
	}

	return &instance, nil
}
