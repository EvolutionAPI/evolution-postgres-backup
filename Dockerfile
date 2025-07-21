FROM golang:1.24-alpine AS builder

# Instalar dependências de build incluindo gcc para SQLite
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the v2 application with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o postgres-backup-v2 cmd/server_v2/main.go

# Start a new stage from scratch
FROM alpine:latest

# Install PostgreSQL client, SQLite and CA certificates
RUN apk --no-cache add postgresql-client sqlite ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the v2 binary from builder stage
COPY --from=builder /app/postgres-backup-v2 ./postgres-backup-v2
COPY --from=builder /app/.env.example .

# Create required directories
RUN mkdir -p /tmp/postgres-backups /app/data /app/db && \
    chown -R appuser:appgroup /app /tmp/postgres-backups

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run v2 server
CMD ["./postgres-backup-v2", "--migrate", "--workers=4"] 