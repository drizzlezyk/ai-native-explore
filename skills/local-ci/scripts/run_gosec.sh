#!/bin/bash
# Run gosec security scanner locally to match the CI environment

set -e

echo "🔒 Running gosec security scanner..."

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "❌ gosec is not installed"
    echo "Install it with: go install github.com/securego/gosec/v2/cmd/gosec@latest"
    exit 1
fi

# Run gosec on all packages
gosec ./...

echo "✅ gosec security scan passed!"
