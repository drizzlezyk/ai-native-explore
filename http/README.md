# HTTP/WebSocket Testing Toolkit

Bash-based testing framework for xihe-server API with cookie-based authentication and scenario-based testing.

## When to Use This Toolkit

**Use this toolkit when:**
- Testing HTTP API endpoints with cookie authentication
- Testing WebSocket connections with JWT tokens
- Running automated API test scenarios
- Validating API responses and data structures
- Testing cloud pod lifecycle operations

**Prerequisites:**
- xihe-server running on `http://localhost:8092` (or configured URL)
- Valid browser cookies from xihe-server
- Required tools: `curl`, `jq`, `bash`

## Quick Start

```bash
# 1. Get fresh cookies from browser DevTools
#    - Open browser DevTools (F12) → Network tab
#    - Make request to localhost:8092
#    - Copy Cookie header value
#    - Update http/cookiesplain with fresh cookie string

# 2. Parse cookies to Netscape format
./http/parse_cookie.sh

# 3. Run scenario tests
./http/scenario/run.sh "" "cloud_lifecycle"

# 4. List available scenarios
./http/scenario/run.sh "" "--list"
```

## Directory Structure

```
http/
├── cookiesplain              # Browser cookie string (update this with fresh cookies)
├── cookies.txt              # Generated Netscape format cookies (auto-created)
├── parse_cookie.sh          # Convert cookiesplain to cookies.txt
├── http_with_cookie.sh      # HTTP request helper with cookie management
├── ws_with_cookie.sh        # WebSocket connection helper with JWT auth
└── scenario/
    ├── run.sh              # Main scenario runner
    ├── scenarios.json        # Test scenario definitions
    ├── ws_mock_server.go    # Mock WebSocket server for testing
    ├── checks/             # Response validation scripts
    │   ├── check_cloud_list_ok.sh
    │   ├── check_pod_history_ok.sh
    │   ├── check_subscribe_ok.sh
    │   └── ...
    └── scenarios/          # Scenario cleanup/finalize scripts
        ├── cloud_lifecycle_cleanup.sh
        └── ...
```

## Core Components

### 1. Cookie Management

**`parse_cookie.sh`** - Convert browser cookies to Netscape format
```bash
# Uses default cookiesplain file
./http/parse_cookie.sh

# Or specify cookie string directly
./http/parse_cookie.sh "name1=value1; name2=value2"
```

**`cookiesplain`** - Browser cookie string format
```
_U_T_=eyJhbGc...; _Y_G_=dMNxPuk...; JSESSIONID=ABC123; ...
```

**`cookies.txt`** - Netscape format (auto-generated)
```
# Netscape HTTP Cookie File
localhost	FALSE	/	FALSE	18934560000	_U_T_	eyJhbGc...
localhost	FALSE	/	FALSE	18934560000	_Y_G_	dMNxPuk...
```

### 2. HTTP/WebSocket Helpers

**`http_with_cookie.sh`** - HTTP requests with automatic cookie refresh
```bash
# Source to use as library
source http/http_with_cookie.sh
http_with_cookie "http://localhost:8092/web/v1/cloud"

# Or run directly
./http/http_with_cookie.sh "http://localhost:8092/web/v1/cloud"
```

**`ws_with_cookie.sh`** - WebSocket connections with JWT authentication
```bash
# Source to use as library
source http/ws_with_cookie.sh
ws_with_cookie "ws://localhost:8092/web/v1/cloud/ascend_002"
```

### 3. Scenario Testing

**`scenario/run.sh`** - Main test runner
```bash
# Run all scenarios
./http/scenario/run.sh

# Run specific scenario by name
./http/scenario/run.sh "" "cloud_lifecycle"

# Run specific scenario by index
./http/scenario/run.sh "" "0"

# List available scenarios
./http/scenario/run.sh "" "--list"
```

**`scenario/scenarios.json`** - Test scenario definitions
```json
{
  "base_url": "http://localhost:8092",
  "scenarios": [
    {
      "name": "cloud_lifecycle",
      "description": "Full cloud pod lifecycle test",
      "steps": [
        {
          "name": "list_cloud_configs",
          "method": "GET",
          "url": "/web/v1/cloud",
          "extract": {
            "variable": "CLOUD_ID",
            "jq_path": ".data[0].id"
          },
          "check_script": "checks/check_cloud_list_ok.sh"
        }
      ],
      "final_verify_script": "scenarios/cloud_lifecycle_final.sh",
      "cleanup_script": "scenarios/cloud_lifecycle_cleanup.sh"
    }
  ]
}
```

