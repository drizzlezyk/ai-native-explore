---
name: nocalhost-environment-control
description: Use nocalhost to deploy xihe-server to k8s test environment and test external authentication. Use when you need to deploy the server to test environment and test the new auth integration with mindspere.cn cookies.
---

# Nocalhost Environment Control

This skill provides a **stateful, automated** workflow for deploying and testing `xihe-server` in a Kubernetes test environment using Nocalhost.

This is **Step 2** of the 4-step nocalhost testing workflow.

## Quick Reference

The `nocalhostctl` tool (located in `.opencode/skills/nocalhost-testing/scripts/nocalhostctl/`) manages the entire lifecycle and maintains session state.

```bash
# 1. Initial setup (install & start dev mode)
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go up

# 2. Rebuild & Restart (sync code, build inside pod, restart server)
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go rebuild

# 3. View logs
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go logs

# 4. Port Forward
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go forward
```

## 1. Prerequisites

- **nocalhost installed**: `npm install -g nocalhost`
- **nhctl CLI** (comes with nocalhost)
- **Go modules vendored locally**: run `go mod vendor` before starting
- **Kubeconfig**: `~/.kube/xihe-test-v2_kubeconfig`

## 2. Environment Variables

Set these before running `nocalhostctl`:

```bash
export XIHE_USERNAME="your-xihe-account-username-for-auth-bypass"
export KUBECONFIG=~/.kube/xihe-test-v2_kubeconfig
```

| Variable | Purpose |
|----------|---------|
| `XIHE_USERNAME` | Xihe account for auth bypass and dev pod identification |
| `KUBECONFIG` | Path to kubeconfig file |

## 3. Workflow

### Step 1: Initialize Dev Environment

This command installs the application (if needed) and starts dev mode in **duplicate mode**. It captures the pod name and saves it to a local `.state.json` file.

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go up
```

### Step 2: First-Time Setup (Sync, Build, Run)

**For first-time setup only**, run these commands sequentially to start the server:

1. **Sync**: Copy files to the pod (excluding `vendor`, `.git`, etc.)
```bash
go mod vendor
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go sync  --sync-vendor
```

2. **Build**: Build the binary inside the pod using `go build -mod=vendor`
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go build
```

3. **Run**: Start the server using the `startup.sh` script inside the pod
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go run --user=$XIHE_USERNAME
```

### Step 3: Incremental Development (Rebuild)

**After first-time setup**, use this command when you change your code. It will:
1. **Sync**: Copy modified files to the pod (excluding `vendor`, `.git`, etc.)
2. **Build**: Build the binary inside the pod using `go build -mod=vendor`.
3. **Run**: Restart the server using the `startup.sh` script inside the pod.

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go rebuild --user=$XIHE_USERNAME
```

### Step 4: Monitor Logs

Tail the server log inside the pod:

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go logs
```

### Step 5: Port Forward

Expose the server locally (defaults to localhost:8092):

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go forward
```

### Step 6: Cleanup

Stop dev mode and uninstall the application:

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go down
```

## 4. Running Tests

### Generate Test Cases

Use the **nocalhost-test-management** skill:
```
User: "Generate test cases for controller/cloud.go GetHistory"
Skill: Creates YAML at tests/nocalhost-test/cloud/pod_history.yaml
```

### Execute Tests

Use the **nocalhost-test-execution** skill:
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=$XIHE_USERNAME
```

## 5. Troubleshooting

| Issue | Solution |
|-------|----------|
| `State not found` | Run the `up` command first. |
| `Pod not found` | Your pod might have been deleted. Run `up` again to re-discover. |
| `Build failed` | Ensure `go mod vendor` was run locally before syncing. |
| `401 errors` | Check if `XIHE_USERNAME` is correctly set and server is in debug mode. |

## 6. Directory Structure

All skill-specific tools are self-contained in:
`.opencode/skills/nocalhost-testing/nocalhost-environment-control/`
├── `configs/`
│   ├── `app.yaml`        # Nocalhost app config
│   └── `config.yaml`     # Nocalhost service config
├── `scripts/`
│   ├── `nocalhostctl/`   # Go orchestrator (main.go)
│   └── `startup.sh`      # Inside-pod entrypoint
└── `SKILL.md`            # This documentation
