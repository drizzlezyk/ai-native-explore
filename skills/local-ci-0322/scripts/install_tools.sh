#!/bin/bash
# Install all required tools for local CI checks

set -e

echo "📥 Installing CI tools..."
echo ""

GOBIN=$(go env GOPATH)/bin
mkdir -p "$GOBIN"

# Install golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" latest
    echo "✅ golangci-lint installed"
else
    echo "✅ golangci-lint already installed"
fi

echo ""

# Install gosec
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo "✅ gosec installed"
else
    echo "✅ gosec already installed"
fi

echo ""

# Install typos
if ! command -v typos &> /dev/null; then
    echo "Installing typos..."

    # Try cargo first
    if command -v cargo &> /dev/null; then
        cargo install typos-cli
        echo "✅ typos installed via cargo"
    else
        # Fallback to binary download
        echo "Cargo not found, downloading pre-built binary..."
        TYPOS_VERSION="1.32.0"
        curl -sSL "https://github.com/crate-ci/typos/releases/download/v${TYPOS_VERSION}/typos-v${TYPOS_VERSION}-x86_64-unknown-linux-musl.tar.gz" | tar xz -C /tmp
        mv /tmp/typos "$GOBIN/"
        chmod +x "$GOBIN/typos"
        echo "✅ typos installed via binary download"
    fi
else
    echo "✅ typos already installed"
fi

echo ""

# Install optional tools for better test coverage
echo "Installing optional tools for test coverage..."

if ! command -v gocov &> /dev/null; then
    echo "Installing gocov..."
    go install github.com/axw/gocov/gocov@latest
    echo "✅ gocov installed"
else
    echo "✅ gocov already installed"
fi

if ! command -v go-test-coverage &> /dev/null; then
    echo "Installing go-test-coverage..."
    go install github.com/vladopajic/go-test-coverage/v2@latest
    echo "✅ go-test-coverage installed"
else
    echo "✅ go-test-coverage already installed"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 All tools installed successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Verify installation
echo "Verifying installation:"
command -v golangci-lint &> /dev/null && echo "  ✅ golangci-lint: $(golangci-lint version 2>&1 | head -1)"
command -v gosec &> /dev/null && echo "  ✅ gosec: $(gosec --version 2>&1 | head -1)"
command -v typos &> /dev/null && echo "  ✅ typos: $(typos --version)"
command -v gocov &> /dev/null && echo "  ✅ gocov: installed"
command -v go-test-coverage &> /dev/null && echo "  ✅ go-test-coverage: installed"

echo ""
echo "Note: Docker must be installed separately if not already present."
echo "Visit: https://docs.docker.com/engine/install/"
echo ""
echo "For Docker builds with private repos, set these environment variables:"
echo "  export GITHUB_USER=your-username"
echo "  export GITHUB_TOKEN=your-token"
echo ""
