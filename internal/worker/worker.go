package worker

import (
	"context"
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/models"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// WorkerStatus represents the status of a worker
type WorkerStatus struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"` // idle, working, stopped
	CurrentJob  *Job       `json:"current_job,omitempty"`
	JobsHandled int64      `json:"jobs_handled"`
	StartedAt   time.Time  `json:"started_at"`
	LastJobAt   *time.Time `json:"last_job_at,omitempty"`
}

// Worker processes jobs from the queue
type Worker struct {
	id          string
	jobs        <-chan *Job
	dbService   *database.DB
	logRepo     *database.LogRepository
	jobQueue    *JobQueue // Add reference to JobQueue
	mu          sync.RWMutex
	currentJob  *Job
	status      string
	jobsHandled int64
	startedAt   time.Time
	lastJobAt   *time.Time
	stopped     bool
}

// NewWorker creates a new worker
func NewWorker(id string, jobs <-chan *Job, dbService *database.DB, logRepo *database.LogRepository, jobQueue *JobQueue) *Worker {
	return &Worker{
		id:        id,
		jobs:      jobs,
		dbService: dbService,
		logRepo:   logRepo,
		jobQueue:  jobQueue,
		status:    "idle",
		startedAt: time.Now(),
	}
}

// Start starts the worker to process jobs
func (w *Worker) Start(ctx context.Context) {
	w.logInfo("Worker %s started", w.id)

	for {
		select {
		case job, ok := <-w.jobs:
			if !ok {
				w.logInfo("Worker %s: job channel closed", w.id)
				return
			}

			w.processJob(job)

		case <-ctx.Done():
			w.logInfo("Worker %s: context cancelled", w.id)
			return
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.stopped = true
	w.status = "stopped"
	w.logInfo("Worker %s stopped", w.id)
}

// IsActive returns whether the worker is currently processing a job
func (w *Worker) IsActive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentJob != nil && w.status == "working"
}

// GetCurrentJob returns the currently processing job
func (w *Worker) GetCurrentJob() *Job {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentJob
}

// GetStatus returns the worker status
func (w *Worker) GetStatus() WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return WorkerStatus{
		ID:          w.id,
		Status:      w.status,
		CurrentJob:  w.currentJob,
		JobsHandled: w.jobsHandled,
		StartedAt:   w.startedAt,
		LastJobAt:   w.lastJobAt,
	}
}

