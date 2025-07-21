package api

import (
	"evolution-postgres-backup/internal/service"
	"evolution-postgres-backup/internal/worker"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupV2Router creates the modern API router with SQLite backend
func SetupV2Router(dbService *service.DatabaseService, jobQueue *worker.JobQueue) *gin.Engine {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(setupCORS())

	// Initialize handlers
	v2Handlers := NewV2Handlers(dbService)
	workerHandlers := NewWorkerHandlers(jobQueue)

	// Public routes (no auth required)
	public := router.Group("/")
	{
		// Basic health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
				"version":   "2.0",
			})
		})

		// Detailed health check
		public.GET("/health/detailed", v2Handlers.GetHealthDetailed)
	}

	// API v2 routes (require authentication)
	v2 := router.Group("/api/v2")
	v2.Use(AuthMiddleware())
	{
		// ==================== Dashboard & Analytics ====================
		dashboard := v2.Group("/dashboard")
		{
			dashboard.GET("", v2Handlers.GetDashboard)
			dashboard.GET("/trends", v2Handlers.GetBackupTrends)
		}

		// ==================== PostgreSQL Instance Management ====================
		postgres := v2.Group("/postgres")
		{
			postgres.GET("", v2Handlers.GetPostgreSQLInstances) // ?enabled=true for filtered
			postgres.POST("", v2Handlers.CreatePostgreSQLInstance)
			postgres.GET("/:id", v2Handlers.GetPostgreSQLInstance)
			postgres.PUT("/:id", v2Handlers.UpdatePostgreSQLInstance)
			postgres.DELETE("/:id", v2Handlers.DeletePostgreSQLInstance)

			// Instance-specific backups
			postgres.GET("/:id/backups", v2Handlers.GetBackupsByInstance)
		}

		// ==================== Advanced Backup Management ====================
		backups := v2.Group("/backups")
		{
			// Advanced filtering: ?postgres_id=x&status=completed&type=daily&limit=10
			backups.GET("", v2Handlers.GetBackupsAdvanced)
			backups.GET("/:id", func(c *gin.Context) {
				// Delegate to database service
				backupID := c.Param("id")
				backup, err := dbService.GetBackup(backupID)
				if err != nil {
					c.JSON(404, gin.H{"error": "Backup not found"})
					return
				}
				c.JSON(200, gin.H{"success": true, "data": backup})
			})
		}

		// ==================== Advanced Log Management ====================
		logs := v2.Group("/logs")
		{
			// Advanced filtering: ?start_date=2025-07-18&level=ERROR&component=BACKUP&job_id=abc123&limit=50
			logs.GET("", v2Handlers.GetLogsAdvanced)
			logs.GET("/job/:job_id", v2Handlers.GetLogsByJobID)
			logs.GET("/backup/:backup_id", v2Handlers.GetLogsByBackupID)
			logs.GET("/stream", func(c *gin.Context) {
				// Basic streaming implementation - returns recent logs as JSON
				// Real-time streaming with SSE could be implemented later
				apiKey := c.Query("api-key")
				if apiKey == "" {
					c.JSON(401, gin.H{"error": "API key required for streaming"})
					return
				}

				// For now, return recent logs instead of real streaming
				logs, err := dbService.GetRecentLogs(50)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to fetch logs: " + err.Error()})
					return
				}

				c.JSON(200, gin.H{
					"success": true,
					"message": "Recent logs (streaming not yet implemented)",
					"data":    logs,
				})
			})
		}

		// ==================== Worker System Management ====================
		workers := v2.Group("/workers")
		{
			// Queue statistics and monitoring
			workers.GET("/stats", workerHandlers.GetQueueStats)
			workers.GET("/health", workerHandlers.GetQueueHealth)
			workers.GET("/metrics", workerHandlers.GetQueueMetrics)
			workers.POST("/restart", workerHandlers.RestartQueue) // Admin only

			// Worker status and management
			workers.GET("/status", workerHandlers.GetWorkerStatus)
			workers.GET("/:worker_id", workerHandlers.GetDetailedWorkerInfo)

			// Job management
			jobs := workers.Group("/jobs")
			{
				jobs.GET("/running", workerHandlers.GetRunningJobs)

				// Create individual jobs
				jobs.POST("/backup", workerHandlers.CreateBackupJob)
				jobs.POST("/restore", workerHandlers.CreateRestoreJob)
				jobs.POST("/cleanup", workerHandlers.CreateCleanupJob)

				// Bulk operations
				jobs.POST("/backup/bulk", workerHandlers.CreateBulkBackupJobs)
			}
		}

		// ==================== Migration Management ====================
		migration := v2.Group("/migration")
		{
			migration.GET("/status", v2Handlers.GetMigrationStatus)
			migration.POST("/execute", v2Handlers.PerformMigration) // Admin only
		}

		// ==================== System Information ====================
		system := v2.Group("/system")
		{
			system.GET("/info", func(c *gin.Context) {
				info := map[string]interface{}{
					"version":       "2.0.0",
					"database_type": "SQLite",
					"worker_system": "enabled",
					"features": []string{
						"advanced_filtering",
						"real_time_workers",
						"bulk_operations",
						"structured_logging",
						"health_monitoring",
						"migration_support",
					},
				}
				c.JSON(200, gin.H{"success": true, "data": info})
			})
		}
	}

	return router
}

