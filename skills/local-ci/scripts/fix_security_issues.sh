#!/bin/bash
# Fix security issues using dual-agent verification approach

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GOSEC_REPORT="$PROJECT_ROOT/.ci-temp/gosec-report.json"
GOSEC_BASELINE="$PROJECT_ROOT/.ci-temp/gosec-baseline.json"
FIX_REPORT="$PROJECT_ROOT/.ci-temp/security-fixes.md"

echo "=================================="
echo "🔐 Security Issue Auto-Fix"
echo "=================================="
echo ""

# Check if gosec report exists
if [ ! -f "$GOSEC_REPORT" ]; then
    echo "❌ No gosec report found. Run gosec scan first:"
    echo "   bash .claude/skills/local-ci/scripts/run_gosec.sh"
    exit 1
fi

# Count issues
ISSUE_COUNT=$(jq -r '.Issues | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")

if [ "$ISSUE_COUNT" -eq 0 ]; then
    echo "✅ No security issues to fix!"
    exit 0
fi

echo "📊 Found $ISSUE_COUNT security issue(s)"
echo ""

# Save current git state
CURRENT_BRANCH=$(git branch --show-current)
BEFORE_COMMIT=$(git rev-parse HEAD)

echo "📸 Saving current state..."
echo "   Branch: $CURRENT_BRANCH"
echo "   Commit: $BEFORE_COMMIT"
echo ""

# Launch Fixer Agent
echo "=================================="
echo "🤖 Phase 1: Launching Fixer Agent"
echo "=================================="
echo ""

cat > /tmp/fixer_agent_prompt.txt <<'FIXER_EOF'
You are a Security Fixer Agent. Your task is to fix security vulnerabilities found by gosec.

**Evidence Sources (read these directly):**
1. Original scan report: .ci-temp/gosec-report.json
2. Baseline report: .ci-temp/gosec-baseline.json (if exists)

**Your Process:**
1. Read the gosec-report.json to understand all security issues
2. For each issue:
   - Read the vulnerable code file
   - Understand the security risk
   - Apply the appropriate fix
   - Commit the fix with a clear message

3. Create a fix summary at .ci-temp/security-fixes.md with:
   - List of all issues found
   - Fix applied for each issue
   - Files modified
   - Commit SHAs

**Common Security Fixes:**
- G101 (credentials): Move to environment variables or secure storage
- G104 (unhandled errors): Add proper error handling
- G201 (SQL injection): Use parameterized queries
- G304 (file traversal): Use filepath.Clean() and validate paths
- G401 (weak crypto): Use SHA256 instead of MD5
- G402 (TLS config): Set MinVersion to TLS 1.2
- G404 (weak random): Use crypto/rand instead of math/rand

**Important:**
- Make atomic commits (one fix per commit)
- Use descriptive commit messages
- Test that code still compiles after each fix
- Do NOT skip any issues

After completing all fixes, output:
- Total issues fixed
- List of commits created
- Summary saved to .ci-temp/security-fixes.md

Wait for Verifier Agent to validate your fixes.
FIXER_EOF

# Check if we should use Claude Code Task tool to dispatch agent
echo "🔧 Dispatching Fixer Agent to analyze and fix security issues..."
echo ""
echo "📋 Fixer Agent will:"
echo "   - Read gosec-report.json directly"
echo "   - Analyze each security issue"
echo "   - Apply fixes with atomic commits"
echo "   - Generate fix summary"
echo ""

# Here the user/main agent would invoke the Task tool to dispatch the fixer agent
# For now, output instructions
echo "⚠️  Manual step required:"
echo ""
echo "Please dispatch a Fixer Agent with the following prompt:"
echo ""
cat /tmp/fixer_agent_prompt.txt
echo ""
echo "=================================="
echo ""
read -p "Press ENTER after Fixer Agent completes its work..."

# Check if fixes were applied
if [ ! -f "$FIX_REPORT" ]; then
    echo "❌ Fixer Agent did not generate fix report!"
    echo "Expected: $FIX_REPORT"
    exit 1
fi

AFTER_COMMIT=$(git rev-parse HEAD)

if [ "$BEFORE_COMMIT" = "$AFTER_COMMIT" ]; then
    echo "⚠️  No commits created. Either no fixes were possible or Agent failed."
    exit 1
fi

# Calculate changes
COMMITS_CREATED=$(git rev-list --count "$BEFORE_COMMIT..$AFTER_COMMIT")
echo "✅ Fixer Agent completed!"
echo "   Commits created: $COMMITS_CREATED"
echo ""

# Launch Verifier Agent
echo "=================================="
echo "🔍 Phase 2: Launching Verifier Agent"
echo "=================================="
echo ""

# Re-run gosec to get new results
NEW_GOSEC_REPORT="$PROJECT_ROOT/.ci-temp/gosec-report-after-fix.json"
echo "🔍 Running security scan on fixed code..."
gosec -fmt=json -out="$NEW_GOSEC_REPORT" ./... 2>&1 || true

cat > /tmp/verifier_agent_prompt.txt <<'VERIFIER_EOF'
You are a Security Verifier Agent. Your task is to independently verify that security fixes were properly applied.

**Evidence Sources (read these directly, do NOT rely on Fixer's report):**
1. Baseline scan: .ci-temp/gosec-baseline.json (or gosec-report.json if baseline doesn't exist)
2. After-fix scan: .ci-temp/gosec-report-after-fix.json
3. Git diff: Run `git diff <before-commit> <after-commit>` to see all changes
4. Fix summary: .ci-temp/security-fixes.md (for reference only, verify independently)

**Your Process:**
1. Read the baseline scan report to identify original issues
2. Read the after-fix scan report to check remaining issues
3. Compare the two reports:
   - Count: How many issues resolved?
   - Verify: Are the claimed fixes actually present in the diff?
   - Validate: Do remaining issues need further attention?

4. For each claimed fix:
   - Read the actual code changes in git diff
   - Verify the fix addresses the security issue correctly
   - Check for any new issues introduced by the fix

5. Re-run lightweight verification:
   - Compile the code: `go build ./...`
   - Run affected tests if any
   - Confirm no regressions

**Output Format:**
Create .ci-temp/verification-report.md with:

```markdown
# Security Fix Verification Report

## Summary
- Original issues: X
- Issues resolved: Y
- Issues remaining: Z
- New issues introduced: N
- Verdict: PASS / FAIL / PARTIAL

## Detailed Verification

### ✅ Successfully Fixed (Y issues)
- [G101] Issue in file.go:line - Fixed by: [commit SHA] - Verified: [how you verified]

### ❌ Not Fixed (Z issues)
- [G304] Issue in file.go:line - Reason: [why fix didn't work]

### ⚠️ New Issues (N issues)
- [G402] Issue in file.go:line - Introduced by: [commit SHA]

## Evidence
- Baseline scan: X issues
- After-fix scan: Y issues
- Git commits reviewed: [list SHAs]
- Code changes verified: [summary of changes]

## Conclusion
[PASS/FAIL/PARTIAL] - [Detailed explanation]

## Recommendations
[Any follow-up actions needed]
```

**Verification Criteria:**
- PASS: All critical/high issues resolved, no new issues
- PARTIAL: Some issues resolved, but some remain or new issues introduced
- FAIL: Fixes didn't work or made things worse

Be objective and thorough. Your verification is independent.
VERIFIER_EOF

echo "🔍 Dispatching Verifier Agent for independent verification..."
echo ""
echo "📋 Verifier Agent will:"
echo "   - Read baseline and after-fix scan reports directly"
echo "   - Compare issue counts"
echo "   - Review git diff independently"
echo "   - Re-run verification tests"
echo "   - Provide PASS/FAIL verdict with evidence"
echo ""
echo "⚠️  Manual step required:"
echo ""
echo "Please dispatch a Verifier Agent with the following prompt:"
echo ""
echo "Context to provide:"
echo "  - before_commit: $BEFORE_COMMIT"
echo "  - after_commit: $AFTER_COMMIT"
echo ""
cat /tmp/verifier_agent_prompt.txt
echo ""
echo "=================================="
echo ""
read -p "Press ENTER after Verifier Agent completes verification..."

# Check verification report
VERIFICATION_REPORT="$PROJECT_ROOT/.ci-temp/verification-report.md"
if [ ! -f "$VERIFICATION_REPORT" ]; then
    echo "❌ Verifier Agent did not generate verification report!"
    echo "Expected: $VERIFICATION_REPORT"
    exit 1
fi

echo ""
echo "📊 Verification Results:"
echo ""
cat "$VERIFICATION_REPORT"
echo ""

# Extract verdict
VERDICT=$(grep -E "^##? (Summary|Verdict)" "$VERIFICATION_REPORT" -A 10 | grep -E "Verdict:|PASS|FAIL|PARTIAL" | head -1 || echo "UNKNOWN")

echo "=================================="
echo "🏁 Final Result"
echo "=================================="
echo ""

if echo "$VERDICT" | grep -q "PASS"; then
    echo "✅ VERIFICATION PASSED!"
    echo ""
    echo "All security issues have been successfully fixed and verified."
    echo "The code is ready for commit."
    exit 0
elif echo "$VERDICT" | grep -q "PARTIAL"; then
    echo "⚠️  VERIFICATION PARTIAL"
    echo ""
    echo "Some issues were fixed, but work remains."
    echo "Review the verification report for details."
    exit 1
else
    echo "❌ VERIFICATION FAILED"
    echo ""
    echo "Security fixes were not successful."
    echo "Review the verification report and try manual fixes."
    exit 1
fi
