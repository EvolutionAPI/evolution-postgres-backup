package worker

import (
	"context"
	"database/sql"
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/models"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobType represents different types of jobs
type JobType string

const (
	JobTypeBackup  JobType = "backup"
	JobTypeRestore JobType = "restore"
	JobTypeCleanup JobType = "cleanup"
)

// JobStatus represents job execution status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

// Job represents a work item in the queue
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Priority    int                    `json:"priority"` // Higher = more priority
	Payload     map[string]interface{} `json:"payload"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	WorkerID    string                 `json:"worker_id,omitempty"`
}

// JobQueue manages the job queue and workers
type JobQueue struct {
	ctx         context.Context
	cancel      context.CancelFunc
	jobs        chan *Job
	workers     []*Worker
	workerCount int
	dbService   *database.DB
	logRepo     *database.LogRepository
	mu          sync.RWMutex
	running     bool
	stats       *QueueStats
}

// QueueStats tracks queue statistics
type QueueStats struct {
	TotalJobs     int64 `json:"total_jobs"`
	PendingJobs   int64 `json:"pending_jobs"`
	RunningJobs   int64 `json:"running_jobs"`
	CompletedJobs int64 `json:"completed_jobs"`
	FailedJobs    int64 `json:"failed_jobs"`
	ActiveWorkers int   `json:"active_workers"`
}

// NewJobQueue creates a new job queue
func NewJobQueue(workerCount int, dbService *database.DB) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	logRepo := database.NewLogRepository(dbService)

	return &JobQueue{
		ctx:         ctx,
		cancel:      cancel,
		jobs:        make(chan *Job, 1000), // Buffer for 1000 jobs
		workers:     make([]*Worker, 0, workerCount),
		workerCount: workerCount,
		dbService:   dbService,
		logRepo:     logRepo,
		stats:       &QueueStats{},
	}
}

// GetDB returns the database connection
func (q *JobQueue) GetDB() *database.DB {
	return q.dbService
}

// Start starts the job queue and workers
func (q *JobQueue) Start() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return fmt.Errorf("queue is already running")
	}

	// Create and start workers
	for i := 0; i < q.workerCount; i++ {
		worker := NewWorker(fmt.Sprintf("worker-%d", i+1), q.jobs, q.dbService, q.logRepo, q)
		q.workers = append(q.workers, worker)

		go worker.Start(q.ctx)
	}

	// Start statistics updater
	go q.updateStats()

	// Start job loader from database
	go q.loadJobsFromDatabase()

	q.running = true
	q.logInfo("Queue started with %d workers", q.workerCount)

	return nil
}

// Stop stops the job queue and all workers
func (q *JobQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	q.cancel()

	// Wait for workers to finish current jobs
	for _, worker := range q.workers {
		worker.Stop()
	}

	close(q.jobs)
	q.running = false
	q.logInfo("Queue stopped")
}

// AddJob adds a new job to the queue
func (q *JobQueue) AddJob(job *Job) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}

	if job.MaxRetries == 0 {
		job.MaxRetries = 3 // Default max retries
	}

	job.Status = JobStatusPending

	// Store job in database for persistence
	if err := q.persistJob(job); err != nil {
		return fmt.Errorf("failed to persist job: %w", err)
	}

	select {
	case q.jobs <- job:
		q.logInfo("Job %s (%s) added to queue", job.ID, job.Type)
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is shutting down")
	default:
		return fmt.Errorf("queue is full")
	}
}

// AddBackupJob creates and adds a backup job
func (q *JobQueue) AddBackupJob(postgresID, databaseName string, backupType models.BackupType, priority int) (*Job, error) {
	// Create job without creating backup record (backup should be created by API)
	job := &Job{
		Type:     JobTypeBackup,
		Priority: priority,
		Payload: map[string]interface{}{
			"postgres_id":   postgresID,
			"database_name": databaseName,
			"backup_type":   string(backupType),
			// backup_id will be added by API layer
		},
		MaxRetries: 3,
	}

	if err := q.AddJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

// AddRestoreJob creates and adds a restore job
func (q *JobQueue) AddRestoreJob(backupID, postgresID, databaseName string, priority int) (*Job, error) {
	job := &Job{
		Type:     JobTypeRestore,
		Priority: priority,
		Payload: map[string]interface{}{
			"backup_id":     backupID,
			"postgres_id":   postgresID,
			"database_name": databaseName,
		},
		MaxRetries: 1, // Restore jobs should not retry automatically
	}

	if err := q.AddJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

// AddCleanupJob creates and adds a cleanup job
func (q *JobQueue) AddCleanupJob(postgresID string, backupType models.BackupType, priority int) (*Job, error) {
	job := &Job{
		Type:     JobTypeCleanup,
		Priority: priority,
		Payload: map[string]interface{}{
			"postgres_id": postgresID,
			"backup_type": string(backupType),
		},
		MaxRetries: 2,
	}

	if err := q.AddJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetStats returns current queue statistics
func (q *JobQueue) GetStats() *QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Count active workers
	activeWorkers := 0
	for _, worker := range q.workers {
		if worker.IsActive() {
			activeWorkers++
		}
	}

	stats := &QueueStats{
		TotalJobs:     q.stats.TotalJobs,
		PendingJobs:   int64(len(q.jobs)),
		RunningJobs:   q.stats.RunningJobs,
		CompletedJobs: q.stats.CompletedJobs,
		FailedJobs:    q.stats.FailedJobs,
		ActiveWorkers: activeWorkers,
	}

	return stats
}

// GetRunningJobs returns currently running jobs
func (q *JobQueue) GetRunningJobs() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var runningJobs []*Job
	for _, worker := range q.workers {
		if job := worker.GetCurrentJob(); job != nil {
			runningJobs = append(runningJobs, job)
		}
	}

	return runningJobs
}

// GetWorkerStatus returns status of all workers
func (q *JobQueue) GetWorkerStatus() []WorkerStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()

	status := make([]WorkerStatus, len(q.workers))
	for i, worker := range q.workers {
		status[i] = worker.GetStatus()
	}

	return status
}

// IsRunning returns whether the queue is running
func (q *JobQueue) IsRunning() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.running
}

// persistJob saves job to database
func (q *JobQueue) persistJob(job *Job) error {
	query := `
		INSERT INTO jobs (id, type, postgres_id, database_name, backup_id, priority, payload, status, retry_count, max_retries, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Extract common fields from payload
	postgresID, _ := job.Payload["postgres_id"].(string)
	databaseName, _ := job.Payload["database_name"].(string)
	backupID, _ := job.Payload["backup_id"].(string)

	// Use NULL for payload instead of empty string
	var payloadJSON interface{} = nil

	_, err := q.dbService.Exec(query,
		job.ID,
		string(job.Type),
		postgresID,
		databaseName,
		backupID,
		job.Priority,
		payloadJSON,
		string(job.Status),
		job.RetryCount,
		job.MaxRetries,
		job.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert job into database: %w", err)
	}

	q.logInfo("Job %s (%s) persisted to database", job.ID, job.Type)
	return nil
}

// updateJobStatus updates job status in database
func (q *JobQueue) updateJobStatus(job *Job) error {
	query := `
		UPDATE jobs 
		SET status = $1, retry_count = $2, started_at = $3, completed_at = $4, error_message = $5
		WHERE id = $6
	`

	_, err := q.dbService.Exec(query,
		string(job.Status),
		job.RetryCount,
		job.StartedAt,
		job.CompletedAt,
		job.Error,
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job status in database: %w", err)
	}

	q.logInfo("Job %s status updated to %s", job.ID, job.Status)
	return nil
}

// UpdateJobStatus is a public method to update job status (used by workers)
func (q *JobQueue) UpdateJobStatus(job *Job) error {
	return q.updateJobStatus(job)
}

// updateStats periodically updates queue statistics
func (q *JobQueue) updateStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			q.refreshStats()
		case <-q.ctx.Done():
			return
		}
	}
}

