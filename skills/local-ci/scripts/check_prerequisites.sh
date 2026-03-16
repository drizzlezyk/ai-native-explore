#!/bin/bash
# Check all prerequisites for CI checks before running

echo "🔍 Checking prerequisites for local CI checks..."
echo ""

MISSING_TOOLS=()
MISSING_ENV_VARS=()
ALL_OK=true

# Check for required tools
echo "📦 Checking installed tools:"

if command -v golangci-lint &> /dev/null; then
    echo "  ✅ golangci-lint: $(golangci-lint version 2>&1 | head -1)"
else
    echo "  ❌ golangci-lint: not installed"
    MISSING_TOOLS+=("golangci-lint")
    ALL_OK=false
fi

if command -v typos &> /dev/null; then
    echo "  ✅ typos: $(typos --version 2>&1)"
else
    echo "  ❌ typos: not installed"
    MISSING_TOOLS+=("typos")
    ALL_OK=false
fi

if command -v gosec &> /dev/null; then
    echo "  ✅ gosec: $(gosec --version 2>&1 | head -1)"
else
    echo "  ❌ gosec: not installed"
    MISSING_TOOLS+=("gosec")
    ALL_OK=false
fi

if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        echo "  ✅ docker: $(docker --version)"
    else
        echo "  ⚠️  docker: installed but daemon not running"
        MISSING_TOOLS+=("docker-daemon")
        ALL_OK=false
    fi
else
    echo "  ❌ docker: not installed"
    MISSING_TOOLS+=("docker")
    ALL_OK=false
fi

if command -v go &> /dev/null; then
    echo "  ✅ go: $(go version)"
else
    echo "  ❌ go: not installed"
    MISSING_TOOLS+=("go")
    ALL_OK=false
fi

echo ""

# Check for optional tools (for better test coverage reporting)
echo "📦 Checking optional tools:"

if command -v gocov &> /dev/null; then
    echo "  ✅ gocov: installed"
else
    echo "  ⚠️  gocov: not installed (optional, for better test coverage reports)"
fi

if command -v go-test-coverage &> /dev/null; then
    echo "  ✅ go-test-coverage: installed"
else
    echo "  ⚠️  go-test-coverage: not installed (optional, for coverage validation)"
fi

echo ""

# Check for environment variables (for Docker build with private repos)
echo "🔐 Checking environment variables:"

if [ -n "$GITHUB_USER" ]; then
    echo "  ✅ GITHUB_USER: set"
else
    echo "  ⚠️  GITHUB_USER: not set (needed for Docker build with private repos)"
    MISSING_ENV_VARS+=("GITHUB_USER")
fi

if [ -n "$GITHUB_TOKEN" ]; then
    echo "  ✅ GITHUB_TOKEN: set"
else
    echo "  ⚠️  GITHUB_TOKEN: not set (needed for Docker build with private repos)"
    MISSING_ENV_VARS+=("GITHUB_TOKEN")
fi

echo ""

# Check for configuration files
echo "📄 Checking configuration files:"

if [ -f .golangci.yml ]; then
    echo "  ✅ .golangci.yml: found"
else
    echo "  ❌ .golangci.yml: not found"
    ALL_OK=false
fi

if [ -f typos.toml ]; then
    echo "  ✅ typos.toml: found"
else
    echo "  ⚠️  typos.toml: not found (optional)"
fi

if [ -f Dockerfile ]; then
    echo "  ✅ Dockerfile: found"
else
    echo "  ❌ Dockerfile: not found"
    ALL_OK=false
fi

if [ -d tests ] && [ -f tests/test_coverage_get.sh ]; then
    echo "  ✅ tests/test_coverage_get.sh: found"
else
    echo "  ❌ tests/test_coverage_get.sh: not found"
    ALL_OK=false
fi

echo ""

# Summary and recommendations
if [ "$ALL_OK" = true ]; then
    if [ ${#MISSING_ENV_VARS[@]} -eq 0 ]; then
        echo "🎉 All prerequisites are met!"
    else
        echo "✅ All required tools are installed!"
        echo ""
        echo "⚠️  Optional: Environment variables not set (only needed for Docker build with private repos):"
        for var in "${MISSING_ENV_VARS[@]}"; do
            echo "  - $var"
        done
        echo ""
        echo "To enable Docker build with private repos, set:"
        echo "  export GITHUB_USER=your-username"
        echo "  export GITHUB_TOKEN=your-token"
    fi
    echo ""
    exit 0
else
    echo "❌ Some required prerequisites are missing!"
    echo ""

    if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
        echo "📥 Missing tools that need to be installed:"
        for tool in "${MISSING_TOOLS[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "To install all missing tools, run:"
        echo "  bash .claude/local-ci/scripts/install_tools.sh"
        echo ""
    fi

    if [ ${#MISSING_ENV_VARS[@]} -gt 0 ]; then
        echo "ℹ️  Note: GITHUB_USER/GITHUB_TOKEN are optional (only for Docker build with private repos)"
        echo ""
    fi

    exit 1
fi
