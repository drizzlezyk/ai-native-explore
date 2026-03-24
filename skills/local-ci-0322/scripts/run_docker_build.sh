#!/bin/bash
# Build Docker image locally to match the CI environment

set -e

echo "🐳 Building Docker image..."

# Check if Dockerfile exists
if [ ! -f Dockerfile ]; then
    echo "❌ Dockerfile not found in project root"
    exit 1
fi

# Check if docker is installed and running
if ! command -v docker &> /dev/null; then
    echo "❌ docker is not installed"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ docker daemon is not running"
    exit 1
fi

# Get credentials from environment
if [ -z "$GITHUB_USER" ] || [ -z "$GITHUB_TOKEN" ]; then
    echo "⚠️  GITHUB_USER and/or GITHUB_TOKEN not set"
    echo "Attempting build without credentials..."
    echo "If build fails, set environment variables and try again:"
    echo "  export GITHUB_USER=your-username"
    echo "  export GITHUB_TOKEN=your-token"
    echo ""
    USER_ARG=""
    PASS_ARG=""
else
    echo "✅ Using GITHUB credentials from environment"
    USER_ARG="--build-arg USER=$GITHUB_USER"
    PASS_ARG="--build-arg PASS=$GITHUB_TOKEN"
fi

# Build the Docker image
TAG="server:local-$(date +%s)"
echo "Building image: $TAG"

if [ -n "$USER_ARG" ] && [ -n "$PASS_ARG" ]; then
    docker build --file Dockerfile \
        $USER_ARG \
        $PASS_ARG \
        --tag "$TAG" .
else
    docker build --file Dockerfile --tag "$TAG" .
fi

echo "✅ Docker image built successfully: $TAG"
