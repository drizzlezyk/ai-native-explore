#!/bin/bash
# Run typos spell checker locally to match the CI environment

set -e

echo "📝 Running typos spell checker..."

# Check if typos is installed
if ! command -v typos &> /dev/null; then
    echo "❌ typos is not installed"
    echo "Install it with: cargo install typos-cli"
    echo "Or download from: https://github.com/crate-ci/typos/releases"
    exit 1
fi

# Check if typos.toml exists
if [ ! -f typos.toml ]; then
    echo "⚠️  Warning: typos.toml not found, using default configuration"
fi

# Run typos check (non-fixing mode first to report issues)
echo "Checking for typos..."
if ! typos ./; then
    echo ""
    echo "❌ Typos found! To fix them automatically, run:"
    echo "   typos --write-changes ./"
    exit 1
fi

echo "✅ No typos found!"
