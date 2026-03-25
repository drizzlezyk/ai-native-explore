---
name: nocalhost-test-execution
description: Test execution and reporting for nocalhost testing
---

# Nocalhost Test Execution

This skill handles test execution against running environment and report generation.

## Purpose

Execute test cases against a running ${DEPLOYMENT_NAME} instance and generate detailed test reports.

## Prerequisites

1. **Server running**: Server must be deployed and running (via nocalhost-environment-control skill)
2. **Test cases generated**: Test cases must exist in `tests/nocalhost-test/<group>/`
3. **Port forwarding**: Server must be accessible (default: http://localhost:8092)

## Quick Start

```bash
go run -tags debug ./.ai/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go --help
```

## Available Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | http://localhost:8092 | Base URL of the server |
| `--group` | cloud | Test group to run (cloud, user, etc.) |
| `--user` | DEVELOPER_NAME env var | Username for auth bypass |

## Examples

```bash
# Run cloud tests
go run -tags debug ./.ai/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud

# Run user tests with specific username
go run -tags debug ./.ai/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --group=user \
  --user=testuser
```

## Test Case Format

Test cases are defined in YAML format at `tests/nocalhost-test/<group>/<endpoint>.yaml`:

```yaml
- name: "Test description"
  url: "/web/v1/..."
  method: "GET"
  expected_status: 200
  auth_required: true
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
| `query_params` | array | No | Query parameters |
| `description` | string | No | Detailed test description |

## Authentication

When `auth_required: true` is set, the runner uses debug mode bypass:

1. Username is taken from `--user` flag or `DEVELOPER_NAME` env var
2. If neither is set, prompts for username
3. Server must be running with debug mode enabled

```yaml
- name: "Test requiring auth"
  url: "/web/v1/endpoint"
  method: "GET"
  expected_status: 200
  auth_required: true
  description: "Test with debug mode auth bypass"
```

## Report Generation

Reports are generated in Markdown format at:
`tests/nocalhost-test-report/{timestamp}-report.md`

Report includes:
- Summary (passed/failed/total)
- Individual test results with:
  - URL and method
  - Expected vs actual status
  - Query parameters
  - Response body (truncated if too long)
  - Error details (if failed)

## Integration with Workflow

This is **Step 3** of the 4-step nocalhost testing workflow:

1. nocalhost-test-management: generate test cases
2. nocalhost-environment-control: start server
3. **nocalhost-test-execution: run tests** ← You are here
4. Review reports and refine tests

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `test directory not found` | Ensure test cases exist at `tests/nocalhost-test/<group>/` |
| `failed to parse YAML` | Check YAML syntax in test case files |
| `connection refused` | Ensure server is running and port-forwarded |
| `401 errors` | Ensure `--user` is set or `DEVELOPER_NAME` env var is configured |
| `expected status mismatch` | Verify expected status code matches actual API behavior |

## See Also

- [nocalhost-testing-workflow](../nocalhost-testing-workflow/SKILL.md) - Complete workflow overview
- [nocalhost-test-management](../nocalhost-test-management/SKILL.md) - Test case generation
- [nocalhost-environment-control](../nocalhost-environment-control/SKILL.md) - Server management
