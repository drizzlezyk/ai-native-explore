#!/bin/bash
# Run gitleaks to detect secrets and sensitive information

set -euo pipefail

echo "🔐 Scanning for secrets and sensitive information..."
echo ""

# Check if gitleaks is installed
if ! command -v gitleaks &> /dev/null; then
    echo "❌ gitleaks is not installed"
    echo ""
    echo "Install from: https://github.com/gitleaks/gitleaks#installing"
    echo ""
    echo "Or run:"
    echo "  bash .claude/skills/local-ci-python/scripts/install_tools.sh"
    exit 1
fi

# Check if this is a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not a git repository"
    echo "This script requires git history scanning for CI parity."
    exit 1
fi

# Check for custom gitleaks configuration
if [ -f .gitleaks.toml ]; then
    echo "Using custom configuration: .gitleaks.toml"
    CONFIG_ARG="--config .gitleaks.toml"
else
    echo "Using default gitleaks configuration"
    CONFIG_ARG=""
fi

echo ""

# Advisory-only uncommitted scan: report risk without affecting exit code.
run_uncommitted_advisory() {
    echo ""
    echo "Advisory scan: uncommitted changes"
    echo "⚠️  Findings in this section are reminders for developers and do not change script exit status"
    echo ""

    if gitleaks detect $CONFIG_ARG --no-git --verbose 2>&1 | tee gitleaks_uncommitted_output.txt; then
        echo ""
        echo "✅ No secrets detected in uncommitted changes (advisory)"
    else
        echo ""
        echo "⚠️  Uncommitted risk detected (advisory only)."
        echo "Please fix before commit; do not add uncommitted findings to .gitleaksignore."
    fi

    rm -f gitleaks_uncommitted_output.txt
}

# Determine scan mode
# Default to committed changes so local checks align with CI-impacting commits.
SCAN_MODE="${1:-committed}"

case "$SCAN_MODE" in
    staged)
        echo "Scanning staged changes (git diff --cached)..."
        echo "This checks files you're about to commit"
        echo ""
        if gitleaks protect $CONFIG_ARG --staged --verbose 2>&1 | tee gitleaks_output.txt; then
            echo ""
            echo "✅ No secrets detected in staged changes!"
            rm -f gitleaks_output.txt
            exit 0
        else
            EXIT_CODE=$?
        fi
        ;;

    uncommitted)
        echo "Scanning uncommitted changes (advisory only)..."
        echo "⚠️  This mode only reports risk and will not update .gitleaksignore"
        echo ""
        if gitleaks detect $CONFIG_ARG --no-git --verbose 2>&1 | tee gitleaks_output.txt; then
            echo ""
            echo "✅ No secrets detected in uncommitted changes!"
        else
            echo ""
            echo "⚠️  Uncommitted risk detected (advisory only)."
            echo "Please fix before commit; do not add uncommitted findings to .gitleaksignore."
        fi
        rm -f gitleaks_output.txt
        exit 0
        ;;

    committed)
        echo "Scanning committed changes only (upstream..HEAD)..."
        echo ""
        echo "Syncing git refs before committed scan (fetch --all --prune --tags)..."
        if git fetch --all --prune --tags > /dev/null 2>&1; then
            echo "✅ Git refs synced"
        else
            echo "⚠️  git fetch failed; continuing with current local refs"
        fi
        echo ""

        UPSTREAM_REF="$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)"
        if [ -n "$UPSTREAM_REF" ]; then
            LOG_RANGE="${UPSTREAM_REF}..HEAD"
            echo "Using upstream range: $LOG_RANGE"
        elif git show-ref --verify --quiet refs/remotes/origin/main; then
            LOG_RANGE="origin/main..HEAD"
            echo "Using fallback range: $LOG_RANGE"
        elif git rev-parse HEAD~1 > /dev/null 2>&1; then
            LOG_RANGE="HEAD~1..HEAD"
            echo "Using fallback range: $LOG_RANGE"
        else
            LOG_RANGE="HEAD"
            echo "Using single-commit range: $LOG_RANGE"
        fi
        echo ""

        COMMIT_COUNT="$(git rev-list --count "$LOG_RANGE" 2>/dev/null || echo 0)"
        if [ "$COMMIT_COUNT" = "0" ]; then
            echo "ℹ️  No new committed changes found in range ($LOG_RANGE)"
            echo "✅ Commit-based gitleaks scan skipped"
            exit 0
        fi

        if gitleaks detect $CONFIG_ARG --log-opts="$LOG_RANGE" --verbose 2>&1 | tee gitleaks_output.txt; then
            echo ""
            echo "✅ No secrets detected in committed changes!"
            rm -f gitleaks_output.txt
            run_uncommitted_advisory
            exit 0
        else
            EXIT_CODE=$?
        fi
        ;;

    history|all-branches)
        echo "Scanning entire git history across all branches..."
        echo "⚠️  This may take a while for large repositories"
        echo ""
        echo "Syncing git refs before history scan (fetch --all --prune --tags)..."
        if git fetch --all --prune --tags > /dev/null 2>&1; then
            echo "✅ Git refs synced"
        else
            echo "⚠️  git fetch failed; continuing with current local refs"
        fi
        echo ""
        if gitleaks detect $CONFIG_ARG --log-opts="--all" --verbose 2>&1 | tee gitleaks_output.txt; then
            echo ""
            echo "✅ No secrets detected in git history (all branches)!"
            rm -f gitleaks_output.txt
            run_uncommitted_advisory
            exit 0
        else
            EXIT_CODE=$?
        fi
        ;;

    *)
        echo "❌ Invalid scan mode: $SCAN_MODE"
        echo ""
        echo "Usage: $0 [committed|all-branches|history|staged|uncommitted]"
        echo "  committed    - Scan committed changes only (upstream..HEAD, default)"
        echo "  all-branches - Scan full git history across all branches"
        echo "  history      - Alias of all-branches"
        echo "  staged       - Scan staged changes only"
        echo "  uncommitted  - Advisory-only scan, never updates ignore and never blocks"
        exit 1
        ;;
esac

# If we get here, secrets were detected
echo ""
echo "❌ Secrets detected!"
echo ""
echo "⚠️  CRITICAL: If these secrets are real, you must:"
echo "  1. Rotate/revoke the compromised credentials IMMEDIATELY"
echo "  2. Remove secrets from code"
echo "  3. If already committed, remove from git history"
echo ""
echo "Common fixes:"
echo ""
echo "1. Use environment variables:"
echo "   # Bad"
echo "   api_key = 'sk-1234567890abcdef'"
echo ""
echo "   # Good"
echo "   api_key = os.getenv('API_KEY')"
echo ""
echo "2. Use configuration files (add to .gitignore):"
echo "   # config.yaml (in .gitignore)"
echo "   api_key: sk-1234567890abcdef"
echo ""
echo "3. Use secret management services:"
echo "   - AWS Secrets Manager"
echo "   - HashiCorp Vault"
echo "   - Azure Key Vault"
echo ""
echo "4. If false positive, add to .gitleaksignore:"
echo "   path/to/file.py:line_number"
echo ""
echo "5. If already committed, remove from history:"
echo "   git filter-branch --force --index-filter \\"
echo "     'git rm --cached --ignore-unmatch path/to/file' \\"
echo "     --prune-empty --tag-name-filter cat -- --all"
echo ""
echo "   Or use BFG Repo-Cleaner: https://rtyley.github.io/bfg-repo-cleaner/"
echo ""
echo "For detailed information, see:"
echo "  .claude/skills/local-ci-python/references/security-best-practices.md"
echo ""

rm -f gitleaks_output.txt
exit $EXIT_CODE
