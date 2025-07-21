# PostgreSQL Backup System - Makefile

.PHONY: help build run stop logs clean test docker-build docker-up docker-down docker-rebuild docker-restart docker-logs dev-frontend dev-api dev-worker dev-3services setup-postgres

# Default target
help:
	@echo "🐳 PostgreSQL Backup System - Docker Commands"
	@echo ""
	@echo "🚀 App-only (External PostgreSQL):"
	@echo "  setup-db         Create database schema in existing PostgreSQL"
	@echo "  migrate-databases Add databases column to existing tables"
	@echo "  migrate-enabled  Add enabled column to existing tables"
	@echo "  migrate-logs     Add missing columns to logs table"
	@echo "  app-build        Build app containers (API + Worker + Frontend)"
	@echo "  app-up           Start app services (use external PostgreSQL)"
	@echo "  app-down         Stop app services"
	@echo "  app-restart      Restart app services"
	@echo "  app-logs         Show app logs"
	@echo "  app-rebuild      Rebuild and restart app services"
	@echo "  app-status       Show app services status"
	@echo ""
	@echo "📦 Docker Full Stack (with PostgreSQL):"
	@echo "  docker-build     Build both frontend and backend"
	@echo "  docker-up        Start complete system (frontend + backend)"
	@echo "  docker-down      Stop and remove all containers"
	@echo "  docker-restart   Restart all services"
	@echo "  docker-logs      Show logs from all services"
	@echo "  docker-rebuild   Rebuild and restart everything"
	@echo ""
	@echo "🔧 Individual Services:"
	@echo "  build-backend    Build only backend container"
	@echo "  build-frontend   Build only frontend container"
	@echo "  restart-backend  Restart only backend service"
	@echo "  restart-frontend Restart only frontend service"
	@echo "  logs-backend     Show backend logs"
	@echo "  logs-frontend    Show frontend logs"
	@echo ""
	@echo "💻 Development:"
	@echo "  dev-frontend     Run frontend locally (Vite)"
	@echo "  dev-api          Run API service locally"
	@echo "  dev-worker       Run Worker service locally"
	@echo "  dev-3services    Instructions for 3-services dev mode"
	@echo "  setup-postgres   Start PostgreSQL container for development"
	@echo "  test             Run backend tests"
	@echo "  clean            Clean Docker resources"
	@echo ""
	@echo "🌐 Access URLs:"
	@echo "  Frontend:  http://localhost:3000"
	@echo "  Backend:   http://localhost:8080"
	@echo "  Health:    http://localhost:8080/health"

# Docker Full Stack Commands (legacy)
docker-build-compose:
	@echo "🔨 Building frontend and backend containers with compose..."
	docker-compose build

docker-up:
	@echo "🚀 Starting complete PostgreSQL backup system..."
	docker-compose up -d
	@echo ""
	@echo "✅ System started successfully!"
	@echo "🌐 Frontend: http://localhost:3000"
	@echo "📡 API: http://localhost:3000/api/v1"
	@echo ""
	@echo "📊 Check status with: make docker-logs"

docker-down:
	@echo "🛑 Stopping all services..."
	docker-compose down

docker-restart:
	@echo "🔄 Restarting all services..."
	docker-compose restart

docker-logs:
	@echo "📋 Showing logs from all services..."
	docker-compose logs -f --tail=50

docker-rebuild:
	@echo "🔄 Rebuilding and restarting complete system..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "✅ System rebuilt and restarted!"

# Individual Service Commands
build-backend:
	@echo "🔨 Building backend container..."
	docker-compose build postgres-backup

build-frontend:
	@echo "🔨 Building frontend container..."
	docker-compose build postgres-backup-frontend

restart-backend:
	@echo "🔄 Restarting backend service..."
	docker-compose restart postgres-backup

restart-frontend:
	@echo "🔄 Restarting frontend service..."
	docker-compose restart postgres-backup-frontend

logs-backend:
	@echo "📋 Backend logs:"
	docker-compose logs -f postgres-backup

logs-frontend:
	@echo "📋 Frontend logs:"
	docker-compose logs -f postgres-backup-frontend

# Development Commands
dev-frontend:
	@echo "💻 Starting frontend in development mode..."
	cd frontend && npm run dev

dev-api:
	@echo "🌐 Starting API service in development mode..."
	go run cmd/api/main.go --port=8080 --dev

