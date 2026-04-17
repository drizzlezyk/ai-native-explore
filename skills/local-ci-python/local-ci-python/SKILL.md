---
name: local-ci-python
description: Run CI checks locally for Python projects before pushing code. Includes unit test coverage validation (10% baseline, 80% incremental) with pytest-cov, security scanning with Bandit, and sensitive information detection with Gitleaks. Use when you want to verify code quality, catch security issues, or validate test coverage before committing. Triggers on requests like "run CI checks", "check test coverage", "scan for secrets", "run security checks", or any mention of local CI validation for Python projects. Supports Linux and Windows.
---

# Local CI for Python Projects

Run comprehensive CI checks locally to catch issues before pushing code.

## Overview

This skill provides three essential CI checks for Python projects:

1. **Unit Test Coverage** - Validate test coverage meets thresholds
   - Baseline: 10% overall coverage
   - Incremental: 80% coverage for new/changed code
   - Tool: `pytest-cov`

2. **Security Scanning** - Detect security vulnerabilities in Python code
   - Tool: `bandit`
   - Checks: SQL injection, hardcoded passwords, weak crypto, command injection, etc.

3. **Sensitive Information Detection** - Prevent secrets from being committed
   - Tool: `gitleaks`
   - Detects: API keys, passwords, tokens, private keys, etc.

**Platform Support**: Linux and Windows

## Quick Start

### First Time Setup

1. **Check prerequisites**:
```bash
bash .claude/skills/local-ci-python/scripts/check_prerequisites.sh
```

2. **Install missing tools**:
```bash
bash .claude/skills/local-ci-python/scripts/install_tools.sh
```

### Running CI Checks

**Run all checks at once**:
```bash
bash .claude/skills/local-ci-python/scripts/run_all_checks.sh
```

**Or run individual checks**:
```bash
bash .claude/skills/local-ci-python/scripts/run_tests.sh      # Test coverage
bash .claude/skills/local-ci-python/scripts/run_security.sh   # Bandit scan
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh   # Secret detection (default: committed)
```

## Individual CI Checks

### 1. Unit Test Coverage

**Purpose**: Ensure adequate test coverage for code quality

**Coverage Thresholds**:
- **Baseline**: 10% overall project coverage
- **Incremental**: 80% coverage for new/changed code (git diff)

**Requirements**:
- Python installed
- pytest and pytest-cov installed
- Test files (typically in `tests/` directory or `test_*.py` files)

**Usage**:
```bash
bash .claude/skills/local-ci-python/scripts/run_tests.sh
```

**What it does**:
1. Runs all tests with coverage: `pytest --cov=. --cov-report=term --cov-report=xml`
2. Checks overall coverage meets 10% threshold
3. Analyzes git diff to identify changed code
4. Validates changed code has 80% coverage
5. Generates detailed coverage report

**Common fixes**:
- Low baseline coverage: Add more unit tests
- Low incremental coverage: Add tests for new/changed functions
- View coverage details: `coverage html` then open `htmlcov/index.html`
- See uncovered lines: `coverage report --show-missing`

### 2. Security Scanning (Bandit)

**Purpose**: Identify security vulnerabilities in Python code

**Requirements**:
- `bandit` installed: `pip install bandit`
- Or use install script

**Usage**:
```bash
bash .claude/skills/local-ci-python/scripts/run_security.sh
```

**What it checks**:
- B101: Hardcoded passwords
- B102: Shell injection via exec/eval
- B103: Setting bad file permissions
- B201/B202: Flask app debug mode
- B301-B324: Unsafe deserialization (pickle, yaml, etc.)
- B501-B507: Weak cryptographic hashing
- B601-B612: SQL injection and shell injection
- And 50+ other security issues

**Common fixes**:
- **Hardcoded passwords (B105, B106)**: Use environment variables
  ```python
  password = os.getenv('DB_PASSWORD')  # Good
  password = "secret123"  # Bad
  ```
- **SQL injection (B608)**: Use parameterized queries
  ```python
  cursor.execute("SELECT * FROM users WHERE id = %s", (user_id,))  # Good
  cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")  # Bad
  ```
- **Weak crypto (B303, B324)**: Use strong algorithms
  ```python
  hashlib.sha256()  # Good
  hashlib.md5()  # Bad
  ```
- **Shell injection (B602, B605)**: Avoid shell=True
  ```python
  subprocess.run(['ls', '-la'])  # Good
  subprocess.run('ls -la', shell=True)  # Bad
  ```

See [references/security-best-practices.md](references/security-best-practices.md) for detailed fixes.

### 3. Sensitive Information Detection (Gitleaks)

**Purpose**: Prevent secrets and credentials from being committed

**Requirements**:
- `gitleaks` installed: Download from https://github.com/gitleaks/gitleaks/releases
- Or use install script

**Usage**:
```bash
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh
```

Default behavior: `committed` mode runs first for CI parity, then an automatic `uncommitted` advisory scan is shown to remind developers about local risks (does not affect exit code and is not used to update ignore).

**What it detects**:
- API keys (AWS, Google, Azure, etc.)
- Authentication tokens (GitHub, GitLab, etc.)
- Database credentials
- Private keys (RSA, SSH, etc.)
- OAuth secrets
- Passwords in code
- JWT tokens
- And 100+ other secret patterns

