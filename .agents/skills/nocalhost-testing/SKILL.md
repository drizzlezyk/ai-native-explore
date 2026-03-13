---
name: nocalhost-testing
description: Use nocalhost to deploy xihe-server to k8s test environment and test external authentication. Use when you need to deploy the server to test environment and test the new auth integration with mindspere.cn cookies.
---

# Nocalhost Testing Skill

This skill helps deploy xihe-server to k8s test environment using nocalhost and test the external authentication integration.

## Quick Reference

```bash
# One-liner for rebuild after code changes
export XIHE_USERNAME=${XIHE_USERNAME:-}
kubectl cp --exclude='vendor' --exclude='.git' --exclude='*.log' $PROJECT_DIR/. ${POD_NAME}:/home/nocalhost-dev/ -c nocalhost-dev -n xihe-test-v2 && \
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c "pkill xihe-server || true" && \
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c "cp -r /vault/backup/* /vault/secrets/ 2>/dev/null || true" && \
if [ -n "$XIHE_USERNAME" ]; then
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c "export XIHE_USERNAME=$XIHE_USERNAME HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
else
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c "export HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
fi
```

## 1. Prerequisites

- **nocalhost installed**: `npm install -g nocalhost`
- **nhctl CLI** (comes with nocalhost)
- **Go modules vendored locally**: run `go mod vendor` before starting
- **Playwright**: use `npx playwright` or install globally
- **Kubeconfig**: `~/.kube/xihe-test-v2_kubeconfig`

## 2. Environment Variables

**Required before invoking this skill:**

```bash
export GITHUB_USERNAME="your-github-username"
export XIHE_USERNAME="your-xihe-account-username-for-auth-bypass"
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig
```

| Variable | Purpose |
|----------|---------|
| `GITHUB_USERNAME` | Identifies your dev pod |
| `XIHE_USERNAME` | Account for auth bypass (debug mode) |
| `KUBECONFIG` | Path to kubeconfig file |

The skill will FAIL FAST if any of these are missing.

## 3. Nocalhost Configuration

Create `.nocalhost/app.yaml` in your project root:

```yaml
configProperties:  
  version: v2  
application:  
  services:  
    - name: xihe-server  
      serviceType: deployment  
      containers:  
        - name: xihe-server  
          dev:  
            image: golang:1.24  
            shell: /bin/bash
```

Create `.nocalhost/config.yaml ` in your project root:
```yaml
name: xihe-server
resourcePath: ["deployments/xihe-server"]
serviceType: deployment
containers:
  - name: xihe-server
    dev:
      gitUrl: ""
      image: golang:1.24
      shell: bash
      workDir: /home/nocalhost-dev
      storageClass: ""
      resources: null
      persistentVolumeDirs: []
      command: null
      debug: null
      sync: null
      env:
        - name: GOCACHE
          value: /home/nocalhost-dev/.cache/go-build
        - name: GOPROXY
          value: https://goproxy.cn,direct
        - name: HOME
          value: /home/nocalhost-dev
      portForward: []
```



## 4. Workflow

### Step 1: Start Dev Duplicate Application