dev-worker:
	@echo "👥 Starting Worker service in development mode..."
	go run cmd/worker/main.go --workers=4 --dev

dev-3services:
	@echo "🚀 Starting 3-services architecture in development mode..."
	@echo "Run in 4 separate terminals:"
	@echo "  Terminal 1: make setup-postgres"
	@echo "  Terminal 2: make dev-api"
	@echo "  Terminal 3: make dev-worker"
	@echo "  Terminal 4: make dev-frontend"

# PostgreSQL Setup
setup-postgres:
	@echo "🐘 Starting PostgreSQL container..."
	@docker-compose up postgres-backup-db -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "✅ PostgreSQL is ready on port 5432"
	@echo "Connection: postgresql://postgres:root@localhost:5432/backup_service"

# V2 System Testing
test-v2-api:
	@echo "🧪 Testing v2 API endpoints..."
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq . || echo "❌ Health endpoint failed"
	@echo "Testing dashboard endpoint..."
	@curl -s -H "Authorization: a4f3a241-7763-4f3b-9101-0e26c5029f17" http://localhost:8080/api/v2/dashboard | jq . || echo "❌ Dashboard endpoint failed"
	@echo "Testing worker stats..."
	@curl -s -H "Authorization: a4f3a241-7763-4f3b-9101-0e26c5029f17" http://localhost:8080/api/v2/workers/stats | jq . || echo "❌ Worker stats failed"

test-v2-job:
	@echo "🎯 Creating test backup job..."
	@curl -s -X POST \
		-H "Authorization: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
		-H "Content-Type: application/json" \
		-d '{"postgres_id":"postgres-1","database_name":"evolution_lb","backup_type":"manual","priority":5}' \
		http://localhost:8080/api/v2/workers/jobs/backup | jq .

# Testing and Maintenance
test:
	@echo "🧪 Running backend tests..."
	go test ./...

clean:
	@echo "🧹 Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f
	@echo "✅ Cleanup completed!"

# Quick start alias
start: docker-up

# Quick stop alias  
stop: docker-down

# Status check
status:
	@echo "📊 System Status:"
	@echo ""
	docker-compose ps
	@echo ""
	@echo "🔍 Health Checks:"
	@curl -s http://localhost:3000/health | grep -o '"status":"[^"]*"' || echo "❌ Frontend not responding"

# Production deployment
deploy-prod:
	@echo "🚀 Deploying to production..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
	@echo "✅ Production deployment completed!"

# ================= Docker v2 Commands - 3 Services Architecture =================

docker-v2-build:
	@echo "🔨 Building v2 images (API + Worker + Frontend)..."
	docker-compose build
	@echo "✅ v2 images built successfully!"

docker-v2-up:
	@echo "🚀 Starting v2 services (API + Worker + Frontend)..."
	docker-compose up -d
	@echo "✅ v2 services started successfully!"
	@echo "📊 Frontend: http://localhost:3000"
	@echo "🌐 API: http://localhost:8080"
	@echo "👥 Worker: Background service"

docker-v2-down:
	@echo "🛑 Stopping v2 services..."
	docker-compose down
	@echo "✅ v2 services stopped!"

docker-v2-restart:
	@echo "🔄 Restarting v2 services..."
	docker-compose restart
	@echo "✅ v2 services restarted!"

docker-v2-logs:
	@echo "📝 Showing all v2 logs..."
	docker-compose logs -f --tail=50

docker-v2-logs-api:
	@echo "📝 Showing API service logs..."
	docker-compose logs -f postgres-backup-api

docker-v2-logs-worker:
	@echo "📝 Showing Worker service logs..."
	docker-compose logs -f postgres-backup-worker

docker-v2-logs-frontend:
	@echo "📝 Showing Frontend service logs..."
	docker-compose logs -f postgres-backup-frontend

docker-v2-rebuild:
	@echo "🔄 Rebuilding v2 services from scratch..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "✅ v2 services rebuilt and started!"

docker-v2-status:
	@echo "📊 v2 Services Status:"
	docker-compose ps

# Individual service commands
docker-v2-restart-api:
	@echo "🔄 Restarting API service..."
	docker-compose restart postgres-backup-api

docker-v2-restart-worker:
	@echo "🔄 Restarting Worker service..."
	docker-compose restart postgres-backup-worker

docker-v2-restart-frontend:
	@echo "🔄 Restarting Frontend service..."
	docker-compose restart postgres-backup-frontend

