#!/bin/bash

# Simple development server with file watching
# No external dependencies required

echo "🚀 Starting development server..."

# Function to run the server
run_server() {
    echo "📦 Building and starting server..."
    go run cmd/server/main.go -dev &
    SERVER_PID=$!
    echo "🔗 Server PID: $SERVER_PID"
    echo "🌐 Server: http://localhost:8080"
    echo "📊 Health: http://localhost:8080/health"
    echo ""
}

# Function to stop the server
stop_server() {
    if [ ! -z "$SERVER_PID" ]; then
        echo "🛑 Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null
        wait $SERVER_PID 2>/dev/null
    fi
}

# Function to restart server
restart_server() {
    stop_server
    sleep 1
    run_server
}

# Trap Ctrl+C
trap 'echo ""; echo "🛑 Shutting down..."; stop_server; exit 0' INT

# Start server
run_server

# Watch for file changes (simple polling approach)
echo "👀 Watching for file changes..."
echo "💡 Save any .go, .json, or .env file to restart"
echo ""

LAST_CHANGE=0

while true; do
    # Find the most recent .go, .json, or .env file modification
    if command -v find > /dev/null; then
        CURRENT_CHANGE=$(find . -name "*.go" -o -name "*.json" -o -name "*.env" 2>/dev/null | grep -v tmp | grep -v vendor | head -20 | xargs stat -f %m 2>/dev/null | sort -nr | head -1)
        
        if [ -z "$CURRENT_CHANGE" ]; then
            CURRENT_CHANGE=0
        fi
        
        if [ "$CURRENT_CHANGE" -gt "$LAST_CHANGE" ]; then
            echo "🔄 File change detected, restarting server..."
            restart_server
            LAST_CHANGE=$CURRENT_CHANGE
        fi
    fi
    
    sleep 2
done 