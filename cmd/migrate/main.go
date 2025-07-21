package main

import (
	"evolution-postgres-backup/internal/database"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	var (
		dataDir    = flag.String("data-dir", "backup-data", "Data directory containing JSON files")
		dryRun     = flag.Bool("dry-run", false, "Show migration status without performing migration")
		skipBackup = flag.Bool("skip-backup", false, "Skip creating backup of JSON files")
		force      = flag.Bool("force", false, "Force migration even if data already exists")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	fmt.Println("🔄 PostgreSQL Backup Service - Database Migration")
	fmt.Println("==================================================")
	fmt.Printf("📁 Data directory: %s\n", *dataDir)

	// Initialize database
	db, err := database.NewDB(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize migration service
	migrationService := database.NewMigrationService(db, *dataDir)

	// Check migration status
	status, err := migrationService.GetMigrationStatus()
	if err != nil {
		log.Fatalf("❌ Failed to check migration status: %v", err)
	}

	fmt.Println("\n📊 Current Database Status:")
	fmt.Printf("   PostgreSQL Instances: %v\n", status["postgresql_instances"])
	fmt.Printf("   Backups: %v\n", status["backups"])
	fmt.Printf("   Has Logs: %v\n", status["has_logs"])

	jsonFiles := status["json_files_exist"].(map[string]bool)
	fmt.Println("\n📋 JSON Files Available:")
	fmt.Printf("   config.json: %v\n", jsonFiles["config.json"])
	fmt.Printf("   backups.json: %v\n", jsonFiles["backups.json"])
	fmt.Printf("   logs directory: %v\n", jsonFiles["logs"])

	// Dry run mode
	if *dryRun {
		fmt.Println("\n🔍 Dry run mode - no changes will be made")
		return
	}

	// Check if data already exists and force flag is not set
	if !*force {
		if status["postgresql_instances"].(int) > 0 || status["backups"].(int) > 0 {
			fmt.Println("\n⚠️  Database already contains data!")
			fmt.Println("Use --force flag to proceed with migration anyway")
			fmt.Println("Use --dry-run to check what would be migrated")
			return
		}
	}

	// Check if any JSON files exist to migrate
	hasAnyJsonFiles := jsonFiles["config.json"] || jsonFiles["backups.json"] || jsonFiles["logs"]
	if !hasAnyJsonFiles {
		fmt.Println("\n⚠️  No JSON files found to migrate")
		return
	}

	// Create backup of JSON files before migration
	if !*skipBackup {
		fmt.Println("\n📂 Creating backup of JSON files...")
		if err := migrationService.CreateBackupDatabase(); err != nil {
			log.Printf("⚠️  Failed to create backup: %v", err)
		}
	}

	// Perform migration
	fmt.Println("\n🚀 Starting migration...")
	if err := migrationService.MigrateAll(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Show final status
	fmt.Println("\n🎉 Migration completed successfully!")

	finalStatus, err := migrationService.GetMigrationStatus()
	if err != nil {
		log.Printf("⚠️  Failed to get final status: %v", err)
		return
	}

	fmt.Println("\n📊 Final Database Status:")
	fmt.Printf("   PostgreSQL Instances: %v\n", finalStatus["postgresql_instances"])
	fmt.Printf("   Backups: %v\n", finalStatus["backups"])
	fmt.Printf("   Has Logs: %v\n", finalStatus["has_logs"])

	// Show database statistics
	stats, err := db.GetStats()
	if err == nil {
		fmt.Println("\n📈 Database Statistics:")
		fmt.Printf("   File Size: %.2f MB\n", stats["file_size_mb"])
		fmt.Printf("   Schema Version: %v\n", stats["schema_version"])
	}

	fmt.Println("\n✅ You can now start the backup service with SQLite database!")
}
