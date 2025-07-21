package api

import (
	"evolution-postgres-backup/internal/models"
	"evolution-postgres-backup/internal/worker"
	"fmt"
	"log"
	"net/http"
	"time"

	"evolution-postgres-backup/internal/database"

	"github.com/gin-gonic/gin"
)

// WorkerHandlers provides API handlers for worker management
type WorkerHandlers struct {
	jobQueue *worker.JobQueue
}

// NewWorkerHandlers creates new worker API handlers
func NewWorkerHandlers(jobQueue *worker.JobQueue) *WorkerHandlers {
	return &WorkerHandlers{
		jobQueue: jobQueue,
	}
}

// ==================== Job Management ====================

// CreateBackupJob creates a new backup job
func (h *WorkerHandlers) CreateBackupJob(c *gin.Context) {
	var req struct {
		PostgresID   string            `json:"postgresql_id" binding:"required"`
		DatabaseName string            `json:"database_name" binding:"required"`
		BackupType   models.BackupType `json:"backup_type" binding:"required"`
		Priority     int               `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Default priority if not specified
	if req.Priority == 0 {
		req.Priority = 5 // Medium priority
	}

	// Create backup record first
	backup := &models.BackupInfo{
		ID:           fmt.Sprintf("backup_%d", time.Now().UnixNano()),
		PostgreSQLID: req.PostgresID,
		DatabaseName: req.DatabaseName,
		BackupType:   req.BackupType,
		Status:       models.BackupStatusPending,
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
	}

	// Save backup record to database
	backupRepo := database.NewBackupRepository(h.jobQueue.GetDB())
	if err := backupRepo.Create(backup); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create backup record: " + err.Error(),
		})
		return
	}

	// Create job with backup_id
	job := &worker.Job{
		Type:     worker.JobTypeBackup,
		Priority: req.Priority,
		Payload: map[string]interface{}{
			"postgres_id":   req.PostgresID,
			"database_name": req.DatabaseName,
			"backup_type":   string(req.BackupType),
			"backup_id":     backup.ID, // Include backup_id for worker
		},
		MaxRetries: 3,
	}

	// Add job to queue
	if err := h.jobQueue.AddJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to add job to queue: " + err.Error(),
		})
		return
	}

	// Associate backup with job and update
	backup.JobID = job.ID
	if err := backupRepo.Update(backup); err != nil {
		// Don't fail, just log the error
		log.Printf("Failed to update backup with job_id: %v", err)
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Backup job created successfully",
		Data:    backup, // Return backup instead of job
	})
}

// CreateRestoreJob creates a new restore job
func (h *WorkerHandlers) CreateRestoreJob(c *gin.Context) {
	var req struct {
		BackupID     string `json:"backup_id" binding:"required"`
		PostgresID   string `json:"postgresql_id" binding:"required"`
		DatabaseName string `json:"database_name" binding:"required"`
		Priority     int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Default priority if not specified
	if req.Priority == 0 {
		req.Priority = 8 // High priority for restores
	}

	job, err := h.jobQueue.AddRestoreJob(req.BackupID, req.PostgresID, req.DatabaseName, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create restore job: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Restore job created successfully",
		Data:    job,
	})
}

// CreateCleanupJob creates a new cleanup job
func (h *WorkerHandlers) CreateCleanupJob(c *gin.Context) {
	var req struct {
		PostgresID string            `json:"postgres_id" binding:"required"`
		BackupType models.BackupType `json:"backup_type" binding:"required"`
		Priority   int               `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Default priority if not specified
	if req.Priority == 0 {
		req.Priority = 3 // Low priority for cleanup
	}

	job, err := h.jobQueue.AddCleanupJob(req.PostgresID, req.BackupType, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create cleanup job: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Cleanup job created successfully",
		Data:    job,
	})
}

// ==================== Queue Monitoring ====================

// GetQueueStats returns queue statistics
func (h *WorkerHandlers) GetQueueStats(c *gin.Context) {
	stats := h.jobQueue.GetStats()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Queue statistics retrieved successfully",
		Data:    stats,
	})
}

// GetRunningJobs returns currently running jobs
func (h *WorkerHandlers) GetRunningJobs(c *gin.Context) {
	jobs := h.jobQueue.GetRunningJobs()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Running jobs retrieved successfully",
		Data:    jobs,
	})
}

// GetWorkerStatus returns status of all workers
func (h *WorkerHandlers) GetWorkerStatus(c *gin.Context) {
	workers := h.jobQueue.GetWorkerStatus()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Worker status retrieved successfully",
		Data:    workers,
	})
}

