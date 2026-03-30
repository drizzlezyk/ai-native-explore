---
name: nocalhost-testing-workflow
description: Main orchestrator for complete nocalhost testing workflow
---

# Nocalhost Testing Workflow

This skill orchestrates the complete nocalhost testing workflow by coordinating test management, environment control, and test execution.

## Workflow

```
1. nocalhost-environment-control: initialize (setup testing environment via subagent)
2. nocalhost-test-management: generate (create test case templates)
3. nocalhost-test-execution: execute (run tests and generate reports)
4. nocalhost-test-management: refine (improve test cases based on results)
```

## Usage

Invoke this skill when you want to run the complete end-to-end testing workflow from test generation through execution and refinement.

Each step can also be invoked independently:
- `nocalhost-test-management` - Generate or refine test cases
- `nocalhost-environment-control` - Manage nocalhost environment
- `nocalhost-test-execution` - Execute tests and generate reports

## Example Complete Workflow

1. **Collect required variables**
   
   Follow the detailed rules in `preparecheck.md` to prepare enviroment

   **Run prepare with --help to see all options:**
   ```bash
   go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl prepare --help
   ```

   **IMPORTANT**: 
   - Verify ALL required fields are collected before proceeding
   - Verify generated scripts match the Dockerfile and server code
   - Missing any required field will cause the prepare command to fail

   Then dispatch a subagent to check the prepare is ok:

   ```javascript
   task({
     description: "check nocalhost testing environment is ok ",
     prompt: "check nocalhost .config.json according to .opencode/skills/nocalhost-testing/nocalhost-environment-control/preparecheck.md. Also verify startup.sh and build.sh are correctly generated.",
     subagent_type: "general"
   })
   ```

   **Then ask user to check the prepare is ok**


2. **Initialize environment**
   First prepare the environment with the collected variables:

   ```bash
   go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl prepare \
     --developer-name="$DEVELOPER_NAME" \
     --kubeconfig="$KUBECONFIG_PATH" \
     --namespace="$NAMESPACE" \
     --orig-deploy-name="$ORIG_DEPLOY_NAME"
   ```

   **Auto-derived parameters** (do NOT pass as flags - derive from codebase):
   - `binary-name`: from Dockerfile `go build -o <binary>`
   - `project-path`: current working directory or specified
   - `remote-port`: from Dockerfile EXPOSE or server code
   - `heartbeat-url`: from server code `/heartbeat` endpoint

   ```bash
   go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl status

   go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl oneclickstart
   ```
   Then dispatch a subagent to set up the testing environment using the task tool:

   ```javascript
   task({
     description: "Initialize nocalhost testing environment",
     prompt: "Use the nocalhost-environment-control skill to set up the testing environment. Run oneclickstart after prepare.",
     subagent_type: "general"
   })
   ```
 
3. Generate test cases
Use the nocalhost-test-management skill to create test case templates.

4. Execute tests
Run the test runner:

``` bash
go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go \
  --url=http://localhost:8092 \
  --group=cloud \
  --user=$DEVELOPER_NAME


go run -tags debug ./.opencode/skills/nocalhost-testing/nocalhost-test-execution/scripts/runner.go --help  
```

5. Refine test cases
Review the test report and update test cases as needed (manual process).

## Workflow Notes

- The workflow is iterative: after refining test cases, re-run steps 3-4
- Each step can be invoked independently for focused work
- **Always use skills for server control** - never run nocalhostctl scripts directly
- Reports persist in `tests/nocalhost-test-report/` for historical analysis

## Subagent Invocation Pattern

Step 2 uses the `task` tool to dispatch a subagent for environment control. This pattern ensures:

1. **Isolation**: Environment setup runs in a separate agent context
2. **Skill loading**: The subagent automatically loads the nocalhost-environment-control skill
3. **Error handling**: Subagent manages its own error recovery and logging
4. **Consistency**: Follows the same pattern used across other workflow skills

Example subagent invocation:
```javascript
task({
  description: "Initialize nocalhost testing environment",
  prompt: "Use the nocalhost-environment-control skill to set up the testing environment. This includes: up, rebuild, and forward operations.",
  subagent_type: "general"
})
```