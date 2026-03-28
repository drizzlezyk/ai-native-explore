#!/bin/bash
# Run all CI checks: coverage, security, and secrets

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Running all CI checks..."
echo ""

# Track failures
FAILURES=()

# 1. Run test coverage check
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1/3: Unit Test Coverage"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if bash "$SCRIPT_DIR/run_tests.sh"; then
    echo ""
    echo "✅ Test coverage check passed"
else
    FAILURES+=("test coverage")
    echo ""
    echo "❌ Test coverage check failed"
fi

echo ""
echo ""

# 2. Run security scan
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2/3: Security Scan (Bandit)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if bash "$SCRIPT_DIR/run_security.sh"; then
    echo ""
    echo "✅ Security scan passed"
else
    FAILURES+=("security scan")
    echo ""
    echo "❌ Security scan failed"
fi

echo ""
echo ""

# 3. Run secret detection
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3/3: Secret Detection (Gitleaks)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if bash "$SCRIPT_DIR/run_gitleaks.sh" staged; then
    echo ""
    echo "✅ Secret detection passed"
else
    FAILURES+=("secret detection")
    echo ""
    echo "❌ Secret detection failed"
fi

echo ""
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ ${#FAILURES[@]} -eq 0 ]; then
    echo "🎉 All CI checks passed!"
    echo ""
    echo "✅ Test coverage"
    echo "✅ Security scan"
    echo "✅ Secret detection"
    echo ""
    echo "Your code is ready to commit!"
    exit 0
else
    echo "❌ ${#FAILURES[@]} check(s) failed:"
    echo ""
    for check in "${FAILURES[@]}"; do
        echo "  - $check"
    done
    echo ""
    echo "Please fix the issues above before committing"
    exit 1
fi
