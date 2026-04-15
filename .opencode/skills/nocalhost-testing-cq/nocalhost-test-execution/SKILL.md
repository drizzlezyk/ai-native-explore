---
name: nocalhost-test-execution
description: Use when running existing nocalhost YAML API tests, checking reports, or validating endpoints against a prepared local debug server
---

# Nocalhost Test Execution

This skill runs existing YAML test cases against a prepared environment and inspects the generated reports.

## When To Use

Use this skill when the user wants to:
- run existing nocalhost API tests
- validate a specific test group against `localhost:8092`
- inspect or compare generated test reports

Do not use this skill to generate new YAML test cases or to prepare the environment from scratch.

## Prerequisites

Before running tests, confirm:
- the server is already reachable
- the required YAML test files exist under `tests/nocalhost-test/<group>/`
- required auth inputs are available when the test group needs them

## Commands

```bash
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go --help

go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud
```

## Output

Reports are written to `tests/nocalhost-test-report/`.
