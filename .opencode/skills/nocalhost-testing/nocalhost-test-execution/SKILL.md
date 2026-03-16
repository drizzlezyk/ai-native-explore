---
name: nocalhost-test-execution
description: Test execution and reporting for nocalhost testing
---

# Nocalhost Test Execution

This skill handles test execution against running environment and report generation.

## Purpose

Execute test cases against against a running xihe-server instance and generate detailed test reports.

## Prerequisites

1. **Server running**: Server must be deployed and running (via nocalhost-environment-control skill)
2. **Test cases generated**: Test cases must exist in `tests/nocalhost-test/<group>/`
3. **Port forwarding**: Server must be accessible (default: http://localhost:8092)

## Workflow

1. Load test cases from YAML files
2. Execute HTTP requests against running server

3. Handle authentication (cookie-based or debug mode bypass)
4. Generate detailed Markdown reports

## Using This Skill

### Step 1: Ensure Server is Running

Use the **nocalhost-environment-control** skill to manage the server:

```
User: "Start the nocalhost testing environment"
Skill: Invokes nocalhost-environment-control skill to:
- Initialize dev environment (up)
- Rebuild server (rebuild)
- Port forward (forward)
```

### Step 2: Execute Tests

Run the test runner:

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=$XIHE_USERNAME
```

## Available Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | http://localhost:8092 | Base URL of the server |
| `--cookie` | (none) | Path to cookie file (optional) |
| `--group` | cloud | Test group to run (cloud, user, etc.) |
| `--user` | XIHE_USERNAME env var | Username for auth bypass |
| `--pod` | xihe-server-i9a-a-3bfe2eb6-69986c97b5-x7kpx | Pod name for nocalhost dev |

## Test Case Format

Test cases are defined in YAML format at `tests/nocalhost-test/<group>/<endpoint>.yaml`:

```yaml
- name: "Test description"
  url: "/web/v1/..."
  method: "GET"
  expected_status: 200
  auth_required: true
  debug_mode_if_no_cookie: true
  query_params:
    - key: "param"
      value: "value"
  description: "Detailed description"
```

### Test Case Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Test case name |
| `url` | string | Yes | API endpoint URL |
| `method` | string | Yes | HTTP method (GET, POST, etc.) |
| `expected_status` | int | Yes | Expected HTTP status code |
| `auth_required` | bool | No | Whether authentication is required (default: false) |
| `debug_mode_if_no_cookie` | bool | No | Enable debug mode bypass if no cookie (default: false) |
| `query_params` | array | No | Query parameters |
| `description` | string | No | Detailed test description |

## Report Generation

Reports are generated in Markdown format at:
`tests/nocalhost-test-report/{timestamp}report.md`

Report includes:
- Summary (passed/failed/total)
- Individual test results with:
  - URL and method
  - Expected vs actual status
  - Query parameters
  - Response body (truncated if too long)
  - Error details (if failed)

### Example Report

```markdown
# Nocalhost Test Report

Generated: 2026-03-16T09:00:00Z
Base URL: http://localhost:8092

## Summary

- **Passed**: 5
- **Failed**: 1
- **Auth Replaced**: 0
- **Total**: 6

## Test Results

### 1. Get pod history with id filter
- **URL**: `GET /web/v1/cloud/pod/history`
- **Expected Status**: 200
- **Actual Status**: 200
- **Auth Required**: true
- **Debug Mode If No Cookie**: true
- **Timestamp**: 2026-03-16T09:00:00Z
- **Query Parameters**:
  - `id`: `ascend_002`
- **Status**: PASSED ✓
- **Response Body**:
```
{"code":0,"msg":"success","data":{...}}
```
```

## Authentication Handling

### Cookie-based Authentication

Provide a cookie file with the `--cookie` flag:

```bash
--cookie=/path
/to/cookies.txt
```

The cookie file should contain the raw cookie string:

```
session_id=abc123; user_token=xyz789
```

### Debug Mode Bypass

Set `debug_mode_if_no_cookie: true` in test case YAML. When no cookie is provided:

1. Prompt for username (or use `--user` flag)
2. Enable debug mode on server (assumes server already running with `--enable_debug`)
3. Bypass authentication for testing

Example:

```yaml
- name: "Test with debug mode"
  url: "/web/v1/endpoint"
  method: "GET"
  expected_status: 200
  auth_required: true
  debug_mode_if_no_cookie: true
  description: "Test with debug mode auth bypass"
```

## Integration with Workflow

This is **Step 3** of the 4-step nocalhost testing workflow:
1. nocalhost-test-management: generate
2. nocalhost-environment-control: initialize
3. **nocalhost-test-execution: validate** ← You are here
4. nocalhost-test-management: refine

## Example Complete Workflow

```bash
# Step 1: Generate test cases
# (Use nocalhost-test-management skill)

# Step 2: Initialize environment
# (Use nocalhost-environment-control skill - DO NOT run scripts directly)

# Step 3: Execute tests
go run .opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=$XIHE_USERNAME

# Step 4: Refine test cases (manual)
# Review report at tests/nocalhost-test-report/ and update test cases
```

## Important Notes

- **Always use the nocalhost-environment-control skill** for server management
- **Do not run nocalhostctl scripts directly** - let the skill handle it
- The test runner assumes the server is already running and port-forwarded
- Reports persist for historical analysis and debugging
- Response bodies are truncated at 10,000 characters to keep reports manageable

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `test file not found` | Ensure test cases exist at `tests/nocalhost-test/<group>/` |
| `failed to parse YAML` | Check YAML syntax in test case files |
| `connection refused` | Ensure server is running and port-forwarded |
| `401 errors` | Check if `debug_mode_if_no_cookie: true` is set or provide cookie file |
| `expected status mismatch` | Verify expected status code matches actual API behavior |

## See Also

- [nocalhost-testing-workflow](../nocalhost-testing-workflow/SKILL.md) - Complete workflow overview
- [nocalhost-test-management](../nocalhost-test-management/SKILL.md) - Test case generation and refinement
- [nocalhost-environment-control](../nocalhost-environment-control/SKILL.md) - Server management
