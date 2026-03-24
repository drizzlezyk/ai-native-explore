---
name: local-ci
description: Run GitHub Actions CI checks locally before pushing code. Covers all CI workflows including golangci-lint, typos spell check, gosec security scan, unit tests with coverage, and Docker builds. Use when the user wants to run CI checks locally, verify code before PR, debug CI failures, or fix linting/test/security issues. Triggers on requests like "run CI locally", "check CI", "run lint", "run tests", "fix CI errors", or any mention of running or debugging GitHub Actions workflows locally. Supports Linux and macOS.
---

# Local CI

Run all GitHub Actions CI checks locally to catch issues before pushing code.

## Overview

This skill enables running the complete CI pipeline locally, matching the GitHub Actions environment. It covers all workflows defined in `.github/workflows/`:

- **golangci-lint** - Static code analysis
- **typos** - Spell checking
- **gosec** - Security vulnerability scanning
- **unit tests** - Test execution with coverage validation
- **Docker build** - Image build verification

**Platform Support**: Linux and macOS

## Quick Start

### First Time Setup

1. **Check prerequisites** (checks all tools and environment):
```bash
bash .claude/local-ci/scripts/check_prerequisites.sh
```

2. **Install missing tools** (if any are missing):
```bash
bash .claude/local-ci/scripts/install_tools.sh
```

3. **Set environment variables** (optional, for Docker build with private repos):
```bash
export GITHUB_USER=your-username
export GITHUB_TOKEN=your-token
```

### Running CI Checks

**Run all CI checks at once** (automatically checks prerequisites first):
```bash
bash .claude/local-ci/scripts/run_all_checks.sh
```

**Or run individual checks**:
```bash
bash .claude/local-ci/scripts/run_lint.sh          # golangci-lint
bash .claude/local-ci/scripts/run_typos.sh         # spell check
bash .claude/local-ci/scripts/run_gosec.sh         # security scan
bash .claude/local-ci/scripts/run_tests.sh         # unit tests
bash .claude/local-ci/scripts/run_docker_build.sh  # docker build
```

## Individual CI Checks

### 1. golangci-lint (Code Quality)

**Purpose**: Static analysis for Go code (unused vars, formatting, error handling, etc.)

**Requirements**:
- `golangci-lint` installed: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin`
- Or use install script: `.claude/local-ci/scripts/install_tools.sh`
- `.golangci.yml` in project root

**Usage**:
```bash
bash .claude/local-ci/scripts/run_lint.sh
```

**Common fixes**:
- Unused variables: Remove or prefix with `_`
- Formatting: Run `gofmt -w .`
- Unchecked errors: Add proper error handling
- See [references/ci-workflows.md](references/ci-workflows.md#1-golangci-lint) for detailed fixes

### 2. Typos (Spell Checking)

**Purpose**: Detect typos in code, comments, and documentation

**Requirements**:
- `typos` installed: `cargo install typos-cli` or download from https://github.com/crate-ci/typos/releases
- Or use install script: `.claude/local-ci/scripts/install_tools.sh`
- `typos.toml` configuration (optional)

**Usage**:
```bash
bash .claude/local-ci/scripts/run_typos.sh
```

**Common fixes**:
- Auto-fix typos: `typos --write-changes ./`
- Add false positives to `typos.toml`:
  ```toml
  [default.extend-words]
  word = "word"  # Marks as intentional
  ```

### 3. Gosec (Security Scanning)

**Purpose**: Identify security vulnerabilities in Go code

**Requirements**:
- `gosec` installed: `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- Or use install script: `.claude/local-ci/scripts/install_tools.sh`

**Usage**:
```bash
bash .claude/local-ci/scripts/run_gosec.sh
```

