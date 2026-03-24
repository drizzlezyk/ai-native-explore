# CI Workflows Reference

This document provides detailed information about each CI workflow and common issues with fixes.

## 1. golangci-lint

**Purpose**: Static code analysis for Go code
**Configuration**: `.golangci.yml` in project root
**Script**: `scripts/run_lint.sh`

### Common Issues and Fixes

#### Unused variables/imports
```
Error: variable 'foo' is unused (unused)
```
**Fix**: Remove the unused variable or use it, or prefix with `_` if intentionally unused

#### Formatting issues
```
Error: File is not gofmt-ed (gofmt)
```
**Fix**: Run `gofmt -w .` or `goimports -w .`

#### Ineffective assignments
```
Error: ineffectual assignment to variable (ineffassign)
```
**Fix**: Remove the ineffective assignment or use the variable

#### Error handling
```
Error: Error return value is not checked (errcheck)
```
**Fix**: Add proper error checking:
```go
if err != nil {
    return err
}
```

## 2. typos

**Purpose**: Spell checking across all files
**Configuration**: `typos.toml` in project root
**Script**: `scripts/run_typos.sh`

### Common Issues and Fixes

#### Typos in code/comments
```
Error: `teh` should be `the`
```
**Fix Option 1**: Fix the typo manually
**Fix Option 2**: Run `typos --write-changes ./` to auto-fix
**Fix Option 3**: Add to `typos.toml` if it's a false positive:
```toml
[default.extend-words]
teh = "teh"  # If this is intentional (e.g., variable name)
```

## 3. gosec

**Purpose**: Security vulnerability scanning
**Script**: `scripts/run_gosec.sh`

### Common Issues and Fixes

#### Weak crypto
```
Error: Use of weak cryptographic primitive (G401)
```
**Fix**: Replace with stronger algorithm (e.g., SHA256 instead of MD5)

#### SQL injection
```
Error: SQL string formatting (G201)
```
**Fix**: Use parameterized queries:
```go
// Bad
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userId)

// Good
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, userId)
```

#### File path traversal
```
Error: Potential file inclusion via variable (G304)
```
**Fix**: Validate and sanitize file paths, use filepath.Clean()

#### Command injection
```
Error: Subprocess launched with variable (G204)
```
**Fix**: Validate input or use safer alternatives

## 4. Unit Tests

**Purpose**: Run unit tests and check coverage
**Configuration**: `.github/workflows/.testcoverage.yml`
**Script**: `scripts/run_tests.sh`

### Common Issues and Fixes

#### Test failures
```
Error: Test failed: TestFunctionName
```
**Fix**: Debug the failing test:
1. Run the specific test: `go test -v -run TestFunctionName`
2. Check test assertions and expected vs actual values
3. Fix the implementation or update the test

#### Low coverage
```
Error: Coverage 0.1% is below minimum 0.2%
```
**Fix**: Add more unit tests to increase coverage

#### Missing test files
```
Error: tests/test_coverage_get.sh not found
```
**Fix**: Ensure the test infrastructure is properly set up

## 5. Docker Build

**Purpose**: Verify Docker image builds successfully
**Configuration**: `Dockerfile` in project root
**Script**: `scripts/run_docker_build.sh`

### Common Issues and Fixes

#### Build failures
```
Error: failed to solve with frontend dockerfile.v0
```
**Fix**: Check Dockerfile syntax and build context

#### Private repository access
```
Error: fatal: could not read Username for 'https://github.com'
```
**Fix**: Set environment variables:
```bash
export GITHUB_USER=your-username
export GITHUB_TOKEN=your-token
bash scripts/run_docker_build.sh
```

#### Dependency issues
```
Error: go get: module not found
```
**Fix**: Ensure go.mod is up to date and dependencies are accessible

## Workflow Integration

The GitHub Actions workflows are defined in `.github/workflows/`:
- `ci.yml` - Docker build and gosec scan
- `lint.yml` - golangci-lint
- `typos.yml` - spell checking
- `unit_test.yml` - unit tests and coverage

All workflows run on pull requests to ensure code quality before merging.