**Common fixes**:
- **Remove secrets from code**: Use environment variables
  ```python
  API_KEY = os.getenv('API_KEY')  # Good
  API_KEY = "sk-1234567890abcdef"  # Bad
  ```
- **Use configuration files** (add to .gitignore)
  ```python
  # config.py (in .gitignore)
  DATABASE_URL = "postgresql://user:pass@localhost/db"
  ```
- **If false positive**: Add to `.gitleaksignore`
- **Already committed secrets**: Rotate credentials immediately and remove from git history

**Scan modes**:
```bash
# Scan committed changes only (default, CI-oriented)
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh

# Explicit committed mode
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh committed

# Scan entire git history across all branches (explicit)
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh all-branches

# Alias of all-branches
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh history

# Scan staged changes only
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh staged

# Scan uncommitted changes (advisory only; never updates ignore and never blocks)
bash .claude/skills/local-ci-python/scripts/run_gitleaks.sh uncommitted
```

Note: For `committed` and `all-branches`/`history`, the script automatically runs `git fetch --all --prune --tags` before scanning to reduce local-vs-CI mismatch caused by stale refs.

## Configuration

### Test Coverage Thresholds

Edit thresholds in `scripts/run_tests.sh`:
```bash
BASELINE_COVERAGE=10    # Overall project coverage (%)
INCREMENTAL_COVERAGE=80 # New/changed code coverage (%)
```

### Bandit Configuration

Create `.bandit` in project root to customize:
```yaml
# .bandit
skips: ['B101']  # Skip specific tests
exclude_dirs: ['/tests/', '/venv/']
severity_level: medium
confidence_level: medium
```

### Gitleaks Configuration

Create `.gitleaks.toml` in project root to customize:
```toml
[extend]
useDefault = true

[allowlist]
paths = [
  ".*test.*",
  ".*mock.*"
]
```

## Workflow: Fixing CI Failures

### Test Coverage Failures

1. Run coverage check
2. View coverage report: `coverage html && open htmlcov/index.html`
3. Identify uncovered code: `coverage report --show-missing`
4. Add tests for uncovered functions
5. Re-run to verify

### Security Scan Failures

1. Run security scan
2. Read error details (file, line, issue type)
3. Consult references/security-best-practices.md
4. Apply fix based on issue type
5. Re-run to verify

### Secret Detection Failures

1. Run gitleaks scan
2. Identify the secret (file and line)
3. Remove secret (move to env var or config file)
4. If false positive, update `.gitleaksignore` with compatibility entries:
   - Required format: `commit:file:rule:line`
   - Use CI-reported fingerprint directly to avoid mismatch
   - Only add committed/history findings; for `uncommitted` mode, only fix code and do not update ignore
   - Get commit SHA with: `git rev-parse HEAD`
5. If already committed: Rotate credential immediately
6. Re-run to verify

Example `.gitleaksignore` entries for one finding:
```text
10616a3c20cadc1f6fe3762f26850cc2f8971bd9:path/to/file.py:generic-api-key:42
```

## Integration with GitHub Actions

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on: [push, pull_request]

jobs:
  test-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          pip install pytest pytest-cov
          pip install -e .

      - name: Run tests with coverage
        run: |
          pytest --cov=. --cov-report=xml --cov-report=term
          COVERAGE=$(python -c "import xml.etree.ElementTree as ET; tree = ET.parse('coverage.xml'); root = tree.getroot(); print(float(root.attrib['line-rate']) * 100)")
          if (( $(echo "$COVERAGE < 10" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 10%"
            exit 1
          fi

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Run Bandit
        run: |
          pip install bandit
          bandit -r . -ll

  secret-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Troubleshooting

**Script not found**:
- Ensure you're in project root
- Check: `ls .claude/skills/local-ci-python/scripts/`

**Permission denied**:
- Make executable: `chmod +x .claude/skills/local-ci-python/scripts/*.sh`

**Tool not installed**:
- Run: `bash .claude/skills/local-ci-python/scripts/install_tools.sh`

**Coverage calculation fails**:
- Ensure tests exist: `find . -name "test_*.py"` or `ls tests/`
- Check tests pass: `pytest`

**Incremental coverage fails**:
- Ensure git repo: `git status`
- Commit changes first

**Bandit false positives**:
- Add exclusions to `.bandit`
- Use `# nosec` comment (use sparingly)

**Gitleaks false positives**:
- Add to `.gitleaksignore`
- Customize `.gitleaks.toml`
- Preferred ignore line format: `commit:file:rule:line`
- Do not add `uncommitted` findings to ignore; treat them as advisory prompts to fix code before commit

**pytest-cov not found**:
- Install: `pip install pytest-cov`
- Or add to requirements.txt

## Best Practices

1. **Run checks before every commit**
2. **Use git hooks** for automation
3. **Focus on incremental coverage** - aim for 80%+ on new code
4. **Never commit secrets** - use environment variables
5. **Fix security issues immediately** - don't ignore bandit warnings
6. **Review coverage reports** - understand what's not tested
7. **Keep tools updated**

## Resources

- [Bandit Documentation](https://bandit.readthedocs.io/)
- [Gitleaks Documentation](https://github.com/gitleaks/gitleaks)
- [pytest-cov Documentation](https://pytest-cov.readthedocs.io/)
- [Python Testing Best Practices](https://docs.python-guide.org/writing/tests/)