// refreshStats updates internal statistics
func (q *JobQueue) refreshStats() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Count running jobs across all workers
	runningJobs := int64(0)
	for _, worker := range q.workers {
		if worker.IsActive() {
			runningJobs++
		}
	}

	q.stats.RunningJobs = runningJobs
	q.stats.PendingJobs = int64(len(q.jobs))
}

// logInfo logs informational messages
func (q *JobQueue) logInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[QUEUE] %s", message)

	// Also log to database
	entry := &database.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "QUEUE",
		Message:   message,
	}
	q.logRepo.Create(entry)
}

// logError logs error messages
func (q *JobQueue) logError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[QUEUE] ERROR: %s", message)

	// Also log to database
	entry := &database.LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Component: "QUEUE",
		Message:   message,
	}
	q.logRepo.Create(entry)
}

// loadJobsFromDatabase periodically loads pending jobs from database
func (q *JobQueue) loadJobsFromDatabase() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	q.logInfo("Database job loader started - checking every 5 seconds")

	for {
		select {
		case <-q.ctx.Done():
			q.logInfo("Database job loader stopping...")
			return
		case <-ticker.C:
			q.logInfo("Checking for pending jobs in database...")
			// Load pending jobs from database
			query := `
				SELECT id, type, postgres_id, database_name, backup_id, priority, payload, retry_count, max_retries, created_at
				FROM jobs 
				WHERE status = 'pending' 
				   OR status = 'retrying'
				   OR (status = 'running' AND started_at < NOW() - INTERVAL '5 minutes')
				ORDER BY priority DESC, created_at ASC
				LIMIT 10
			`

			rows, err := q.dbService.Query(query)
			if err != nil {
				q.logError("Failed to query pending jobs: %v", err)
				continue
			}

			var jobsLoaded int
			q.logInfo("Scanning jobs from query result...")
			for rows.Next() {
				var job Job
				var payload sql.NullString
				var createdAtStr string
				var postgresID, databaseName, backupID string

				err := rows.Scan(
					&job.ID, &job.Type, &postgresID, &databaseName,
					&backupID, &job.Priority, &payload, &job.RetryCount,
					&job.MaxRetries, &createdAtStr,
				)
				if err != nil {
					q.logError("Failed to scan job row: %v", err)
					continue
				}

				// Parse created_at
				if createdAt, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", createdAtStr); err == nil {
					job.CreatedAt = createdAt
				} else if createdAt, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", createdAtStr); err == nil {
					job.CreatedAt = createdAt
				} else {
					job.CreatedAt = time.Now()
				}

				// Create payload from database fields
				job.Payload = map[string]interface{}{
					"postgres_id":   postgresID,
					"database_name": databaseName,
					"backup_id":     backupID,
					"backup_type":   "manual", // Valid backup type from CHECK constraint
				}

				// If payload from database exists, try to parse it (for future use)
				if payload.Valid && payload.String != "" {
					// For now, we use the fields from database directly
					// In the future, we could parse JSON payload here
				}

				job.Status = JobStatusPending

				// Mark job as running in database to avoid duplicate processing
				updateQuery := `UPDATE jobs SET status = 'running', started_at = $1 WHERE id = $2 AND status = 'pending'`
				result, err := q.dbService.Exec(updateQuery, time.Now(), job.ID)
				if err != nil {
					q.logError("Failed to mark job as running: %v", err)
					continue
				}

				rowsAffected, _ := result.RowsAffected()
				if rowsAffected == 0 {
					// Job was already picked up by another worker
					continue
				}

				// Try to add job to queue (non-blocking)
				select {
				case q.jobs <- &job:
					jobsLoaded++
					q.logInfo("Loaded job %s from database (%s)", job.ID, job.Type)
				default:
					// Queue is full, mark job back as pending
					rollbackQuery := `UPDATE jobs SET status = 'pending', started_at = NULL WHERE id = $1`
					q.dbService.Exec(rollbackQuery, job.ID)
					q.logInfo("Job queue full, job %s rolled back to pending", job.ID)
				}
			}
			rows.Close()

			if jobsLoaded > 0 {
				q.logInfo("Loaded %d jobs from database", jobsLoaded)
			} else {
				q.logInfo("No pending jobs found in database")
			}
		}
	}
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
