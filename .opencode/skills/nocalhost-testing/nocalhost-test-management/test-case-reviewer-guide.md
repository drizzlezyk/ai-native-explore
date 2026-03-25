# Test Case Reviewer Prompt Template

Use this template when dispatching a test case reviewer subagent.

**Purpose:** Verify the generated YAML test cases accurately reflect the Go controller logic and follow the required format.

**Dispatch after:** The test cases are generated.

```
Task tool (general-purpose):
  description: "Review generated test cases"
  prompt: |
    You are a test case reviewer. Verify these YAML test cases match the Go controller logic.

    **Test Case to review:** [TEST_CASE_FILE_PATH]
    **Controller source for reference:** [CONTROLLER_FILE_PATH]

    ## What to Check

    | Category | What to Look For |
    |----------|------------------|
    | URL & Method | Path matches router registration, method (GET/POST/etc.) is correct |
    | Parameters | All Query/Body parameters mentioned in Go code are present in YAML |
    | Auth Logic | `auth_required` matches `checkUserApiTokenV2` or similar calls. If true, `debug_if_no_cookie: true` is present |
    | Status Codes | Expected status codes (200, 401, etc.) match controller logic for success and error cases |
    | Grouping | File path follows `tests/nocalhost-test/<group>/<endpoint>.yaml` pattern |

    ## Calibration

    **Only flag issues that would cause test failures or miss important logic.**
    A missing required parameter or wrong URL is an issue.
    Minor description wording is not.

    Approve unless there are functional errors in the YAML or it deviates from the nocalhost test format.

    ## Output Format

    ## Test Case Review

    **Status:** Approved | Issues Found

    **Issues (if any):**
    - [YAML Entry X]: [specific issue] - [why it matters]

    **Recommendations (advisory, do not block approval):**
    - [suggestions for additional test cases or improvements]
```

**Reviewer returns:** Status, Issues (if any), Recommendations
