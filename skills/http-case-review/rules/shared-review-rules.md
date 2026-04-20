# Shared Review Rules

Review only HTTP testcase artifacts under `http/scenario/**`.

Hard boundaries:

- Never review files outside `http/scenario/**` as finding targets.
- If external files are needed for context, use them only to understand testcase intent, not as review subjects.
- Review testcase meaning, not application code or execution plumbing.
- Ignore shell style unless it weakens testcase meaning.

Review testcase artifacts as behavior proofs, not scripts.

For this skill, testcase quality means:

- strong oracles for transport status and business outcome
- deterministic setup and fixture selection
- complete enough branch coverage for the claimed scenario
- correct correlation between setup inputs and observed outputs
- explicit verification of side effects and terminal state

Common failure modes:

- success is inferred from `.code == ""` without checking HTTP status or required business fields
- empty result sets still pass, creating false positives
- setup chooses unstable fixtures like `.data[0]`
- cleanup exists but does not prove resource state changed correctly
- auth is tested only for deny or only for allow, not both where the contract needs both
- scenario extracts IDs or URLs but never uses them to prove outcome
- checks verify JSON shape but not business meaning
- scenario logs sensitive values that should not be printed broadly
- scenario description claims more than the steps actually verify

Every finding must be concrete and actionable.

Confidence should reflect how directly the testcase evidence supports the finding:

- higher confidence when the issue is visible directly in `http/scenario/**`
- higher confidence when the testcase can pass for the wrong reason or fail to prove the claimed behavior
- lower confidence when the concern is vague, depends on assumptions outside `http/scenario/**`, or lacks a precise file/line anchor
- do not use application code visibility as a confidence signal; this skill does not review application code as a target

Good findings include:

- exact file and line
- what behavior is not actually being proven
- why the testcase can pass incorrectly
- the missing assertion or missing scenario branch

Avoid:

- generic shell style comments
- abstract "improve coverage" advice without a concrete missing case
- complaints about runtime helpers or runner logic outside scope

Return compact JSON with:

- `reviewer`
- `findings`
- `dimension_status`
- `residual_risks`
- `testing_gaps`

Each finding should include:

- `title`
- `severity`
- `file`
- `line`
- `confidence`
- `autofix_class`
- `owner`
- `requires_verification`
- `pre_existing`
- `suggested_fix`
- `why_it_matters`
- `evidence`
