#!/bin/bash
# Install all required tools for local CI checks

set -e

echo "📥 Installing tools for local CI checks..."
echo ""

# Detect OS
OS=$(uname -s)
echo "Detected OS: $OS"
echo ""

# Check if Python is installed
if ! command -v python3 &> /dev/null && ! command -v python &> /dev/null; then
    echo "❌ Python is not installed. Please install Python first:"
    echo "   https://www.python.org/downloads/"
    exit 1
fi

# Determine Python command
if command -v python3 &> /dev/null; then
    PYTHON_CMD="python3"
    PIP_CMD="pip3"
else
    PYTHON_CMD="python"
    PIP_CMD="pip"
fi

echo "Using Python: $PYTHON_CMD ($($PYTHON_CMD --version))"
echo ""

# Install Python packages
echo "📦 Installing Python packages..."

if ! command -v pytest &> /dev/null; then
    echo "Installing pytest..."
    $PIP_CMD install pytest
fi

if ! $PYTHON_CMD -c "import pytest_cov" 2>/dev/null; then
    echo "Installing pytest-cov..."
    $PIP_CMD install pytest-cov
fi

if ! command -v bandit &> /dev/null; then
    echo "Installing bandit..."
    $PIP_CMD install bandit
fi

if ! command -v coverage &> /dev/null; then
    echo "Installing coverage (optional)..."
    $PIP_CMD install coverage
fi

echo ""

# Install Gitleaks
if ! command -v gitleaks &> /dev/null; then
    echo "📥 Installing Gitleaks..."

    if [[ "$OS" == "Linux" ]]; then
        # Linux installation
        GITLEAKS_VERSION="8.18.2"
        GITLEAKS_URL="https://github.com/gitleaks/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_linux_x64.tar.gz"

        echo "Downloading Gitleaks for Linux..."
        curl -sSfL "$GITLEAKS_URL" -o /tmp/gitleaks.tar.gz

        echo "Extracting..."
        tar -xzf /tmp/gitleaks.tar.gz -C /tmp

        echo "Installing to /usr/local/bin (may require sudo)..."
        sudo mv /tmp/gitleaks /usr/local/bin/
        sudo chmod +x /usr/local/bin/gitleaks

        rm -f /tmp/gitleaks.tar.gz

    elif [[ "$OS" == MINGW* ]] || [[ "$OS" == CYGWIN* ]] || [[ "$OS" == MSYS* ]]; then
        # Windows installation
        GITLEAKS_VERSION="8.18.2"
        GITLEAKS_URL="https://github.com/gitleaks/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_windows_x64.zip"

        echo "Downloading Gitleaks for Windows..."
        curl -sSfL "$GITLEAKS_URL" -o /tmp/gitleaks.zip

        echo "Extracting..."
        unzip -q /tmp/gitleaks.zip -d /tmp

        echo "Installing to /usr/local/bin..."
        mkdir -p /usr/local/bin
        mv /tmp/gitleaks.exe /usr/local/bin/

        rm -f /tmp/gitleaks.zip

    else
        echo "⚠️  Unsupported OS for automatic Gitleaks installation: $OS"
        echo "Please install manually from: https://github.com/gitleaks/gitleaks/releases"
    fi

    echo ""
fi

# Verify installations
echo "✅ Verifying installations..."
echo ""

ERRORS=0

if command -v pytest &> /dev/null; then
    echo "  ✅ pytest: $(pytest --version 2>&1 | head -1)"
else
    echo "  ❌ pytest: installation failed"
    ERRORS=$((ERRORS + 1))
fi

if $PYTHON_CMD -c "import pytest_cov" 2>/dev/null; then
    echo "  ✅ pytest-cov: installed"
else
    echo "  ❌ pytest-cov: installation failed"
    ERRORS=$((ERRORS + 1))
fi

if command -v bandit &> /dev/null; then
    echo "  ✅ bandit: $(bandit --version 2>&1 | head -1)"
else
    echo "  ❌ bandit: installation failed"
    ERRORS=$((ERRORS + 1))
fi

if command -v gitleaks &> /dev/null; then
    echo "  ✅ gitleaks: $(gitleaks version 2>&1)"
else
    echo "  ❌ gitleaks: installation failed"
    ERRORS=$((ERRORS + 1))
fi

echo ""

if [ $ERRORS -eq 0 ]; then
    echo "🎉 All tools installed successfully!"
    echo ""
    echo "You can now run CI checks:"
    echo "  bash .claude/skills/local-ci-python/scripts/run_all_checks.sh"
else
    echo "⚠️  $ERRORS tool(s) failed to install"
    echo "Please install them manually and try again"
    exit 1
fi
