---
name: nocalhost-test-management
description: Analyzes Go controller code and generates YAML test cases for API testing. Use when you need to create test cases for a new or existing API endpoint.
---

# Nocalhost Test Case Generator

This skill analyzes Go controller source files to extract API endpoint information and generates YAML test cases.

**info you should know** `.open/skills/nocalhost-testing/nocalhost-test-management/implement.md`

## Usage

When user asks to generate test cases for an API, provide:
1. Controller file path (e.g., `controller/cloud.go`)
2. Function name (e.g., `GetHistory`)

The skill will:
1. Read the Go source file
2. Parse to extract: URL, method, query params, auth requirements
3. Generate YAML test cases with variations

## Code Analysis Patterns

### URL and Method
Look for router registration:
```go
rg.GET("/v1/cloud/pod/history", ctl.GetHistory)
```

### Query Parameters
Look for:
```go
ctx.Query("spec")
ctx.Query("image_alias")
ctx.Query("page_num")
ctx.Query("page_size")
```

### Auth Requirements
Look for:
```go
pl, visitor, ok := ctl.checkUserApiTokenV2(ctx, true)
```

## Output Format

Generate YAML to `tests/nocalhost-test/<group>/<endpoint>.yaml`:

**Important:** When `auth_required: true`, always set `debug_if_no_cookie: true` by default. This will prompt for username and automatically bypass auth.

```yaml
- name: "<description>"
  url: "/web/v1/..."
  method: "GET"
  expected_status: 200
  auth_required: true
  debug_if_no_cookie: true
  query_params:
    - key: "param"
      value: "test-value"
  description: "..."

- name: "<description> (without auth)"
  url: "/web/v1/..."
  method: "GET"
  expected_status: 401
  auth_required: false
  description: "Should return 401 without authentication"
```

## Examples

### Input
"Generate test cases for controller/cloud.go GetHistory function"

### Output
Save to `tests/nocalhost-test/cloud/pod_history.yaml`

## Test Case Review Loop

After generating the test cases:

1. Dispatch a single `test-case-reviewer` subagent (see test-case-reviewer-prompt.md) with precisely crafted review context — never your session history. This keeps the reviewer focused on the test cases, not your thought process.
   - Provide: path to the generated YAML test case file, path to the source Go controller file.
2. If ❌ Issues Found: fix the issues (e.g., incorrect URL, missing parameters, wrong status codes), re-dispatch reviewer for the whole file.
3. If ✅ Approved: proceed to next steps (initialization and execution).

**Review loop guidance:**
- Same agent that generated the test cases fixes them (preserves context).
- If loop exceeds 3 iterations, surface to human for guidance.
- Reviewers are advisory — explain disagreements if you believe feedback is incorrect.

## Refine Mode (Step 4)

After test execution (via nocalhost-test-execution skill), review the generated report at `tests/nocalhost-test-report/{timestamp}report.md` to identify failing tests.

Common refinement tasks:
- Update expected status codes
- Fix incorrect query parameters
- Adjust authentication requirements
- Add edge case test scenarios

Refine test cases in `tests/nocalhost-test/<group>/` and re-run tests using the nocalhost-test-execution skill to verify fixes.

## Integration with Workflow

This skill handles **Step 1** (generate) and **Step 4** (refine) of the 4-step nocalhost testing workflow:
1. **nocalhost-test-management: generate** ← Generate mode
2. nocalhost-environment-control: initialize
3. nocalhost-test-execution: validate
4. **nocalhost-test-management: refine** ← Refine mode