## Available Scenarios

| Scenario name | Description | Purpose |
|-------------|-------------|-----------|
| `cloud_lifecycle` | Full pod lifecycle: list → subscribe → wait → info → release | Complete workflow test |
| `start_cloud` | Subscribe and wait for pod ready | Test pod creation |
| `stop_cloud` | Find running pod and release it | Test pod cleanup |
| `pod_history_basic` | Basic history read with pagination | Test data retrieval |
| `pod_history_filters` | Test filter parameters | Test query parameters |

## Key API Endpoints

**Important:** Different endpoints expect different parameters:

| Endpoint | Parameter | Description |
|----------|-----------|-------------|
| `GET /v1/cloud/pod/:cid` | `cloud_id` | Get pod info for cloud config |
| `DELETE /v1/cloud/pod/:id` | `pod_id` | Release specific pod |
| `GET /v1/ws/cloud/pod/:id` | `pod_id` | WebSocket for pod updates |
| `GET /v1/cloud/pod/history` | - | Get pod history with filters |

## Common Issues and Solutions

### Issue: "Empty response body" errors
**Cause:** Expired or invalid cookies
**Solution:**
```bash
# 1. Get fresh cookies from browser DevTools
# 2. Update http/cookiesplain with new cookie string
# 3. Re-parse cookies
./http/parse_cookie.sh
```

### Issue: "cloud_not_allowed" error
**Cause:** Running pod already exists
**Solution:** Run `stop_cloud` scenario first to release existing pod

### Issue: HTTP 204 No Content failures
**Cause:** DELETE operations return empty body (this is normal)
**Solution:** Framework handles this automatically - no action needed

### Issue: "doc doesn't exist" error
**Cause:** Pod already released or doesn't exist
**Solution:** Run `start_cloud` to create new pod

## Creating Custom Scenarios

1. **Add scenario to `scenarios.json`:**
```json
{
  "name": "my_test",
  "description": "My custom test",
  "steps": [
    {
      "name": "my_step",
      "method": "GET",
      "url": "/web/v1/endpoint",
      "body": "{\"param\":\"value\"}",
      "extract": {
        "variable": "MY_VAR",
        "jq_path": ".data.field"
      },
      "check_script": "checks/my_check.sh"
    }
  ]
}
```

2. **Create validation script** (`checks/my_check.sh`):
```bash
#!/bin/bash
RESPONSE_FILE="$1"
# Validate response using jq
code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] Expected empty code, got: $code" >&2
    exit 1
fi
echo "  [CHECK] Response valid"
exit 0
```

3. **Run scenario:**
```bash
./http/scenario/run.sh "" "my_test"
```

## Testing Workflow

1. **Setup:**
   - Ensure xihe-server is running
   - Get fresh cookies from browser
   - Update `http/cookiesplain`

2. **Parse cookies:**
   ```bash
   ./http/parse_cookie.sh
   ```

3. **Run tests:**
   ```bash
   # List scenarios
   ./http/scenario/run.sh "" "--list"
   
   # Run specific scenario
   ./http/scenario/run.sh "" "cloud_lifecycle"
   
   # Run all scenarios
   ./http/scenario/run.sh
   ```

4. **Debug:**
   - Check `http/cookies.txt` for valid cookies
   - Examine individual check scripts
   - Review scenario definitions in `scenarios.json`
   - Check server logs for errors

## Mock WebSocket Server

For local testing without full server:

```bash
# Run mock server (default port 19093)
cd http/scenario
go run ws_mock_server.go

# Or specify custom port
go run ws_mock_server.go 8080
```

Mock server simulates:
- Pod lifecycle events
- WebSocket connections with JWT auth
- Error conditions for testing

## Agent Usage Guidelines

**When agent should use this toolkit:**
1. User asks to test HTTP/WebSocket APIs
2. User mentions API testing or validation
3. User wants to test xihe-server endpoints
4. User needs automated API test scenarios

**Agent workflow:**
1. Check if cookies are valid (test simple endpoint)
2. If cookies expired, instruct user to get fresh cookies
3. Parse cookies: `./http/parse_cookie.sh`
4. List available scenarios: `./http/scenario/run.sh "" "--list"`
5. Run appropriate scenario based on user request
6. Analyze results and provide clear feedback

**Common agent mistakes to avoid:**
- Don't repeatedly run failing tests with expired cookies
- Don't invent syntax errors in working code
- Stop and suggest fresh cookies when authentication fails
- Check HTTP status codes before analyzing responses
- Understand that HTTP 204 No Content is valid for DELETE operations