// GetQueueHealth returns queue health information
func (h *WorkerHandlers) GetQueueHealth(c *gin.Context) {
	stats := h.jobQueue.GetStats()
	workers := h.jobQueue.GetWorkerStatus()

	// Determine health status
	health := "healthy"
	issues := []string{}

	// Check if queue is running
	if !h.jobQueue.IsRunning() {
		health = "unhealthy"
		issues = append(issues, "queue is not running")
	}

	// Check for failed workers
	activeWorkers := 0
	for _, worker := range workers {
		if worker.Status == "working" || worker.Status == "idle" {
			activeWorkers++
		}
	}

	if activeWorkers == 0 {
		health = "unhealthy"
		issues = append(issues, "no active workers")
	}

	// Check for too many failed jobs
	if stats.FailedJobs > 10 {
		health = "degraded"
		issues = append(issues, "high number of failed jobs")
	}

	response := map[string]interface{}{
		"status":         health,
		"active_workers": activeWorkers,
		"total_workers":  len(workers),
		"pending_jobs":   stats.PendingJobs,
		"running_jobs":   stats.RunningJobs,
		"failed_jobs":    stats.FailedJobs,
		"completed_jobs": stats.CompletedJobs,
		"issues":         issues,
	}

	statusCode := http.StatusOK
	if health == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, models.APIResponse{
		Success: health != "unhealthy",
		Message: "Queue health status retrieved",
		Data:    response,
	})
}

// ==================== Advanced Job Operations ====================

// CreateBulkBackupJobs creates multiple backup jobs at once
func (h *WorkerHandlers) CreateBulkBackupJobs(c *gin.Context) {
	var req struct {
		Jobs []struct {
			PostgresID   string            `json:"postgres_id" binding:"required"`
			DatabaseName string            `json:"database_name" binding:"required"`
			BackupType   models.BackupType `json:"backup_type" binding:"required"`
			Priority     int               `json:"priority"`
		} `json:"jobs" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	var createdJobs []*worker.Job
	var errors []string

	for i, jobReq := range req.Jobs {
		priority := jobReq.Priority
		if priority == 0 {
			priority = 5 // Medium priority
		}

		job, err := h.jobQueue.AddBackupJob(jobReq.PostgresID, jobReq.DatabaseName, jobReq.BackupType, priority)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Job %d: %v", i+1, err))
		} else {
			createdJobs = append(createdJobs, job)
		}
	}

	response := map[string]interface{}{
		"created_jobs":    createdJobs,
		"created_count":   len(createdJobs),
		"total_requested": len(req.Jobs),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	statusCode := http.StatusCreated
	message := "Bulk backup jobs created successfully"

	if len(errors) > 0 {
		if len(createdJobs) == 0 {
			statusCode = http.StatusBadRequest
			message = "Failed to create any backup jobs"
		} else {
			statusCode = http.StatusPartialContent
			message = "Some backup jobs created with errors"
		}
	}

	c.JSON(statusCode, models.APIResponse{
		Success: len(createdJobs) > 0,
		Message: message,
		Data:    response,
	})
}

// GetQueueMetrics returns detailed queue metrics
func (h *WorkerHandlers) GetQueueMetrics(c *gin.Context) {
	stats := h.jobQueue.GetStats()
	workers := h.jobQueue.GetWorkerStatus()
	runningJobs := h.jobQueue.GetRunningJobs()

	// Calculate additional metrics
	totalWorkers := len(workers)
	idleWorkers := 0
	workingWorkers := 0

	for _, worker := range workers {
		switch worker.Status {
		case "idle":
			idleWorkers++
		case "working":
			workingWorkers++
		}
	}

	// Job type breakdown for running jobs
	jobTypeBreakdown := make(map[string]int)
	for _, job := range runningJobs {
		jobTypeBreakdown[string(job.Type)]++
	}

	metrics := map[string]interface{}{
		"queue_stats": stats,
		"worker_metrics": map[string]interface{}{
			"total":   totalWorkers,
			"idle":    idleWorkers,
			"working": workingWorkers,
			"stopped": totalWorkers - idleWorkers - workingWorkers,
		},
		"job_type_breakdown": jobTypeBreakdown,
		"queue_capacity": map[string]interface{}{
			"max_jobs":     1000, // From queue buffer size
			"current_jobs": stats.PendingJobs,
			"utilization":  float64(stats.PendingJobs) / 1000.0 * 100,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Queue metrics retrieved successfully",
		Data:    metrics,
	})
}

// ==================== Worker Control ====================

// RestartQueue restarts the job queue (admin only)
func (h *WorkerHandlers) RestartQueue(c *gin.Context) {
	if !h.jobQueue.IsRunning() {
		if err := h.jobQueue.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to start queue: " + err.Error(),
			})
			return
		}
	} else {
		// Stop and restart
		h.jobQueue.Stop()
		if err := h.jobQueue.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to restart queue: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Queue restarted successfully",
	})
}

// GetDetailedWorkerInfo returns detailed information about a specific worker
func (h *WorkerHandlers) GetDetailedWorkerInfo(c *gin.Context) {
	workerID := c.Param("worker_id")
	if workerID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Worker ID is required",
		})
		return
	}

	workers := h.jobQueue.GetWorkerStatus()

	var targetWorker *worker.WorkerStatus
	for _, w := range workers {
		if w.ID == workerID {
			worker := w // Create a copy
			targetWorker = &worker
			break
		}
	}

	if targetWorker == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Worker not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Worker information retrieved successfully",
		Data:    targetWorker,
	})
}