# Shell access
docker-v2-shell-api:
	@echo "🐚 Opening shell in API container..."
	docker-compose exec postgres-backup-api sh

docker-v2-shell-worker:
	@echo "🐚 Opening shell in Worker container..."
	docker-compose exec postgres-backup-worker sh

docker-v2-shell-frontend:
	@echo "🐚 Opening shell in Frontend container..."
	docker-compose exec postgres-backup-frontend sh

# Production v2 commands
docker-v2-prod:
	@echo "🚀 Deploying v2 to production (3 Services Architecture)..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
	@echo "✅ v2 production deployment completed!"
	@echo "📊 API: 6 workers in production mode"
	@echo "👥 Worker: High-performance configuration"
	@echo "🌐 Frontend: Production optimized"

# Test v2 Docker services
docker-v2-test:
	@echo "🧪 Testing v2 Docker services..."
	@echo "Waiting for services to start..."
	@sleep 15
	@echo "Testing API service..."
	@curl -f http://localhost:8080/health && echo "✅ API healthy" || echo "❌ API unhealthy"
	@echo "Testing Frontend..."
	@curl -f http://localhost:3000 && echo "✅ Frontend accessible" || echo "❌ Frontend inaccessible"
	@echo "Testing Worker connectivity..."
	@docker-compose exec -T postgres-backup-worker pgrep -f postgres-backup-worker && echo "✅ Worker running" || echo "❌ Worker not running"

# Service scaling (production)
docker-v2-scale-workers:
	@echo "📈 Scaling worker service..."
	docker-compose up -d --scale postgres-backup-worker=2
	@echo "✅ Worker service scaled!"

# ================= App-only Commands (for existing PostgreSQL) =================

app-build:
	@echo "🔨 Building app-only containers (API + Worker + Frontend)..."
	docker-compose -f docker-compose-app.yml build
	@echo "✅ App containers built successfully!"

app-up:
	@echo "🚀 Starting app services (using external PostgreSQL)..."
	@echo "Assuming PostgreSQL is running on localhost:5432"
	docker-compose -f docker-compose-app.yml up -d
	@echo "✅ App services started successfully!"
	@echo "📊 Frontend: http://localhost:3000"
	@echo "🌐 API: http://localhost:8080"
	@echo "👥 Worker: Background service"
	@echo "⚠️  Make sure PostgreSQL has the backup_service database!"

app-down:
	@echo "🛑 Stopping app services..."
	docker-compose -f docker-compose-app.yml down
	@echo "✅ App services stopped!"

app-restart:
	@echo "🔄 Restarting app services..."
	docker-compose -f docker-compose-app.yml restart
	@echo "✅ App services restarted!"

app-logs:
	@echo "📝 Showing app logs..."
	docker-compose -f docker-compose-app.yml logs -f --tail=50

app-rebuild:
	@echo "🔄 Rebuilding app services from scratch..."
	docker-compose -f docker-compose-app.yml down
	docker-compose -f docker-compose-app.yml build --no-cache
	docker-compose -f docker-compose-app.yml up -d
	@echo "✅ App services rebuilt and started!"

app-status:
	@echo "📊 App Services Status:"
	docker-compose -f docker-compose-app.yml ps

setup-db:
	@echo "🗄️  Setting up database schema in existing PostgreSQL..."
	@echo "Make sure PostgreSQL is running and accessible!"
	@echo "Creating database and schema..."
	@if command -v psql >/dev/null 2>&1; then \
		createdb -h localhost -p 5432 -U postgres backup_service 2>/dev/null || echo "Database backup_service already exists"; \
		psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/schema_postgres.sql; \
		echo "✅ Database schema created successfully!"; \
	else \
		echo "❌ psql not found. Please install PostgreSQL client or run manually:"; \
		echo "createdb -h localhost -p 5432 -U postgres backup_service"; \
		echo "psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/schema_postgres.sql"; \
	fi

migrate-databases:
	@echo "🔄 Adding databases column to existing postgresql_instances table..."
	@echo "Make sure PostgreSQL is running and accessible!"
	@if command -v psql >/dev/null 2>&1; then \
		psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_databases.sql; \
		echo "✅ Databases column migration completed successfully!"; \
	else \
		echo "❌ psql not found. Please install PostgreSQL client or run manually:"; \
		echo "psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_databases.sql"; \
	fi

