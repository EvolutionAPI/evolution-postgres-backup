#!/bin/bash

# Script para instalar ou atualizar o Air
echo "🔧 Installing/updating Air for live reload..."

# Remove old air if exists
if command -v air > /dev/null; then
    echo "🗑️  Removing old Air installation..."
    rm -f $(which air) 2>/dev/null || true
fi

# Install new air from correct repository
echo "📦 Installing Air from github.com/air-verse/air..."
go install github.com/air-verse/air@latest

if command -v air > /dev/null; then
    echo "✅ Air installed successfully!"
    air -v
else
    echo "❌ Failed to install Air"
    exit 1
fi 