Create a duplicate dev pod (doesn't affect original deployment):

```bash
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# Install nocalhost app (first time only)
nhctl install xihe-server-$GITHUB_USERNAME -n xihe-test-v2 \
    --type rawManifestLocal \
    --local-path $PROJECT_DIR \
    --outer-config .nocalhost/app.yaml \
    --kubeconfig $KUBECONFIG

# Start dev mode in duplicate mode
output=$(nhctl dev start xihe-server-$GITHUB_USERNAME -n xihe-test-v2 \
    --dev-mode duplicate \
    -s $PROJECT_DIR \
    -d xihe-server \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  2>&1)

echo "$output"

# Extract deployment name
export DEPLOY_NAME=$(echo "$output" | grep "Kind=Deployment" | head -1 | awk '{print $2}' | cut -d'(' -f1)

# Extract pod name
export POD_NAME=$(echo "$output" | grep "Pod .* now (Running|Pending)" | head -1 | awk '{print $2}')

echo "DEPLOY_NAME: $DEPLOY_NAME"
echo "POD_NAME: $POD_NAME"
```

### Step 2: Copy Files to Pod

```bash
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# Run go mod vendor locally first
cd $PROJECT_DIR && go mod vendor

# Copy all files (including vendor directory) to pod
kubectl cp $PROJECT_DIR/. \
    ${POD_NAME}:/home/nocalhost-dev/ \
    -c nocalhost-dev \
    -n xihe-test-v2
```

### Step 3: Build on Pod

```bash
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# Backup vault secrets (in case config is removed)
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "mkdir -p /vault/backup && cp -r /vault/secrets/* /vault/backup/ 2>/dev/null || true"

# Copy template files (required for xihe-server to start)
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "mkdir -p /opt/app/points/task-docs-templates && \
     cp ./points/infrastructure/taskdocimpl/doc_chinese.tmpl /opt/app/points/task-docs-templates/ && \
     cp ./points/infrastructure/taskdocimpl/doc_english.tmpl /opt/app/points/task-docs-templates/"

# Build with vendor
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "HOME=/home/nocalhost-dev cd /home/nocalhost-dev && go build --buildvcs=false -mod=vendor ."
```

### Step 4: Start Server

```bash
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# Restore vault secrets (in case config was removed)
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "cp -r /vault/backup/* /vault/secrets/ 2>/dev/null || true"

# Start with debug mode (if XIHE_USERNAME is set)
if [ -n "$XIHE_USERNAME" ]; then
    echo "Starting server with debug mode for user: $XIHE_USERNAME"
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
        "export XIHE_USERNAME=$XIHE_USERNAME; export HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
else
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
        "export HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
fi

# Check server log
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- cat server.log
```

### Step 5: Port Forward

```bash
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig 
kubectl port-forward -n xihe-test-v2 $POD_NAME 8092:8000 --kubeconfig $KUBECONFIG 2>&1 &
```

### Step 6: Get Cookies (Optional)

For production testing with real auth:

```bash
npx playwright install chromium  # one-time setup
npx -y playwright node scripts/get_cookies.js
```

Login to mindspere.cn, press Enter.

## 5. Running Tests

### Generate Test Cases

Use the **nocalhost-test-case-generator** skill to generate test cases:

```
User: "Generate test cases for controller/cloud.go GetHistory"
Skill: Analyzes code, creates YAML at tests/nocalhost-test/cloud/pod_history.yaml
```

### Execute Tests

```bash
# With debug mode auth bypass (uses XIHE_USERNAME env var)
go run tests/nocalhost-test/runner.go --url=http://localhost:8092 --user=$XIHE_USERNAME

# Run specific group
go run tests/nocalhost-test/runner.go --url=http://localhost:8092 --group=cloud --user=$XIHE_USERNAME

# With cookie file (recommended for production testing)
go run tests/nocalhost-test/runner.go --url=http://localhost:8092 --cookie=auth.json

# Skip cleanup (keep pod running after tests)
go run tests/nocalhost-test/runner.go --url=http://localhost:8092 --cleanup=false
```

### Test Case Format

```yaml
- name: "Test name"
  url: "/api/v1/endpoint"
  method: "GET"
  expected_status: 200
  auth_required: true
  debug_mode_if_no_cookie: true
  query_params:
    - key: "param"
      value: "value"
  description: "Test description"
```

### Auth Options

| Method | When to Use | Command |
|--------|-------------|---------|
| **Debug mode** | Local dev, quick testing | `--user=$XIHE_USERNAME` |
| **Cookie file** | Production testing | `--cookie=auth.json` |
| **Real auth** | Full integration test | (no flags, use browser) |

Debug mode works by reading `XIHE_USERNAME` env var and bypassing auth in `checkUserApiTokenV2`.

### Cleanup Behavior

- If all tests pass → stops nocalhost dev deployment automatically, clears `XIHE_USERNAME` env var
- If any test fails → keeps pod running for debugging

## 6. Incremental Development

After initial build, modify code locally and rebuild quickly:

```bash
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# 1. Copy modified files (excluding vendor, .git, etc.)
kubectl cp --exclude='vendor' --exclude='.git' --exclude='*.log' \
    $PROJECT_DIR/. \
    ${POD_NAME}:/home/nocalhost-dev/ \
    -c nocalhost-dev \
    -n xihe-test-v2

# 2. Rebuild on pod
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "HOME=/home/nocalhost-dev cd /home/nocalhost-dev && go build --buildvcs=false -mod=vendor ."

# 3. Restart server
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c "pkill xihe-server || true"
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
    "touch /vault/secrets/application.yml"
if [ -n "$XIHE_USERNAME" ]; then
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
        "export XIHE_USERNAME=$XIHE_USERNAME; export HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
else
    kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash -c \
        "export HOME=/home/nocalhost-dev; cd /home/nocalhost-dev && nohup ./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug > server.log 2>&1 &"
fi

# Check log
kubectl exec -n xihe-test-v2 $POD_NAME -c nocalhost-dev --kubeconfig $KUBECONFIG -- cat server.log
```

## 7. Cleanup

```bash
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# End dev mode
nhctl dev end xihe-server-$GITHUB_USERNAME -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

# Delete deployment
kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG

# Uninstall app (complete cleanup)
nhctl uninstall xihe-server-$GITHUB_USERNAME -n xihe-test-v2 --kubeconfig $KUBECONFIG
```

## 8. Common Commands

```bash
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig

# List dev pods (filtered by GITHUB_USERNAME)
kubectl get pods -n xihe-test-v2 --kubeconfig $KUBECONFIG | grep xihe-server-$GITHUB_USERNAME

# Check server logs
kubectl logs -n xihe-test-v2 <pod-name> -c nocalhost-dev --kubeconfig $KUBECONFIG

# Execute in dev container
kubectl exec -n xihe-test-v2 <pod-name> -c nocalhost-dev --kubeconfig $KUBECONFIG -- bash
```

## 9. Troubleshooting

| Issue | Solution |
|-------|----------|
| Go version error | Use `golang:1.24` image or set `GOTOOLCHAIN=auto` |
| Config not found | Use `--rm-cfg true` flag |
| 401 errors | Check if external auth API is accessible |
| Connection refused | Verify port-forward is working |

---

### Useful Nocalhost Commands

```bash
# Uninstall app
nhctl uninstall xihe-server-$GITHUB_USERNAME -n xihe-test-v2 --kubeconfig $KUBECONFIG

# List app
nhctl list xihe-server-$GITHUB_USERNAME -n xihe-test-v2 --kubeconfig $KUBECONFIG

# End dev
nhctl dev end xihe-server-${GITHUB_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

# Reset dev
nhctl dev reset xihe-server-${GITHUB_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

# Delete deployment directly
kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG
```
