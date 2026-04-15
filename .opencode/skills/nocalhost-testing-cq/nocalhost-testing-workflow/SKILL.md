---
name: nocalhost-testing-workflow
description: Use when working on nocalhost-based API testing, including test-case generation, test execution, Kubernetes dev-environment setup, or `oneclickstart` requests
---

# Nocalhost Testing Workflow

This skill is the entry router for the nocalhost testing package.

It must classify the request, route to the correct sub-skill, and block eager environment startup until readiness is explicitly proven.

## Routing Rules

Route by user intent instead of forcing a fixed sequence.

- Generate or refine YAML test cases: use `nocalhost-test-management`
- Run existing YAML test cases and inspect reports: use `nocalhost-test-execution`
- Prepare, validate, start, rebuild, forward, or troubleshoot the nocalhost environment: use `nocalhost-environment-control`
- Full end-to-end workflow: start with `nocalhost-environment-control` readiness, then continue to test generation or execution as needed

## Hard Readiness Gate

Before any `nocalhost` startup action, the environment skill must prove readiness.

Readiness is not assumed. Readiness is not inferred from intent. Readiness is not bypassed by `oneclickstart`.

Do not run any of these startup commands while readiness is `unknown` or `not-ready`:
- `up`
- `oneclickstart`
- `rebuild`
- `forward`
- `run`

When readiness has not been proven, stop and collect or verify the missing information first. `prepare` is allowed only after the required inputs are known, and startup remains blocked until the post-prepare checks pass.

## Required Readiness Evidence

The environment skill should confirm all of the following before startup:
- required user-owned inputs are present
- repo-derived values are resolved from the current codebase
- generated config and helper scripts exist and are plausible after `prepare`
- `preparecheck.md` passes
- the user confirms any uncertain derived values before startup continues

## Minimal Workflow

```
1. Classify the request.
2. Route to the correct nocalhost skill.
3. If environment actions are needed, require readiness evidence first.
4. Only after readiness is proven, allow startup or execution steps.
```

## Examples

- "Generate test cases for `controller/cloud.go` `GetHistory`" -> `nocalhost-test-management`
- "Run the cloud nocalhost tests against localhost:8092" -> `nocalhost-test-execution`
- "Prepare and start the nocalhost debug environment" -> `nocalhost-environment-control`
- "Do the full nocalhost testing flow" -> readiness first, then route to generation or execution

## What This Skill Must Not Do

This skill must not:
- embed a full startup recipe
- tell the agent to run `oneclickstart` right after `prepare`
- assume the environment is ready because the user asked for nocalhost testing
- skip readiness collection just to move faster
