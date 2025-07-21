package scheduler

import (
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/models"
	"evolution-postgres-backup/internal/worker"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron      *cron.Cron
	jobQueue  *worker.JobQueue
	dbService *database.DB
}

func NewScheduler(jobQueue *worker.JobQueue) *Scheduler {
	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		jobQueue:  jobQueue,
		dbService: jobQueue.GetDB(),
	}
}

func (s *Scheduler) Start() error {
	// Hourly backups - every hour at minute 0
	_, err := s.cron.AddFunc("0 0 * * * *", func() {
		log.Println("🕐 Starting automatic hourly backup jobs")
		count := s.createBackupJobsForAllEnabledInstances(models.BackupTypeHourly)
		log.Printf("✅ Created %d hourly backup jobs", count)
	})
	if err != nil {
		return err
	}

	// Daily backups - every day at 2:00 AM
	_, err = s.cron.AddFunc("0 0 2 * * *", func() {
		log.Println("🌅 Starting automatic daily backup jobs")
		count := s.createBackupJobsForAllEnabledInstances(models.BackupTypeDaily)
		log.Printf("✅ Created %d daily backup jobs", count)
	})
	if err != nil {
		return err
	}

	// Weekly backups - every Sunday at 3:00 AM
	_, err = s.cron.AddFunc("0 0 3 * * 0", func() {
		log.Println("📅 Starting automatic weekly backup jobs")
		count := s.createBackupJobsForAllEnabledInstances(models.BackupTypeWeekly)
		log.Printf("✅ Created %d weekly backup jobs", count)
	})
	if err != nil {
		return err
	}

	// Monthly backups - first day of month at 4:00 AM
	_, err = s.cron.AddFunc("0 0 4 1 * *", func() {
		log.Println("📆 Starting automatic monthly backup jobs")
		count := s.createBackupJobsForAllEnabledInstances(models.BackupTypeMonthly)
		log.Printf("✅ Created %d monthly backup jobs", count)
	})
	if err != nil {
		return err
	}

	// Start the cron scheduler
	s.cron.Start()
	log.Println("⏰ Automatic backup scheduler started successfully")

	// Log all scheduled jobs
	entries := s.cron.Entries()
	log.Printf("📋 Scheduled %d automatic backup jobs", len(entries))

	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("⏰ Automatic backup scheduler stopped")
}

func (s *Scheduler) AddCustomJob(spec string, backupType models.BackupType) error {
	_, err := s.cron.AddFunc(spec, func() {
		log.Printf("🔧 Starting custom %s backup jobs", backupType)
		count := s.createBackupJobsForAllEnabledInstances(backupType)
		log.Printf("✅ Created %d custom %s backup jobs", count, backupType)
	})
	return err
}

// createBackupJobsForAllEnabledInstances creates backup jobs for all enabled PostgreSQL instances
func (s *Scheduler) createBackupJobsForAllEnabledInstances(backupType models.BackupType) int {
	// Get all enabled PostgreSQL instances
	pgRepo := database.NewPostgreSQLRepository(s.dbService)
	instances, err := pgRepo.GetEnabled()
	if err != nil {
		log.Printf("❌ Failed to get enabled PostgreSQL instances: %v", err)
		return 0
	}

	jobsCreated := 0
	backupRepo := database.NewBackupRepository(s.dbService)

	for _, instance := range instances {
		// Create backup jobs for each database in this instance
		databases := instance.Databases
		if len(databases) == 0 {
			// Fallback to default database if none specified
			databases = []string{"postgres"}
		}

		for _, dbName := range databases {
			// Create backup record first (same as API does)
			backup := &models.BackupInfo{
				ID:           fmt.Sprintf("backup_%d", time.Now().UnixNano()),
				PostgreSQLID: instance.ID,
				DatabaseName: dbName,
				BackupType:   backupType,
				Status:       models.BackupStatusPending,
				StartTime:    time.Now(),
				CreatedAt:    time.Now(),
			}

			// Save backup record to database
			if err := backupRepo.Create(backup); err != nil {
				log.Printf("❌ Failed to create backup record for %s/%s: %v", instance.Name, dbName, err)
				continue
			}

			// Create job with backup_id
			job := &worker.Job{
				Type:     worker.JobTypeBackup,
				Priority: 7, // High priority for automatic backups
				Payload: map[string]interface{}{
					"postgres_id":   instance.ID,
					"database_name": dbName,
					"backup_type":   string(backupType),
					"backup_id":     backup.ID, // Include backup_id for worker
				},
				MaxRetries: 3,
			}

			// Add job to queue
			if err := s.jobQueue.AddJob(job); err != nil {
				log.Printf("❌ Failed to create %s backup job for %s/%s: %v", backupType, instance.Name, dbName, err)
				continue
			}

			// Associate backup with job and update
			backup.JobID = job.ID
			if err := backupRepo.Update(backup); err != nil {
				// Don't fail, just log the error
				log.Printf("⚠️ Failed to update backup with job_id for %s/%s: %v", instance.Name, dbName, err)
			}

			log.Printf("📋 Created %s backup job for %s/%s (job: %s, backup: %s)", backupType, instance.Name, dbName, job.ID, backup.ID)
			jobsCreated++
		}
	}

	return jobsCreated
}
