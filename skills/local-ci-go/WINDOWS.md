# Windows Setup Guide for local-ci-go (Git Bash Only)

This guide standardizes Windows usage to Bash scripts only, matching Ubuntu workflow and reducing maintenance cost.

## Prerequisites

- Git for Windows (includes Git Bash): https://git-scm.com/download/win
- Go 1.16+: https://go.dev/doc/install
- Git repository with a Go project

## Quick Start

### 1. Open Git Bash

Open **Git Bash** from the Start menu, then go to your project root:

```bash
cd /c/path/to/your/project
```

### 2. Check Prerequisites

```bash
bash .claude/skills/local-ci-go/scripts/check_prerequisites.sh
```

### 3. Install Tools

```bash
bash .claude/skills/local-ci-go/scripts/install_tools.sh
```

### 4. Run CI Checks

```bash
# Run all checks
bash .claude/skills/local-ci-go/scripts/run_all_checks.sh

# Or run individual checks
bash .claude/skills/local-ci-go/scripts/run_tests.sh
bash .claude/skills/local-ci-go/scripts/run_security.sh
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh
```

## Gitleaks Modes

```bash
# Default (committed scan, CI-oriented)
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh

# Explicit modes
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh committed
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh all-branches
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh history
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh staged
bash .claude/skills/local-ci-go/scripts/run_gitleaks.sh uncommitted
```

## Security Auto-Fix

```bash
bash .claude/skills/local-ci-go/scripts/security-fix/orchestrator.sh --auto-fix
```

## Troubleshooting

### `gitleaks` Not Found

```bash
which gitleaks
bash .claude/skills/local-ci-go/scripts/install_tools.sh
```

If still not found, restart Git Bash and verify your PATH.

### `go` Not Found

```bash
go version
```

If command is missing, install Go and reopen Git Bash.

### Line Ending Issues

Git on Windows may convert CRLF/LF automatically. If scripts fail unexpectedly:

```bash
git config core.autocrlf input
```

Then re-check out affected files.

## Notes

- Windows and Ubuntu now share one script path: `.sh`.
- Do not use `.ps1` scripts in this skill (removed by design).
- For general usage, see `README.md`.
