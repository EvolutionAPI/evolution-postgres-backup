package models

import (
	"time"
)

type BackupType string

const (
	BackupTypeHourly  BackupType = "hourly"
	BackupTypeDaily   BackupType = "daily"
	BackupTypeWeekly  BackupType = "weekly"
	BackupTypeMonthly BackupType = "monthly"
	BackupTypeManual  BackupType = "manual"
)

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

type BackupInfo struct {
	ID           string       `json:"id"`
	PostgreSQLID string       `json:"postgresql_id"`
	DatabaseName string       `json:"database_name"`
	BackupType   BackupType   `json:"backup_type"`
	Status       BackupStatus `json:"status"`
	StartTime    time.Time    `json:"start_time"`
	EndTime      *time.Time   `json:"end_time,omitempty"`
	FilePath     string       `json:"file_path"`
	FileSize     int64        `json:"file_size"`
	JobID        string       `json:"job_id,omitempty"` // Associated job ID for log correlation
	S3Key        string       `json:"s3_key"`
	ErrorMessage string       `json:"error_message,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type RestoreRequest struct {
	BackupID     string `json:"backup_id" binding:"required"`
	PostgreSQLID string `json:"postgresql_id" binding:"required"`
	DatabaseName string `json:"database_name" binding:"required"`
}

type BackupRequest struct {
	PostgreSQLID string     `json:"postgresql_id" binding:"required"`
	BackupType   BackupType `json:"backup_type"`
	DatabaseName string     `json:"database_name,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