**Common fixes**:
- Weak crypto (G401): Use SHA256 instead of MD5
- SQL injection (G201): Use parameterized queries
- File traversal (G304): Validate paths with `filepath.Clean()`
- See [references/ci-workflows.md](references/ci-workflows.md#3-gosec) for detailed fixes

### 4. Unit Tests (Coverage Check)

**Purpose**: Run tests and validate coverage meets threshold (0.2% minimum)

**Requirements**:
- `tests/test_coverage_get.sh` script exists
- `.github/workflows/.testcoverage.yml` configuration
- Optional: `go install github.com/vladopajic/go-test-coverage/v2@latest`

**Usage**:
```bash
bash .claude/local-ci/scripts/run_tests.sh
```

**Common fixes**:
- Test failures: Run specific test with `go test -v -run TestName`
- Low coverage: Add more unit tests
- Check `cover.out` for uncovered lines: `go tool cover -html=cover.out`

### 5. Docker Build

**Purpose**: Verify Docker image builds successfully

**Requirements**:
- Docker installed and running
- `Dockerfile` in project root
- Optional: `GITHUB_USER` and `GITHUB_TOKEN` env vars for private repo access

**Usage**:
```bash
# Without credentials
bash .claude/local-ci/scripts/run_docker_build.sh

# With credentials for private repos
export GITHUB_USER=your-username
export GITHUB_TOKEN=your-token
bash .claude/local-ci/scripts/run_docker_build.sh
```

**Common fixes**:
- Build errors: Check Dockerfile syntax
- Private repo access: Set GITHUB_USER and GITHUB_TOKEN
- Dependency issues: Update go.mod

## Prerequisites Management

### Check All Prerequisites

Before running any CI checks, verify all tools are installed and configured:

```bash
bash .claude/local-ci/scripts/check_prerequisites.sh
```

This checks:
- **Required tools**: golangci-lint, typos, gosec, docker, go
- **Optional tools**: gocov, go-test-coverage (for better test reports)
- **Environment variables**: GITHUB_USER, GITHUB_TOKEN (for Docker builds)
- **Configuration files**: .golangci.yml, Dockerfile, test scripts

### Install All Tools

If prerequisites are missing, install them all at once:

```bash
bash .claude/local-ci/scripts/install_tools.sh
```

This installs:
- golangci-lint (latest version)
- gosec (latest version)
- typos (via cargo or binary download)
- gocov and go-test-coverage (for test coverage)

**Note**: The `run_all_checks.sh` script automatically runs prerequisite checks first, so you don't need to run them manually when using the full suite.

## Workflow: Fixing CI Failures

When CI checks fail, follow this workflow:

1. **Run the failing check locally**
   ```bash
   bash .claude/local-ci/scripts/run_lint.sh  # or the specific failing check
   ```

2. **Read the error message carefully**
   - Note the file, line number, and error type
   - Errors include the specific issue and often suggest fixes

3. **Consult the references for common fixes**
   ```
   Read references/ci-workflows.md section for the failing check
   ```

4. **Apply the fix**
   - Edit the problematic code
   - For typos: Can auto-fix with `typos --write-changes`
   - For lint: Often need to refactor or add error handling
   - For tests: Debug with `go test -v -run TestName`

5. **Re-run the check to verify**
   ```bash
   bash .claude/local-ci/scripts/run_lint.sh
   ```

6. **Run all checks before pushing**
   ```bash
   bash .claude/local-ci/scripts/run_all_checks.sh
   ```

## Interpreting Errors

All scripts provide clear error messages and exit codes:
- ✅ Success: Green checkmark, exit code 0
- ❌ Failure: Red X with error details, non-zero exit code

The `run_all_checks.sh` script shows a summary of all checks and which ones passed/failed.

## Integration with GitHub Actions

The local scripts mirror these GitHub Actions workflows:
- `.github/workflows/lint.yml` → `run_lint.sh`
- `.github/workflows/typos.yml` → `run_typos.sh`
- `.github/workflows/ci.yml` → `run_gosec.sh` + `run_docker_build.sh`
- `.github/workflows/unit_test.yml` → `run_tests.sh`

Run checks locally to catch issues before they appear in GitHub Actions.

## Troubleshooting

**Script not found errors**:
- Ensure you're in the project root directory
- Check script exists: `ls .claude/local-ci/scripts/run_*.sh`

**Permission denied**:
- Make scripts executable: `chmod +x .claude/local-ci/scripts/*.sh`

**Tool not installed**:
- Follow installation instructions in error message
- Or run: `bash .claude/local-ci/scripts/install_tools.sh`

**Environment differences**:
- GitHub Actions uses `self-hosted, Linux` runners
- Ensure Go version matches (check `go version`)
- Private repos need GITHUB_USER and GITHUB_TOKEN set

For detailed error-specific fixes, see [references/ci-workflows.md](references/ci-workflows.md).
