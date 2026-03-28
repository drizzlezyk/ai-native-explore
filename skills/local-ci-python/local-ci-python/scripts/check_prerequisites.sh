#!/bin/bash
# Check all prerequisites for CI checks before running

echo "🔍 Checking prerequisites for local CI checks..."
echo ""

MISSING_TOOLS=()
ALL_OK=true

# Check for required tools
echo "📦 Checking installed tools:"

if command -v python3 &> /dev/null; then
    echo "  ✅ python3: $(python3 --version)"
elif command -v python &> /dev/null; then
    echo "  ✅ python: $(python --version)"
else
    echo "  ❌ python: not installed"
    MISSING_TOOLS+=("python")
    ALL_OK=false
fi

if command -v pytest &> /dev/null; then
    echo "  ✅ pytest: $(pytest --version 2>&1 | head -1)"
else
    echo "  ❌ pytest: not installed"
    MISSING_TOOLS+=("pytest")
    ALL_OK=false
fi

# Check for pytest-cov
if python3 -c "import pytest_cov" 2>/dev/null || python -c "import pytest_cov" 2>/dev/null; then
    echo "  ✅ pytest-cov: installed"
else
    echo "  ❌ pytest-cov: not installed"
    MISSING_TOOLS+=("pytest-cov")
    ALL_OK=false
fi

if command -v bandit &> /dev/null; then
    echo "  ✅ bandit: $(bandit --version 2>&1 | head -1)"
else
    echo "  ❌ bandit: not installed"
    MISSING_TOOLS+=("bandit")
    ALL_OK=false
fi

if command -v gitleaks &> /dev/null; then
    echo "  ✅ gitleaks: $(gitleaks version 2>&1)"
else
    echo "  ❌ gitleaks: not installed"
    MISSING_TOOLS+=("gitleaks")
    ALL_OK=false
fi

echo ""

# Check for optional tools
echo "📦 Checking optional tools:"

if command -v coverage &> /dev/null; then
    echo "  ✅ coverage: installed"
else
    echo "  ⚠️  coverage: not installed (optional, for detailed coverage reports)"
fi

echo ""

# Check for git repository
echo "📁 Checking project setup:"

if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "  ✅ git repository: initialized"
else
    echo "  ⚠️  git repository: not initialized (needed for incremental coverage)"
    echo "     Run: git init"
fi

# Check for Python project files
if [ -f setup.py ] || [ -f pyproject.toml ] || [ -f requirements.txt ]; then
    echo "  ✅ Python project: found (setup.py/pyproject.toml/requirements.txt)"
else
    echo "  ⚠️  Python project files: not found"
    echo "     Consider creating requirements.txt or pyproject.toml"
fi

# Check for test files
if find . -name "test_*.py" -o -path "*/tests/*" -name "*.py" 2>/dev/null | grep -q .; then
    test_count=$(find . -name "test_*.py" -o -path "*/tests/*" -name "*.py" 2>/dev/null | wc -l)
    echo "  ✅ test files: found ($test_count files)"
else
    echo "  ⚠️  test files: no test_*.py files or tests/ directory found"
fi

echo ""

# Summary and recommendations
if [ "$ALL_OK" = true ]; then
    echo "🎉 All required tools are installed!"
    echo ""
    echo "You can now run CI checks:"
    echo "  bash .claude/skills/local-ci-python/scripts/run_all_checks.sh"
    echo ""
    exit 0
else
    echo "❌ Some required tools are missing!"
    echo ""

    if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
        echo "📥 Missing tools that need to be installed:"
        for tool in "${MISSING_TOOLS[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "To install all missing tools, run:"
        echo "  bash .claude/skills/local-ci-python/scripts/install_tools.sh"
        echo ""
        echo "Or install individually:"
        for tool in "${MISSING_TOOLS[@]}"; do
            case "$tool" in
                "python")
                    echo "  - python: https://www.python.org/downloads/"
                    ;;
                "pytest")
                    echo "  - pytest: pip install pytest"
                    ;;
                "pytest-cov")
                    echo "  - pytest-cov: pip install pytest-cov"
                    ;;
                "bandit")
                    echo "  - bandit: pip install bandit"
                    ;;
                "gitleaks")
                    echo "  - gitleaks: https://github.com/gitleaks/gitleaks#installing"
                    ;;
            esac
        done
        echo ""
    fi

    exit 1
fi
