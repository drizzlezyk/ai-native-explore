---
name: nocalhost-environment-control
description: Use when preparing, validating, starting, rebuilding, forwarding, troubleshooting, or `oneclickstart` requests for the nocalhost Kubernetes dev environment for API testing
---

# Nocalhost Environment Control

This skill handles environment readiness and environment actions for nocalhost-based testing.

Its first responsibility is to determine whether the environment is ready. It must not rush into startup.

## State Model

Treat the environment as one of these states:
- `unknown`: no readiness evidence yet
- `not-ready`: missing inputs or failed validation
- `ready`: required inputs and validations are complete
- `running`: environment is already started and reachable

## Hard Gate

Never start `nocalhost` while state is `unknown` or `not-ready`.

Blocked startup actions before readiness:
- `up`
- `oneclickstart`
- `sync`
- `build`
- `run`
- `rebuild`
- `forward`

When state is not `ready`, stop and fix readiness first. `prepare` is the bridge from collected inputs to validated readiness, so it is allowed only after the required inputs are known.

## Readiness Checklist

Required user-owned inputs:
- `developer-name`
- `kubeconfig`
- `namespace`
- `orig-deploy-name`

Required repo-derived inputs:
- `binary-name` from the Dockerfile build output
- `project-path` from the current repository
- `remote-port` from Dockerfile `EXPOSE` or server code
- `heartbeat-url` from the server code

The `prepare` command exposes override flags for these values, but the skill should treat them as derived defaults. Only override them when auto-derivation is wrong and the user confirms the correction.

Required validation evidence:
- generated files match the current repository shape
- `.config.json` and helper scripts are present after `prepare`
- `preparecheck.md` passes
- any low-confidence derived values are confirmed by the user before startup continues

## Working Sequence

1. Determine current state.
2. If state is `unknown` or `not-ready`, collect missing values and validate them.
3. Run `prepare` only after the required inputs are known.
4. Re-check readiness evidence, including generated files and `preparecheck.md`.
5. Only when state is `ready`, allow `up`, `run`, `rebuild`, `forward`, or `oneclickstart`.

## Commands

Use these commands only after the readiness gate passes.

```bash
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl prepare --help
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl status
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl up
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl rebuild --sync-vendor
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl forward
```

Use `oneclickstart` only when readiness has already been proven for this repository and environment.

## Common Mistakes

- Starting `nocalhost` immediately after the user mentions testing
- Treating `prepare` as proof that the environment is ready
- Skipping `preparecheck.md`
- Guessing derived values instead of reading the repository
- Running `oneclickstart` to bypass missing inputs