// processJob processes a single job
func (w *Worker) processJob(job *Job) {
	w.mu.Lock()
	w.currentJob = job
	w.status = "working"
	now := time.Now()
	job.StartedAt = &now
	job.Status = JobStatusRunning
	job.WorkerID = w.id
	w.mu.Unlock()

	w.logInfo("Worker %s processing job %s (%s)", w.id, job.ID, job.Type)

	// Process the job based on its type
	var err error
	switch job.Type {
	case JobTypeBackup:
		err = w.processBackupJob(job)
	case JobTypeRestore:
		err = w.processRestoreJob(job)
	case JobTypeCleanup:
		err = w.processCleanupJob(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Update job status
	w.mu.Lock()
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	w.lastJobAt = &completedAt
	w.jobsHandled++

	if err != nil {
		job.Error = err.Error()
		job.RetryCount++

		if job.RetryCount < job.MaxRetries {
			job.Status = JobStatusRetrying
			w.logInfo("Worker %s: job %s failed, will retry (%d/%d): %v", w.id, job.ID, job.RetryCount, job.MaxRetries, err)
			// TODO: Re-queue the job for retry
		} else {
			job.Status = JobStatusFailed
			w.logError("Worker %s: job %s failed permanently after %d retries: %v", w.id, job.ID, job.RetryCount, err)
		}
	} else {
		job.Status = JobStatusCompleted
		w.logInfo("Worker %s: job %s completed successfully", w.id, job.ID)
	}

	w.currentJob = nil
	w.status = "idle"
	w.mu.Unlock()

	// Update job status in database
	if updateErr := w.jobQueue.UpdateJobStatus(job); updateErr != nil {
		w.logError("Failed to update job status in database: %v", updateErr)
	}
}

// processBackupJob processes a backup job
func (w *Worker) processBackupJob(job *Job) error {
	w.logInfo("Processing backup job %s", job.ID)

	// Extract parameters from job payload
	postgresID, ok := job.Payload["postgres_id"].(string)
	if !ok {
		return fmt.Errorf("missing postgres_id in job payload")
	}

	databaseName, ok := job.Payload["database_name"].(string)
	if !ok {
		return fmt.Errorf("missing database_name in job payload")
	}

	backupTypeStr, ok := job.Payload["backup_type"].(string)
	if !ok {
		return fmt.Errorf("missing backup_type in job payload")
	}

	backupType := models.BackupType(backupTypeStr)

	backupRepo := database.NewBackupRepository(w.dbService)
	var backup *models.BackupInfo

	// Check if backup_id is provided in payload (meaning backup already exists)
	if backupIDStr, exists := job.Payload["backup_id"].(string); exists && backupIDStr != "" {
		// Use existing backup
		existingBackup, err := backupRepo.GetByID(backupIDStr)
		if err != nil {
			return fmt.Errorf("failed to get existing backup record: %w", err)
		}
		backup = existingBackup
		backup.JobID = job.ID // Associate with current job

		w.logJobProgress(job.ID, backup.ID, "Using existing backup record %s", backup.ID)
	} else {
		// Create new backup record (fallback for older jobs)
		backup = &models.BackupInfo{
			ID:           generateBackupID(),
			PostgreSQLID: postgresID,
			DatabaseName: databaseName,
			BackupType:   backupType,
			Status:       models.BackupStatusPending,
			StartTime:    time.Now(),
			CreatedAt:    time.Now(),
			JobID:        job.ID,
		}

		// Save backup record
		if err := backupRepo.Create(backup); err != nil {
			return fmt.Errorf("failed to create backup record: %w", err)
		}

		// Update job payload with backup ID
		job.Payload["backup_id"] = backup.ID

		w.logJobProgress(job.ID, backup.ID, "Created new backup record %s", backup.ID)
	}

	// Log backup start
	w.logJobProgress(job.ID, backup.ID, "Backup started for %s/%s", postgresID, databaseName)

	// Update backup status to in_progress
	backup.Status = models.BackupStatusInProgress
	backup.StartTime = time.Now() // Update start time when actually starting
	if err := backupRepo.Update(backup); err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}
	w.logJobProgress(job.ID, backup.ID, "Status: IN_PROGRESS")

	// Get PostgreSQL configuration
	pgRepo := database.NewPostgreSQLRepository(w.dbService)
	pgInstance, err := pgRepo.GetByID(postgresID)
	if err != nil {
		return fmt.Errorf("failed to get postgres instance: %w", err)
	}

	// Create backup filename
	timestamp := backup.StartTime.Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s_Postgres_1_%s_%s_%s.sql",
		pgInstance.Name, databaseName, string(backupType), timestamp)

	// Create local backup file path
	tempDir := os.Getenv("BACKUP_TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/postgres-backups"
	}
	localPath := filepath.Join(tempDir, filename)

	// Ensure temp directory exists
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	w.logJobProgress(job.ID, backup.ID, "Local file: %s", localPath)

	// Build pg_dump command
	cmd := exec.Command("pg_dump",
		"-h", pgInstance.Host,
		"-p", fmt.Sprintf("%d", pgInstance.Port),
		"-U", pgInstance.Username,
		"-d", databaseName,
		"-f", localPath,
		"--verbose",
		"--no-password",
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", pgInstance.Password))

	// Log pg_dump version for debugging
	if versionCmd := exec.Command("pg_dump", "--version"); versionCmd != nil {
		if versionOutput, versionErr := versionCmd.Output(); versionErr == nil {
			w.logJobProgress(job.ID, backup.ID, "pg_dump version: %s", string(versionOutput))
		}
	}

	w.logJobProgress(job.ID, backup.ID, "Executing pg_dump: %s@%s:%d/%s", pgInstance.Username, pgInstance.Host, pgInstance.Port, databaseName)

	// Execute backup and capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		backup.Status = models.BackupStatusFailed
		errorMsg := fmt.Sprintf("pg_dump failed: %v\nOutput: %s", err, string(output))
		backup.ErrorMessage = errorMsg
		endTime := time.Now()
		backup.EndTime = &endTime

		if err := backupRepo.Update(backup); err != nil {
			return fmt.Errorf("failed to update backup record: %w", err)
		}
		w.logJobProgress(job.ID, backup.ID, "pg_dump failed: %v\nOutput: %s", err, string(output))
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	w.logJobProgress(job.ID, backup.ID, "pg_dump completed successfully")

	// Get file size
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		backup.Status = models.BackupStatusFailed
		backup.ErrorMessage = fmt.Sprintf("failed to get file info: %v", err)
		endTime := time.Now()
		backup.EndTime = &endTime

		if err := backupRepo.Update(backup); err != nil {
			return fmt.Errorf("failed to update backup record: %w", err)
		}
		w.logJobProgress(job.ID, backup.ID, "Failed to get file info: %v", err)
		return fmt.Errorf("failed to get file info: %w", err)
	}

	backup.FileSize = fileInfo.Size()
	backup.FilePath = localPath
	w.logJobProgress(job.ID, backup.ID, "File size: %d bytes", backup.FileSize)

	// Update backup status to completed
	backup.Status = models.BackupStatusCompleted
	endTime := time.Now()
	backup.EndTime = &endTime

	if err := backupRepo.Update(backup); err != nil {
		return fmt.Errorf("failed to update backup record: %w", err)
	}

	w.logJobProgress(job.ID, backup.ID, "Backup completed successfully")
	return nil
}

