---
name: nocalhost-test-case-generator
description: Analyzes Go controller code and generates YAML test cases for API testing. Use when you need to create test cases for a new or existing API endpoint.
---

# Nocalhost Test Case Generator

This skill analyzes Go controller source files to extract API endpoint information and generates YAML test cases.

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

**Important:** When `auth_required: true`, always set `replace_if_no_cookie: true` by default. This will prompt for username and automatically bypass auth.

```yaml
- name: "<description>"
  url: "/web/v1/..."
  method: "GET"
  expected_status: 200
  auth_required: true
  replace_if_no_cookie: true
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
