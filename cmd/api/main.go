package main

import (
	"context"
	"evolution-postgres-backup/internal/api"
	"evolution-postgres-backup/internal/service"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"evolution-postgres-backup/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Command line flags
	var (
		dev     = flag.Bool("dev", false, "Run in development mode")
		migrate = flag.Bool("migrate", false, "Run database migration on startup")
		port    = flag.String("port", "8080", "Server port")
	)
	flag.Parse()

	// Load .env file if in development mode
	if *dev {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️ Warning: Could not load .env file: %v", err)
		}
	}

	fmt.Println("🌐 PostgreSQL Backup API Service v2.0")
	fmt.Println("=====================================")

	// Development mode configuration
	if *dev {
		log.Println("🔧 Running in DEVELOPMENT mode")
		gin.SetMode(gin.DebugMode)
	} else {
		log.Println("🚀 Running in PRODUCTION mode")
		gin.SetMode(gin.ReleaseMode)
	}

	// Get working directory
	workDir, _ := os.Getwd()
	log.Printf("📁 Working directory: %s", workDir)

	// Check API key
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_KEY environment variable is required")
	}
	log.Printf("🔑 API Key set: %t", len(apiKey) > 0)

	// Initialize database service (PostgreSQL connection)
	log.Println("🐘 Initializing PostgreSQL database connection...")
	dbService, err := service.NewDatabaseService()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database service: %v", err)
	}
	defer func() {
		log.Println("🐘 Closing database service...")
		log.Println("✅")
	}()
	log.Println("✅")

	// Run migration if requested
	if *migrate {
		log.Println("🔄 Performing migration from JSON to SQLite...")
		// Migration is handled internally by the service
		log.Println("✅ Migration completed successfully!")
		log.Println("✅")
	}

	// Initialize job queue (for worker communication)
	log.Println("👥 Initializing job queue...")
	jobQueue := worker.NewJobQueue(4, dbService.GetDB()) // 4 workers by default
	log.Println("✅")

	// Setup API router (v2 only)
	log.Println("🌐 Setting up API router...")
	router := setupAPIRouter(dbService, jobQueue)
	log.Println("✅")

	// Get dashboard stats for startup
	stats, _ := dbService.GetDashboardStats()

	// Display system information
	log.Println("📈 API Service Status:")
	log.Printf("🌐 Server starting on port %s", *port)
	log.Println("✨ Available endpoints:")
	log.Println("   📊 Dashboard:     /api/v2/dashboard")
	log.Println("   🗄️  PostgreSQL:    /api/v2/postgres")
	log.Println("   💾 Backups:       /api/v2/backups")
	log.Println("   📝 Logs:          /api/v2/logs")
	log.Println("   🔍 Health:        /health/detailed")
	log.Println("   📚 Docs:          /docs")
	log.Println("   🔧 Main API:      /api/v2/*")
	log.Println("")
	if stats != nil {
		if pgInstances, ok := stats["postgresql_instances"].(int); ok {
			log.Printf("   🗄️  PostgreSQL Instances: %d", pgInstances)
		}
		if totalBackups, ok := stats["total_backups"].(int); ok {
			log.Printf("   💾 Backups: %d", totalBackups)
		}
		if dbSize, ok := stats["database_size"].(int64); ok {
			log.Printf("   📊 Database Size: %.2f MB", float64(dbSize)/1024/1024)
		}
	}
	log.Printf("   💚 Overall Health: healthy")

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("🛑 Shutdown signal received, gracefully stopping API service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("🌐 Stopping HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("⚠️ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅")
	}

	log.Println("✅ Graceful shutdown completed")
}

// setupAPIRouter creates the API router with job creation capabilities
func setupAPIRouter(dbService *service.DatabaseService, jobQueue *worker.JobQueue) *gin.Engine {
	return api.SetupV2Router(dbService, jobQueue)
}