migrate-enabled:
	@echo "🔄 Adding enabled column to existing postgresql_instances table..."
	@echo "Make sure PostgreSQL is running and accessible!"
	@if command -v psql >/dev/null 2>&1; then \
		psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_enabled.sql; \
		echo "✅ Enabled column migration completed successfully!"; \
	else \
		echo "❌ psql not found. Please install PostgreSQL client or run manually:"; \
		echo "psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_enabled.sql"; \
	fi

# Migrate logs table (add details and created_at columns)
migrate-logs:
	@echo "🔄 Applying logs table migration..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_logs_columns.sql
	@echo "✅ Logs migration completed"

# Migrate jobs table (add jobs table for worker system)
migrate-jobs:
	@echo "🔄 Adding jobs table for worker system..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -f internal/database/migrate_add_jobs_table.sql
	@echo "✅ Jobs table migration completed" 

# Debug: Check jobs in database
debug-jobs:
	@echo "🔍 Checking jobs in database..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -c "SELECT id, type, status, postgres_id, database_name, created_at FROM jobs ORDER BY created_at DESC LIMIT 10;"

# Debug: Check backups in database
debug-backups:
	@echo "🔍 Checking backups in database..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -c "SELECT id, postgresql_id, status, job_id, created_at FROM backups ORDER BY created_at DESC LIMIT 10;"

# Debug: Check logs
debug-logs:
	@echo "🔍 Checking logs in database..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -c "SELECT component, job_id, message, timestamp FROM logs ORDER BY timestamp DESC LIMIT 20;"

# Reset orphaned jobs (running for more than 5 minutes)
reset-orphaned-jobs:
	@echo "🔄 Resetting orphaned jobs..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -c "UPDATE jobs SET status = 'pending', started_at = NULL WHERE status = 'running' AND started_at < NOW() - INTERVAL '5 minutes';"
	@echo "✅ Orphaned jobs reset"

# Check enabled PostgreSQL instances (for automatic backups)
check-enabled-instances:
	@echo "🔍 Checking enabled PostgreSQL instances for automatic backups..."
	PGPASSWORD=root psql -h localhost -p 5432 -U postgres -d backup_service -c "SELECT id, name, host, port, databases, enabled FROM postgresql_instances WHERE enabled = true;"

# Test scheduler manually (create hourly backup for all enabled instances)
test-scheduler:
	@echo "🧪 Testing automatic backup scheduler..."
	@echo "This will create hourly backup jobs for all enabled instances"
	@echo "Check logs and job queue after running this"

# =============================================================================
# 🐳 DOCKER REGISTRY COMMANDS
# =============================================================================

# Build Docker images locally
docker-build:
	@echo "🐳 Building Docker images locally..."
	./scripts/build-images.sh latest "linux/amd64"

# Build Docker images for multiple platforms
docker-build-multi:
	@echo "🐳 Building Docker images for multiple platforms..."
	./scripts/build-images.sh latest "linux/amd64,linux/arm64"

# Build and push Docker images to registry
docker-push:
	@echo "🚀 Building and pushing Docker images to registry..."
	./scripts/build-images.sh latest "linux/amd64,linux/arm64" true

# Pull latest images from registry
docker-pull:
	@echo "📥 Pulling latest images from registry..."
	docker pull ghcr.io/your-username/evolution-postgres-backup-api:latest
	docker pull ghcr.io/your-username/evolution-postgres-backup-worker:latest
	docker pull ghcr.io/your-username/evolution-postgres-backup-frontend:latest

# Run with registry images
docker-registry-up:
	@echo "🚀 Starting services with registry images..."
	docker-compose -f docker-compose.registry.yml up -d

# Stop registry services
docker-registry-down:
	@echo "🛑 Stopping registry services..."
	docker-compose -f docker-compose.registry.yml down

# Test registry images
docker-test:
	@echo "🧪 Testing registry images..."
	docker-compose -f docker-compose.registry.yml up --abort-on-container-exit

# Clean up Docker images
docker-clean:
	@echo "🧹 Cleaning up Docker images..."
	docker image prune -f
	docker system prune -f

# Show Docker images info
docker-info:
	@echo "📊 Docker images information:"
	@echo ""
	@echo "Local images:"
	docker images | grep evolution-postgres-backup || echo "No local images found"
	@echo ""
	@echo "Running containers:"
	docker ps | grep evolution-postgres-backup || echo "No running containers found" 