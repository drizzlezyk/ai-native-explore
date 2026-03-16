---
name: nocalhost-testing-workflow
description: Main orchestrator for complete nocalhost testing workflow
---

# Nocalhost Testing Workflow

This skill orchestrates the complete nocalhost testing workflow by coordinating test management, environment control, and test execution.

## Workflow

```
1. nocalhost-test-management: generate (create test case templates)
2. nocalhost-environment-control: initialize (setup testing environment)
3. nocalhost-test-execution: validate (run tests and generate reports)
4. nocalhost-test-management: refine (improve test cases based on results)
```

## Usage

Invoke this skill when you want to run the complete end-to-end testing workflow from test generation through execution and refinement.

Each step can also be invoked independently:
- `nocalhost-test-management` - Generate or refine test cases
- `nocalhost-environment-control` - Manage nocalhost environment
- `nocalhost-test-execution` - Execute tests and generate reports

## Example Complete Workflow

```bash
# Step 1: Generate test cases
# (Use nocalhost-test-management skill)

# Step 2: Initialize environment
# (Use nocalhost-environment-control skill - it handles: up, rebuild, forward)

# Step 3: Execute tests
go run .opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=$XIHE_USERNAME

# Step 4: Refine test cases (manual)
# Review report and update test cases as needed
```

## Workflow Notes

- The workflow is iterative: after refining test cases, re-run steps 3-4
- Each step can be invoked independently for focused work
- **Always use skills for server control** - never run nocalhostctl scripts directly
- Reports persist in `tests/nocalhost-test-report/` for historical analysis