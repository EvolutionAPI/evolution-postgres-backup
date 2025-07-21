package api

import (
	"evolution-postgres-backup/internal/config"
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/models"
	"evolution-postgres-backup/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// V2Handlers provides modern API handlers using SQLite
type V2Handlers struct {
	dbService *service.DatabaseService
}

// NewV2Handlers creates new V2 API handlers
func NewV2Handlers(dbService *service.DatabaseService) *V2Handlers {
	return &V2Handlers{
		dbService: dbService,
	}
}

// ==================== Dashboard & Stats ====================

// GetDashboard returns comprehensive dashboard data
func (h *V2Handlers) GetDashboard(c *gin.Context) {
	stats, err := h.dbService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get dashboard stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Dashboard data retrieved successfully",
		Data:    stats,
	})
}

// GetBackupTrends returns backup trends and analytics
func (h *V2Handlers) GetBackupTrends(c *gin.Context) {
	days := 30 // Default to 30 days
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	trends, err := h.dbService.GetBackupTrends(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get backup trends: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Backup trends retrieved successfully",
		Data:    trends,
	})
}

// ==================== Advanced Backup Management ====================

// GetBackupsAdvanced returns backups with advanced filtering
func (h *V2Handlers) GetBackupsAdvanced(c *gin.Context) {
	var filters []database.BackupFilter

	// Parse filters from query parameters
	if postgresID := c.Query("postgres_id"); postgresID != "" {
		filters = append(filters, database.FilterByPostgreSQLID(postgresID))
	}

	if status := c.Query("status"); status != "" {
		filters = append(filters, database.FilterByStatus(models.BackupStatus(status)))
	}

	if backupType := c.Query("type"); backupType != "" {
		filters = append(filters, database.FilterByType(models.BackupType(backupType)))
	}

	backups, err := h.dbService.GetBackups(filters...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get backups: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Backups retrieved successfully",
		Data:    backups,
	})
}

// GetBackupsByInstance returns all backups for a specific PostgreSQL instance
func (h *V2Handlers) GetBackupsByInstance(c *gin.Context) {
	postgresID := c.Param("id")
	if postgresID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "PostgreSQL instance ID is required",
		})
		return
	}

	backups, err := h.dbService.GetBackupsByInstance(postgresID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get backups: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Instance backups retrieved successfully",
		Data:    backups,
	})
}

// ==================== Advanced Log Management ====================

// GetLogsAdvanced returns logs with advanced filtering
func (h *V2Handlers) GetLogsAdvanced(c *gin.Context) {
	filters := database.LogFilters{}

	// Parse date filters
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = t.Add(24 * time.Hour) // End of day
		}
	}

	// Parse other filters
	filters.Level = c.Query("level")
	filters.Component = c.Query("component")
	filters.JobID = c.Query("job_id")
	filters.BackupID = c.Query("backup_id")

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100 // Default limit
	}

	logs, err := h.dbService.GetLogs(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Logs retrieved successfully",
		Data:    logs,
	})
}

// GetLogsByJobID returns all logs for a specific job
func (h *V2Handlers) GetLogsByJobID(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Job ID is required",
		})
		return
	}

	logs, err := h.dbService.GetLogsByJobID(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get job logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Job logs retrieved successfully",
		Data:    logs,
	})
}

// GetLogsByBackupID returns all logs for a specific backup
func (h *V2Handlers) GetLogsByBackupID(c *gin.Context) {
	backupID := c.Param("backup_id")
	if backupID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Backup ID is required",
		})
		return
	}

	logs, err := h.dbService.GetLogsByBackupID(backupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get backup logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Backup logs retrieved successfully",
		Data:    logs,
	})
}

// ==================== PostgreSQL Instance Management ====================

// GetPostgreSQLInstances returns PostgreSQL instances with filtering
func (h *V2Handlers) GetPostgreSQLInstances(c *gin.Context) {
	var instances []*config.PostgreSQLConfig
	var err error

	if c.Query("enabled") == "true" {
		instances, err = h.dbService.GetEnabledPostgreSQLInstances()
	} else {
		instances, err = h.dbService.GetPostgreSQLInstances()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get PostgreSQL instances: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "PostgreSQL instances retrieved successfully",
		Data:    instances,
	})
}

// GetPostgreSQLInstance returns a specific PostgreSQL instance
func (h *V2Handlers) GetPostgreSQLInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Instance ID is required",
		})
		return
	}

	instance, err := h.dbService.GetPostgreSQLInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "PostgreSQL instance not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "PostgreSQL instance retrieved successfully",
		Data:    instance,
	})
}

// CreatePostgreSQLInstance creates a new PostgreSQL instance
func (h *V2Handlers) CreatePostgreSQLInstance(c *gin.Context) {
	var instance config.PostgreSQLConfig
	if err := c.ShouldBindJSON(&instance); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid JSON format: " + err.Error(),
		})
		return
	}

	if err := h.dbService.CreatePostgreSQLInstance(&instance); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create PostgreSQL instance: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "PostgreSQL instance created successfully",
		Data:    instance,
	})
}

// UpdatePostgreSQLInstance updates a PostgreSQL instance
func (h *V2Handlers) UpdatePostgreSQLInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Instance ID is required",
		})
		return
	}

	var instance config.PostgreSQLConfig
	if err := c.ShouldBindJSON(&instance); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid JSON format: " + err.Error(),
		})
		return
	}

	instance.ID = id // Ensure ID matches URL parameter

	if err := h.dbService.UpdatePostgreSQLInstance(&instance); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to update PostgreSQL instance: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "PostgreSQL instance updated successfully",
		Data:    instance,
	})
}

// DeletePostgreSQLInstance deletes a PostgreSQL instance
func (h *V2Handlers) DeletePostgreSQLInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Instance ID is required",
		})
		return
	}

	if err := h.dbService.DeletePostgreSQLInstance(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to delete PostgreSQL instance: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "PostgreSQL instance deleted successfully",
	})
}

// ==================== Health Checks ====================

// GetHealthDetailed returns detailed health information
func (h *V2Handlers) GetHealthDetailed(c *gin.Context) {
	health := h.dbService.HealthCheck()

	// Determine overall status
	overallStatus := "healthy"
	for _, component := range health {
		if comp, ok := component.(map[string]interface{}); ok {
			if status, exists := comp["status"]; exists && status != "healthy" {
				overallStatus = "unhealthy"
				break
			}
		}
	}

	response := map[string]interface{}{
		"status":     overallStatus,
		"timestamp":  time.Now(),
		"components": health,
	}

	if overallStatus == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// ==================== Migration Management ====================

// GetMigrationStatus returns migration status
func (h *V2Handlers) GetMigrationStatus(c *gin.Context) {
	status, err := h.dbService.GetMigrationStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get migration status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Migration status retrieved successfully",
		Data:    status,
	})
}

// PerformMigration performs migration from JSON to SQLite
func (h *V2Handlers) PerformMigration(c *gin.Context) {
	if err := h.dbService.PerformMigration(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Migration failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Migration completed successfully",
	})
}

// ==================== Utility Functions ====================

// parseTimeFromQuery parses time from query parameter
func parseTimeFromQuery(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

// parseStringSlice parses comma-separated string into slice
func parseStringSlice(str string) []string {
	if str == "" {
		return nil
	}

	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
