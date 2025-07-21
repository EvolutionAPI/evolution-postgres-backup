package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type PostgreSQLConfig struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	Database  string   `json:"database,omitempty"`  // Backwards compatibility
	Databases []string `json:"databases,omitempty"` // New: multiple databases
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Enabled   bool     `json:"enabled"`
	SSLMode   string   `json:"ssl_mode,omitempty"` // For PostgreSQL connections: disable, allow, prefer, require
}

// GetSSLMode returns the SSL mode for PostgreSQL connection, with default fallback
func (pg *PostgreSQLConfig) GetSSLMode() string {
	if pg.SSLMode != "" {
		return pg.SSLMode
	}
	return "prefer" // Default SSL mode
}

// GetDatabases returns all databases for this PostgreSQL instance
func (pg *PostgreSQLConfig) GetDatabases() []string {
	// If databases array is specified, use it
	if len(pg.Databases) > 0 {
		return pg.Databases
	}

	// Check if Database field contains comma-separated values
	if pg.Database != "" {
		// Split by comma and trim spaces
		databases := strings.Split(pg.Database, ",")
		var cleanDatabases []string
		for _, db := range databases {
			db = strings.TrimSpace(db)
			if db != "" {
				cleanDatabases = append(cleanDatabases, db)
			}
		}
		if len(cleanDatabases) > 0 {
			return cleanDatabases
		}
	}

	// Default to 'postgres' if nothing specified
	return []string{"postgres"}
}

// GetDefaultDatabase returns the first database (for API compatibility)
func (pg *PostgreSQLConfig) GetDefaultDatabase() string {
	databases := pg.GetDatabases()
	if len(databases) > 0 {
		return databases[0]
	}
	return "postgres"
}

type RetentionPolicy struct {
	Hourly  int `json:"hourly"`  // hours
	Daily   int `json:"daily"`   // days
	Weekly  int `json:"weekly"`  // weeks
	Monthly int `json:"monthly"` // months
}

type Config struct {
	PostgreSQLInstances []PostgreSQLConfig `json:"postgresql_instances"`
	RetentionPolicy     RetentionPolicy    `json:"retention_policy"`
	S3Config            S3Config           `json:"s3_config"`
}

type S3Config struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	UseSSL          bool   `json:"use_ssl"`
}

func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override S3 configuration from environment variables if available
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		config.S3Config.Endpoint = endpoint
	}
	if region := os.Getenv("S3_REGION"); region != "" {
		config.S3Config.Region = region
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		config.S3Config.Bucket = bucket
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY_ID"); accessKey != "" {
		config.S3Config.AccessKeyID = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_ACCESS_KEY"); secretKey != "" {
		config.S3Config.SecretAccessKey = secretKey
	}
	if useSSL := os.Getenv("S3_USE_SSL"); useSSL != "" {
		config.S3Config.UseSSL = useSSL == "true"
	}

	// Validate required S3 configuration
	if config.S3Config.Region == "" {
		return nil, fmt.Errorf("S3_REGION environment variable is required")
	}
	if config.S3Config.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}
	if config.S3Config.AccessKeyID == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY_ID environment variable is required")
	}
	if config.S3Config.SecretAccessKey == "" {
		return nil, fmt.Errorf("S3_SECRET_ACCESS_KEY environment variable is required")
	}

	return &config, nil
}

func (c *Config) Save(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func (c *Config) AddPostgreSQL(pg PostgreSQLConfig) {
	c.PostgreSQLInstances = append(c.PostgreSQLInstances, pg)
}

func (c *Config) GetPostgreSQLByID(id string) (*PostgreSQLConfig, bool) {
	for i, pg := range c.PostgreSQLInstances {
		if pg.ID == id {
			return &c.PostgreSQLInstances[i], true
		}
	}
	return nil, false
}

func (c *Config) RemovePostgreSQLByID(id string) bool {
	for i, pg := range c.PostgreSQLInstances {
		if pg.ID == id {
			c.PostgreSQLInstances = append(c.PostgreSQLInstances[:i], c.PostgreSQLInstances[i+1:]...)
			return true
		}
	}
	return false
}
