---
name: http-scenario-generator
description: Generate runnable HTTP scenario files and JSON/YAML scenario specs for xihe cloud features. Use this skill whenever the user asks for HTTP test scenarios, end-to-end feature scenarios, or wants scenarios derived from `docs/knowledgebase/jupyter.md`, the codebase, kubeconfig, namespace, deployments, and logs. Prefer this skill even for a single feature request, because it should trace the whole cloud data flow and produce runnable scenarios that use the real HTTP port 8092.
---

# HTTP Scenario Generator

Generate runnable scenario files for one xihe feature by tracing the full data flow.

## What to produce

Produce both:
- a JSON/YAML scenario specification
- runnable scenario files under `http/scenario/`

Typical outputs:
- `http/scenario/scenarios.json`
- `http/scenario/checks/*.sh`
- `http/scenario/scenarios/*.sh` when cleanup or final verify is needed

Do not change application code unless the user explicitly asks.

## What to inspect

Use these sources in order:

1. `docs/knowledgebase/jupyter.md`
2. the feature’s own docs/spec files
3. the codebase routes, controllers, services, message adapters, and repositories
4. related repos if needed for the full flow:
   - `../xihe-message-server`
   - `../xihe-sdk`
   - `../xihe-jupyter-server`
   - `../xihe-internal-server`
   - `../xihe-grpc-protocol`
   - `../deploy`
5. runtime environment details when live validation is needed:
   - `kubeconfig`
   - namespace
   - server deployment names
   - pod names from the correct deployment
   - service logs for each hop in the flow

If the live environment is available, confirm the flow with logs instead of guessing.

## Scenario design rules

Build the scenario from the full cloud path, not just one endpoint.

Follow this chain when relevant:
- external HTTP API on `xihe-server`
- MQ publish/consume
- internal HTTP call through `xihe-sdk`
- `xihe-jupyter-server` pod/CRD operation
- gRPC callback to `xihe-internal-server`
- shared storage update
- final HTTP or websocket verification on `xihe-server`

For each scenario, include:
- setup
- request
- variable extraction
- expected response checks
- log checks when useful
- cleanup

Prefer explicit checks for:
- request body shape
- response status and code fields
- extracted IDs such as `cloud_id` and `pod_id`
- pod lifecycle state
- callback completion

## Environment discovery

When live validation is needed, discover these in order:

1. kubeconfig path
2. namespace
3. server deployment
4. message-server deployment
5. jupyter-server deployment
6. internal-server deployment
7. pod names for each deployment

Use the correct pod from the correct deployment. Do not reuse an unrelated pod just because it is running in the same namespace.

Local HTTP tests should use port `8092` unless the user says otherwise.

## Logs to inspect

Inspect logs across the flow when they help confirm behavior:
- `xihe-server`
- `xihe-message-server`
- `xihe-jupyter-server`
- `xihe-internal-server`

Use logs to confirm:
- message publish/consume
- internal API calls
- pod create/release
- gRPC callbacks
- state updates

## Output style

When generating scenarios, write the files so another agent can run them directly.

The scenario spec should be clear enough to answer:
- what is being tested
- which service is the source of truth for each step
- what must succeed before the next step runs
- what state is considered success

If the user asks for one feature, stay focused on that feature, but still trace the whole flow before writing the scenario.

## Working method

1. Read the feature doc and code.
2. Map the end-to-end data flow.
3. Identify the real services, deployments, and pods.
4. Define scenario steps and checks.
5. Write runnable files.
6. Verify against logs or the live environment when possible.

If the flow is ambiguous, prefer the code and logs over the doc, and prefer the doc over guessing.
