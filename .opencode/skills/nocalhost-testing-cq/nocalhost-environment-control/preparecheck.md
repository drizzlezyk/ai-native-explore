# Nocalhost Prepare Checklist

## Before Running `prepare` Command

You MUST collect ALL of the following parameters. Missing any required field will cause the prepare command to fail.

---

### Required Parameters (CLI Flags - Ask User for Each)

These 4 parameters MUST be provided as CLI flags to `prepare`:

| JSON Field (camelCase) | CLI Flag (kebab-case) | Description | Example | Status |
|------------------------|----------------------|-------------|---------|--------|
| `developerName` | `--developer-name` | Your developer identifier | `john-doe` | ☐ |
| `kubeconfig` | `--kubeconfig` | Full path to kubeconfig file | `~/.kube/xihe-test_kubeconfig` | ☐ |
| `namespace` | `--namespace` | Kubernetes namespace from kubeconfig context | `xihe-test` | ☐ |
| `origDeployName` | `--orig-deploy-name` | Original deployment name in Kubernetes | `xihe-server` | ☐ |

**Total: 4 CLI flag parameters**

---

### Auto-Derived Values (calculate yourself,ask user if need)

These 8 values are derived by analyzing the codebase (stored in config.json, not CLI flags):

| JSON Field | How to Derive | Where to Look |
|------------|---------------|---------------|
| `binaryName` | Parse Dockerfile `go build -o <binary>` | Dockerfile |
| `remotePort` | Parse Dockerfile EXPOSE or ENTRYPOINT `--port` | Dockerfile |
| `heartbeatUrl` | Parse server code for `/heartbeat` endpoint | server/*.go |
| `projectPath` | Use current working directory | - |
| `appConfig` | Fixed path | Always `.nocalhost/app.yaml` |
| `deployConfig` | Fixed path | Always `.nocalhost/config.yaml` |


### Auto-Derived then ask user to check the script
| `startupScript` | Generate from Dockerfile ENTRYPOINT + server code | See Rule 1 below |
| `buildScript` | Generate from Dockerfile build command | See Rule 2 below |

**Total: 8 auto-derived values**

---

## Parameter Count Verification

- **4 CLI flags** + **8 auto-derived** = **12 total** (matches config.json)

---

## How to Auto-Derive Parameters

### Rule 1: Generate `startupScript` (startup.sh)

**Template location:** `.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/startup.sh`

**Steps to generate project-specific startup.sh:**

1. **Find port:**
   - Search Dockerfile for `EXPOSE` or `ENTRYPOINT` with `--port`
   - If not found, search server code for `flag.*port` or default port values
   - Command: `grep -r "\--port\|EXPOSE\|:port" Dockerfile server/`

2. **Find config file path:**
   - Search Dockerfile ENTRYPOINT for `--config-file`
   - Command: `grep "config-file" Dockerfile`

3. **Find heartbeat endpoint:**
   - Search server code for `/heartbeat` or `/health` endpoint
   - Command: `grep -r "GET.*heartbeat\|GET.*health" server/`
   - Verify the endpoint returns JSON with "status":"ok"

4. **Find binary name:**
   - From Dockerfile `go build -o <binary>` or `ENTRYPOINT`
   - Command: `grep "go build.*-o\|ENTRYPOINT" Dockerfile`

5. **Find template files (if any):**
   - Search Dockerfile for `COPY` commands that copy template/config files
   - These need to be copied in startup.sh before server starts

**Generated startup.sh should:**
- Setup any required directories
- Copy template/config files if needed
- Handle secrets backup/restore (vault pattern)
- Start the binary with correct port and config file path

**Important:** The `config-file` path (e.g., `/vault/secrets/application.yml`) comes from Dockerfile ENTRYPOINT `--config-file` flag. This is NOT a separate parameter - it should be hardcoded in the generated startup.sh based on the Dockerfile.

**Example generated output:**
```bash
#!/bin/bash
set -e

# Setup templates
mkdir -p /opt/app/points/task-docs-templates
cp ./points/infrastructure/taskdocimpl/doc_chinese.tmpl /opt/app/points/task-docs-templates/ 2>/dev/null || true
cp ./points/infrastructure/taskdocimpl/doc_english.tmpl /opt/app/points/task-docs-templates/ 2>/dev/null || true

# Setup secrets
if [ -d "/vault/secrets" ] && [ "$(ls -A /vault/secrets 2>/dev/null)" ]; then
    mkdir -p /vault/backup
    cp -r /vault/secrets/* /vault/backup/ 2>/dev/null || true
fi

# Start server
export HOME=/home/nocalhost-dev
./{{.binaryName}} --port {{.remotePort}} --config-file /vault/secrets/application.yml --enable_debug
```

---

### Rule 2: Generate `buildScript` (build.sh)

**Template location:** `.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/build.sh`

**Steps to generate project-specific build.sh:**

1. **Find binary name:**
   - Command: `grep "go build.*-o" Dockerfile`
   - Extract the `-o <binary>` value

2. **Find source path:**
   - Command: `grep "go build.*\./" Dockerfile`
   - Usually `./main.go` or `./cmd/server`

3. **Find build flags:**
   - Look for `-mod=vendor`, `-buildvcs=false`, CGO settings
   - Command: `grep "go build" Dockerfile`

4. **Determine if vendor sync needed:**
   - If Dockerfile uses `-mod=vendor`, include `--sync-vendor` in build script
   - Check if `vendor/` directory exists locally

**Generated build.sh should:**
- Setup HOME and GOCACHE directories
- Use `-mod=vendor` if required by project
- Build to the correct binary name

**Example generated output:**
```bash
#!/bin/bash
set -e

export HOME=/home/nocalhost-dev
export GOCACHE=/home/nocalhost-dev/.cache/go-build
mkdir -p /home/nocalhost-dev/.cache
cd /home/nocalhost-dev
go build --buildvcs=false -mod=vendor -o {{.binaryName}} {{./source/path}}
```

---

### Rule 3: Derive `remotePort` and `heartbeatUrl`

**remotePort:**
1. Search Dockerfile: `grep "EXPOSE\|--port" Dockerfile`
2. Search server main: `grep "port.*:\|flag.*port" main.go server/*.go`
3. Default to common ports (5000, 8000, 8080) if not found

**heartbeatUrl:**
1. Search server code: `grep -r "GET.*heartbeat\|GET.*health\|POST.*health" .`
2. Extract the route path
3. Construct URL: `http://localhost:{remotePort}{heartbeatPath}`

---

## Where to Find Deployment YAML Info

Get `origDeployName`, `namespace`, and `kubeconfig` from your Kubernetes cluster:

```bash
# List deployments to find origDeployName
kubectl get deployment -A | grep -i <project-name>

# Get deployment details (contains namespace, port, etc.)
kubectl get deployment <origDeployName> -n <namespace> -o yaml

# Or get the kubeconfig path
ls ~/.kube/*kubeconfig*
```

---

## Checklist Before Proceeding

### CLI Flags (4 - MUST ask user)
- [ ] `--developer-name` provided
- [ ] `--kubeconfig` provided  
- [ ] `--namespace` provided
- [ ] `--orig-deploy-name` provided

### Auto-Derived Values (8 - calculate yourself, DO NOT ask user)

**From Dockerfile:**
- [ ] `binaryName` - extracted from `go build -o <binary>`
- [ ] `remotePort` - extracted from Dockerfile EXPOSE or `--port`
- [ ] `configFile` - extracted from Dockerfile `--config-file` (used in startup.sh)

**From Server Code:**
- [ ] `heartbeatUrl` - constructed from `/heartbeat` endpoint + `remotePort`
- [ ] `projectPath` - current working directory

**Fixed Paths:**
- [ ] `appConfig` - `.nocalhost/app.yaml`
- [ ] `deployConfig` - `.nocalhost/config.yaml`

**Generated Scripts:**
- [ ] `startupScript` - `.nocalhost/startup.sh` generated from template
- [ ] `buildScript` - `.nocalhost/build.sh` generated from template

### Generation Verification
- [ ] `startup.sh` generated: binary name matches Dockerfile, port matches, config-file path matches
- [ ] `build.sh` generated: source path exists in Dockerfile, build flags correct
- [ ] `heartbeatUrl` validated: `/heartbeat` endpoint exists in server code
- [ ] All 12 values present in `.nocalhost/.config.json`

### Infrastructure Validation
- [ ] Kubeconfig file path is valid and accessible
- [ ] Namespace exists in kubeconfig context
- [ ] origDeployName exists in the namespace
- [ ] Heartbeat URL is correct and reachable after server starts (test with `curl`)
