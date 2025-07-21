#!/bin/bash

# Evolution PostgreSQL Backup - Docker Build Script
set -e

echo "🐳 Building Docker images for Evolution PostgreSQL Backup"
echo "========================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
REGISTRY="ghcr.io/your-username"
TAG="${1:-latest}"
PLATFORM="${2:-linux/amd64,linux/arm64}"
PUSH="${3:-false}"

echo -e "${BLUE}Configuration:${NC}"
echo "  Registry: $REGISTRY"
echo "  Tag: $TAG"
echo "  Platform: $PLATFORM"
echo "  Push: $PUSH"
echo ""

# Function to build image
build_image() {
    local component=$1
    local dockerfile=$2
    local context=$3
    local image_name="$REGISTRY/evolution-postgres-backup-$component:$TAG"
    
    echo -e "${YELLOW}🔨 Building $component image...${NC}"
    echo "  Image: $image_name"
    echo "  Dockerfile: $dockerfile"
    echo "  Context: $context"
    
    if [ "$PUSH" = "true" ]; then
        echo "  Action: Build and Push"
        docker buildx build \
            --platform "$PLATFORM" \
            --file "$dockerfile" \
            --tag "$image_name" \
            --push \
            "$context"
    else
        echo "  Action: Build only"
        docker buildx build \
            --platform "$PLATFORM" \
            --file "$dockerfile" \
            --tag "$image_name" \
            --load \
            "$context"
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $component image built successfully${NC}"
    else
        echo -e "${RED}❌ Failed to build $component image${NC}"
        exit 1
    fi
    echo ""
}

# Ensure buildx is set up
echo -e "${BLUE}🔧 Setting up Docker Buildx...${NC}"
docker buildx create --use --name evolution-builder 2>/dev/null || docker buildx use evolution-builder || true
docker buildx inspect --bootstrap
echo ""

# Build API image
build_image "api" "Dockerfile.api" "."

# Build Worker image  
build_image "worker" "Dockerfile.worker" "."

# Build Frontend image
build_image "frontend" "frontend/Dockerfile" "frontend"

echo -e "${GREEN}🎉 All images built successfully!${NC}"
echo ""

if [ "$PUSH" = "false" ]; then
    echo -e "${YELLOW}📋 Local images created:${NC}"
    echo "  • $REGISTRY/evolution-postgres-backup-api:$TAG"
    echo "  • $REGISTRY/evolution-postgres-backup-worker:$TAG"
    echo "  • $REGISTRY/evolution-postgres-backup-frontend:$TAG"
    echo ""
    echo -e "${BLUE}💡 To test locally:${NC}"
    echo "  docker-compose -f docker-compose.registry.yml up"
    echo ""
    echo -e "${BLUE}💡 To push to registry:${NC}"
    echo "  $0 $TAG $PLATFORM true"
else
    echo -e "${GREEN}🚀 Images pushed to registry successfully!${NC}"
    echo ""
    echo -e "${BLUE}💡 To use in production:${NC}"
    echo "  docker-compose -f docker-compose.registry.yml pull"
    echo "  docker-compose -f docker-compose.registry.yml up -d"
fi

echo ""
echo -e "${BLUE}📊 Image sizes:${NC}"
docker images | grep evolution-postgres-backup | grep "$TAG" 