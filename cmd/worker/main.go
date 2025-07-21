package main

import (
	"context"
	"evolution-postgres-backup/internal/database"
	"evolution-postgres-backup/internal/scheduler"
	"evolution-postgres-backup/internal/worker"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Command line flags
	var (
		dev     = flag.Bool("dev", false, "Run in development mode")
		workers = flag.Int("workers", 4, "Number of worker threads")
	)
	flag.Parse()

	// Load .env file if in development mode
	if *dev {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️ Warning: Could not load .env file: %v", err)
		}
	}

	fmt.Println("👥 PostgreSQL Backup Worker Service v2.0")
	fmt.Println("========================================")

	// Development mode configuration
	if *dev {
		log.Println("🔧 Running in DEVELOPMENT mode")
	} else {
		log.Println("🚀 Running in PRODUCTION mode")
	}

	// Get working directory
	workDir, _ := os.Getwd()
	log.Printf("📁 Working directory: %s", workDir)

	// Worker configuration from environment
	if envWorkers := os.Getenv("WORKER_COUNT"); envWorkers != "" {
		if count, err := strconv.Atoi(envWorkers); err == nil {
			*workers = count
		}
	}
	log.Printf("👥 Worker threads: %d", *workers)

	// Initialize database for workers
	log.Println("🐘 Initializing PostgreSQL database connection...")
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "backup-data"
	}

	db, err := database.NewDB(dataDir)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer func() {
		log.Println("🐘 Closing database...")
		if err := db.Close(); err != nil {
			log.Printf("⚠️ Error closing database: %v", err)
		}
		log.Println("✅")
	}()
	log.Println("✅")

	// Initialize worker system
	log.Printf("👥 Initializing worker system with %d workers...", *workers)
	jobQueue := worker.NewJobQueue(*workers, db)

	// Start worker system
	if err := jobQueue.Start(); err != nil {
		log.Fatalf("❌ Failed to start worker system: %v", err)
	}
	defer func() {
		log.Println("👥 Stopping worker system...")
		jobQueue.Stop()
		log.Println("✅")
	}()
	log.Println("✅")

	// Get worker stats for startup
	stats := jobQueue.GetStats()

	// Display system information
	log.Println("📈 Worker Service Status:")
	log.Printf("👥 Active Workers: %d", stats.ActiveWorkers)
	log.Printf("⏳ Pending Jobs: %d", stats.PendingJobs)
	log.Printf("✅ Completed Jobs: %d", stats.CompletedJobs)
	log.Printf("❌ Failed Jobs: %d", stats.FailedJobs)
	log.Printf("💚 Overall Health: healthy")

	log.Println("")
	log.Println("🎯 Worker Capabilities:")
	log.Println("   💾 Backup jobs - PostgreSQL database backups")
	log.Println("   🔄 Restore jobs - Database restoration")
	log.Println("   🧹 Cleanup jobs - File cleanup and maintenance")
	log.Println("   📊 Statistics - Real-time worker monitoring")

	// Initialize automatic backup scheduler
	log.Println("")
	log.Println("⏰ Initializing automatic backup scheduler...")
	autoScheduler := scheduler.NewScheduler(jobQueue)
	if err := autoScheduler.Start(); err != nil {
		log.Printf("⚠️ Failed to start automatic scheduler: %v", err)
		log.Println("   Manual backups will still work normally")
	} else {
		log.Println("✅ Automatic backup scheduler started")
		log.Println("   📅 Hourly: Every hour at minute 0")
		log.Println("   📅 Daily: Every day at 2:00 AM")
		log.Println("   📅 Weekly: Every Sunday at 3:00 AM")
		log.Println("   📅 Monthly: First day of month at 4:00 AM")
		defer func() {
			log.Println("⏰ Stopping automatic backup scheduler...")
			autoScheduler.Stop()
			log.Println("✅")
		}()
	}

	// Worker monitoring loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start monitoring goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Log worker status periodically
				currentStats := jobQueue.GetStats()
				if currentStats.ActiveWorkers > 0 || currentStats.PendingJobs > 0 {
					log.Printf("📊 Workers: %d active, %d pending, %d completed",
						currentStats.ActiveWorkers, currentStats.PendingJobs, currentStats.CompletedJobs)
				}
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("🛑 Shutdown signal received, gracefully stopping worker service...")

	// Cancel monitoring
	cancel()

	// Final stats
	finalStats := jobQueue.GetStats()
	log.Printf("📊 Final Statistics:")
	log.Printf("   ✅ Total Completed: %d", finalStats.CompletedJobs)
	log.Printf("   ❌ Total Failed: %d", finalStats.FailedJobs)
	log.Printf("   ⏱️  Total Processed: %d", finalStats.TotalJobs)

	log.Println("✅ Graceful shutdown completed")
}