// setupCORS configures CORS middleware
func setupCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "*"}, // Vite dev server + production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With", "api-key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// API Documentation endpoints
func SetupDocsRoutes(router *gin.Engine) {
	docs := router.Group("/docs")
	{
		docs.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Evolution PostgreSQL Backup Service API Documentation",
				"version": "2.0.0",
				"endpoints": map[string]interface{}{
					"health": map[string]string{
						"GET /health":          "Basic health check",
						"GET /health/detailed": "Detailed system health",
					},
					"dashboard": map[string]string{
						"GET /api/v2/dashboard":        "Dashboard statistics",
						"GET /api/v2/dashboard/trends": "Backup trends",
					},
					"postgres": map[string]string{
						"GET /api/v2/postgres":             "List PostgreSQL instances",
						"POST /api/v2/postgres":            "Create PostgreSQL instance",
						"GET /api/v2/postgres/:id":         "Get specific instance",
						"PUT /api/v2/postgres/:id":         "Update instance",
						"DELETE /api/v2/postgres/:id":      "Delete instance",
						"GET /api/v2/postgres/:id/backups": "Get instance backups",
					},
					"backups": map[string]string{
						"GET /api/v2/backups":     "List backups (with advanced filtering)",
						"GET /api/v2/backups/:id": "Get specific backup",
					},
					"logs": map[string]string{
						"GET /api/v2/logs":                   "List logs (with advanced filtering)",
						"GET /api/v2/logs/job/:job_id":       "Get logs for specific job",
						"GET /api/v2/logs/backup/:backup_id": "Get logs for specific backup",
					},
					"workers": map[string]string{
						"GET /api/v2/workers/stats":             "Queue statistics",
						"GET /api/v2/workers/health":            "Worker system health",
						"GET /api/v2/workers/metrics":           "Detailed metrics",
						"GET /api/v2/workers/status":            "Worker status",
						"POST /api/v2/workers/restart":          "Restart queue",
						"GET /api/v2/workers/jobs/running":      "Running jobs",
						"POST /api/v2/workers/jobs/backup":      "Create backup job",
						"POST /api/v2/workers/jobs/restore":     "Create restore job",
						"POST /api/v2/workers/jobs/cleanup":     "Create cleanup job",
						"POST /api/v2/workers/jobs/backup/bulk": "Create bulk backup jobs",
					},
				},
				"query_parameters": map[string]interface{}{
					"backups": []string{
						"postgres_id=uuid",
						"status=pending|in_progress|completed|failed",
						"type=hourly|daily|weekly|monthly|manual",
					},
					"logs": []string{
						"start_date=2025-07-18",
						"end_date=2025-07-19",
						"level=INFO|WARN|ERROR|DEBUG",
						"component=BACKUP|RESTORE|WORKER|QUEUE",
						"job_id=short_job_id",
						"backup_id=backup_uuid",
						"limit=100",
					},
				},
			})
		})

		docs.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"endpoints": map[string]interface{}{
					"GET /health": map[string]interface{}{
						"description": "Basic health check",
						"response": map[string]string{
							"status":    "healthy",
							"timestamp": "2025-07-18T10:30:00Z",
							"version":   "2.0",
						},
					},
					"GET /health/detailed": map[string]interface{}{
						"description": "Detailed system health with component status",
						"response": map[string]interface{}{
							"status":    "healthy",
							"timestamp": "2025-07-18T10:30:00Z",
							"components": map[string]interface{}{
								"database":             map[string]string{"status": "healthy"},
								"postgresql_instances": map[string]interface{}{"status": "healthy", "count": 2},
								"backups":              map[string]interface{}{"status": "healthy", "total": 6},
							},
						},
					},
				},
			})
		})
	}
}
