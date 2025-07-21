package main

import (
	"context"
	"evolution-postgres-backup/internal/api"
	"evolution-postgres-backup/internal/service"
	"evolution-postgres-backup/internal/worker"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	var (
		devMode     = flag.Bool("dev", false, "Run in development mode")
		port        = flag.String("port", "", "Server port (overrides env)")
		workerCount = flag.Int("workers", 4, "Number of worker threads")
		migrate     = flag.Bool("migrate", false, "Perform migration on startup")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		if *devMode {
			log.Println("⚠️  No .env file found - make sure to create one from .env.example")
		} else {
			log.Println("No .env file found")
		}
	}

	fmt.Println("🚀 PostgreSQL Backup Service v2.0 - SQLite + Workers")
	fmt.Println("=====================================================")

	if *devMode {
		log.Println("🔧 Running in DEVELOPMENT mode")
		log.Printf("📁 Working directory: %s", getWorkingDir())
		log.Printf("🔑 API Key set: %t", os.Getenv("API_KEY") != "")
		log.Printf("👥 Worker threads: %d", *workerCount)
	}

	// Initialize SQLite database service
	fmt.Print("📊 Initializing SQLite database... ")
	dbService, err := service.NewDatabaseService()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database service: %v", err)
	}
	defer dbService.Close()
	fmt.Println("✅")

	// Perform migration if requested
	if *migrate {
		fmt.Print("🔄 Performing migration from JSON to SQLite... ")
		if err := dbService.PerformMigration(); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		fmt.Println("✅")
	}

	// Initialize worker system
	fmt.Printf("👥 Initializing worker system with %d workers... ", *workerCount)
	jobQueue := worker.NewJobQueue(*workerCount, dbService.GetDB())
	if err := jobQueue.Start(); err != nil {
		log.Fatalf("❌ Failed to start worker system: %v", err)
	}
	defer jobQueue.Stop()
	fmt.Println("✅")

	// Setup router with integrated APIs
	fmt.Print("🌐 Setting up API router... ")
	router := api.SetupV2Router(dbService, jobQueue)
	api.SetupDocsRoutes(router)
	fmt.Println("✅")

	// Get server port
	serverPort := os.Getenv("PORT")
	if *port != "" {
		serverPort = *port
	}
	if serverPort == "" {
		serverPort = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("🌐 Server starting on port %s\n", serverPort)
		fmt.Println("✨ Available endpoints:")
		fmt.Println("   📊 Dashboard:     /api/v2/dashboard")
		fmt.Println("   🗄️  PostgreSQL:    /api/v2/postgres")
		fmt.Println("   💾 Backups:       /api/v2/backups")
		fmt.Println("   📝 Logs:          /api/v2/logs")
		fmt.Println("   👥 Workers:       /api/v2/workers")
		fmt.Println("   🔍 Health:        /health/detailed")
		fmt.Println("   📚 Docs:          /docs")
		fmt.Println("   🔧 Main API:      /api/v2/*")
		fmt.Println("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Print system status
	fmt.Println("📈 System Status:")
	printSystemStatus(dbService, jobQueue)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutdown signal received, gracefully stopping...")

	// Graceful shutdown
	gracefulShutdown(server, jobQueue, dbService)
}

// printSystemStatus displays current system status
func printSystemStatus(dbService *service.DatabaseService, jobQueue *worker.JobQueue) {
	// Database stats
	stats, err := dbService.GetDashboardStats()
	if err != nil {
		fmt.Printf("   ❌ Database: Error getting stats\n")
	} else {
		if pgStats, ok := stats["postgresql"].(map[string]interface{}); ok {
			fmt.Printf("   🗄️  PostgreSQL Instances: %v\n", pgStats["total_instances"])
		}
		if backupStats, ok := stats["backups"].(map[string]interface{}); ok {
			fmt.Printf("   💾 Backups: %v\n", getMapValue(backupStats, "total", 0))
		}
		if dbInfo, ok := stats["database"].(map[string]interface{}); ok {
			if size, exists := dbInfo["file_size_mb"]; exists {
				fmt.Printf("   📊 Database Size: %.2f MB\n", size)
			}
		}
	}

	// Worker stats
	workerStats := jobQueue.GetStats()
	fmt.Printf("   👥 Workers: %d active, %d pending jobs, %d completed\n",
		workerStats.ActiveWorkers, workerStats.PendingJobs, workerStats.CompletedJobs)

	// Health check
	health := dbService.HealthCheck()
	overallHealth := "healthy"
	for _, component := range health {
		if comp, ok := component.(map[string]interface{}); ok {
			if status, exists := comp["status"]; exists && status != "healthy" {
				overallHealth = "unhealthy"
				break
			}
		}
	}
	fmt.Printf("   💚 Overall Health: %s\n", overallHealth)
}

// gracefulShutdown performs graceful shutdown
func gracefulShutdown(server *http.Server, jobQueue *worker.JobQueue, dbService *service.DatabaseService) {
	// Set timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	fmt.Print("🌐 Stopping HTTP server... ")
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Println("✅")
	}

	// Stop worker system
	fmt.Print("👥 Stopping worker system... ")
	jobQueue.Stop()
	fmt.Println("✅")

	// Close database
	fmt.Print("📊 Closing database... ")
	if err := dbService.Close(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Println("✅")
	}

	fmt.Println("✅ Graceful shutdown completed")
}

// getWorkingDir returns current working directory
func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

// getMapValue safely gets a value from a map with default
func getMapValue(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, exists := m[key]; exists {
		return value
	}
	return defaultValue
}