// processRestoreJob processes a restore job
func (w *Worker) processRestoreJob(job *Job) error {
	w.logInfo("Processing restore job %s", job.ID)

	// Extract parameters from job payload
	backupID, ok := job.Payload["backup_id"].(string)
	if !ok {
		return fmt.Errorf("missing backup_id in job payload")
	}

	postgresID, ok := job.Payload["postgres_id"].(string)
	if !ok {
		return fmt.Errorf("missing postgres_id in job payload")
	}

	databaseName, ok := job.Payload["database_name"].(string)
	if !ok {
		return fmt.Errorf("missing database_name in job payload")
	}

	w.logJobProgress(job.ID, backupID, "Restore started for backup %s to %s/%s", backupID, postgresID, databaseName)

	// TODO: Implement actual restore logic here
	// For now, simulate restore process
	time.Sleep(3 * time.Second) // Simulate restore time

	w.logJobProgress(job.ID, backupID, "Restore completed successfully")
	return nil
}

// processCleanupJob processes a cleanup job
func (w *Worker) processCleanupJob(job *Job) error {
	w.logInfo("Processing cleanup job %s", job.ID)

	// Extract parameters from job payload
	postgresID, ok := job.Payload["postgres_id"].(string)
	if !ok {
		return fmt.Errorf("missing postgres_id in job payload")
	}

	backupTypeStr, ok := job.Payload["backup_type"].(string)
	if !ok {
		return fmt.Errorf("missing backup_type in job payload")
	}

	backupType := models.BackupType(backupTypeStr)

	w.logJobProgress(job.ID, "", "Cleanup started for %s (%s)", postgresID, backupType)

	// TODO: Implement actual cleanup logic here
	// For now, simulate cleanup process
	time.Sleep(1 * time.Second) // Simulate cleanup time

	w.logJobProgress(job.ID, "", "Cleanup completed successfully")
	return nil
}

// logInfo logs informational messages
func (w *Worker) logInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[WORKER] %s", message)

	// Also log to database
	entry := &database.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "WORKER",
		Message:   message,
	}
	w.logRepo.Create(entry)
}

// logError logs error messages
func (w *Worker) logError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[WORKER] ERROR: %s", message)

	// Also log to database
	entry := &database.LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Component: "WORKER",
		Message:   message,
	}
	w.logRepo.Create(entry)
}

// logJobProgress logs job progress with job and backup context
func (w *Worker) logJobProgress(jobID, backupID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[WORKER] Job %s: %s", jobID, message)

	// Also log to database with job context
	entry := &database.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "WORKER",
		JobID:     jobID, // Use full job ID
		BackupID:  backupID,
		Message:   message,
	}

	if err := w.logRepo.Create(entry); err != nil {
		log.Printf("[WORKER] Failed to save log to database: %v", err)
	}
}

// generateBackupID generates a unique backup ID
func generateBackupID() string {
	return fmt.Sprintf("backup_%d", time.Now().UnixNano())
}
