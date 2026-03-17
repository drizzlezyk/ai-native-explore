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
# 0.check file fisrt
ls .nocalhost/.config.json
ls .nocalhost/.state.json

# 1. Initial setup (install & start dev mode)
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go up

# 2. Rebuild & Restart (sync code including vendor, build inside pod, restart server)
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go rebuild --sync-vendor

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

## 2. Workflow

### Step 0: Prepare Configuration

Save required configuration parameters:

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go prepare \
  --xihe-user="your-xihe-account-username" \
  --kubeconfig=~/.kube/xihe-test-v2_kubeconfig
```

### Step 1: Initialize Dev Environment

This command installs the application (if needed) and starts dev mode in **duplicate mode**. It captures the pod name and saves it to a local `.state.json` file.

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go up
```

### Step 2: First-Time Setup (Sync, Build, Run)

**For first-time setup only**, run these commands sequentially to start the server:

1. **Sync**: Copy files to the pod (including `vendor` if --sync-vendor is used)
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go sync --sync-vendor
```

2. **Build**: Build the binary inside the pod using `go build -mod=vendor`
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go build
```

3. **Run**: Start the server using the `startup.sh` script inside the pod
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go run
```

### Step 3: Incremental Development (Rebuild)

**After first-time setup**, use this command when you change your code. It will:
1. **Sync**: Copy modified files to the pod (excluding `vendor`, `.git`, etc.)
2. **Build**: Build binary inside the pod using `go build -mod=vendor`.
3. **Run**: Restart the server using `startup.sh` script inside the pod.

**Important**: The build command uses `-mod=vendor` which requires the vendor directory. Use `--sync-vendor` flag to include it:

```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go rebuild --sync-vendor
```

**Note**: For subsequent rebuilds after the first vendor sync, you can omit `--sync-vendor` for faster sync:
```bash
go run .opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/main.go rebuild
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

When running in a subagent context, provide this tip to the main agent:

"Execute tests using the nocalhost-test-execution skill by running:
go run .opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=<xihe-username-from-prepare>"

The main agent will execute this command.

## 5. Troubleshooting

| Issue | Solution |
|-------|----------|
| `Configuration not found` | Run `prepare` command first with required parameters. |
| `State not found` | Run `up` command first. |
| `Pod not found` | Your pod might have been deleted. Run `up` again to re-discover. |
| `Build failed` | Ensure dependencies are correct in go.mod. The script auto-runs `go mod vendor` when using `--sync-vendor`. |
| `Build failed: missing vendor directory` | Use `--sync-vendor` flag with rebuild command. Build uses `-mod=vendor` which requires vendor directory. |
| `401 errors` | Check if XIHE_USERNAME was set correctly in prepare command. |

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
