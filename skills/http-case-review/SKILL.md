---
name: http-testcase-review
description: Use when reviewing HTTP testcase artifacts under http/scenario for weak oracles, missing business-path coverage, auth gaps, brittle scenario design, or unclear user-journey verification.
---

# HTTP Testcase Review

## Overview

Reviews HTTP testcase artifacts under `http/scenario/**` using parallel reviewer subagents.

This skill is for **testcase quality**, not application code review and not scenario execution logic review. It focuses on whether the cases prove the intended behavior with strong assertions, realistic flows, and safe coverage.

## Scope

Review only files under `http/scenario/**`, especially:

- `http/scenario/scenarios.json`
- `http/scenario/checks/*.sh`
- `http/scenario/scenarios/*.sh`

Never review files outside `http/scenario/**` as review targets.

If external files are needed for context, use them only to understand the testcase contract, not as subjects of review findings.

Do **not** spend time on shell style or runtime plumbing unless it weakens testcase meaning.

## When to Use

- Reviewing generated testcase files from `http-scenario-generator`
- Reviewing hand-written API scenario cases before using them
- Checking whether a scenario actually proves security, reliability, business logic, and user-visible behavior
- Comparing multiple scenario files for missing negative coverage or weak checks

## Reviewer Team

Always dispatch these reviewers in parallel:

| Reviewer | Focus |
|---|---|
| `http-testcase-security-reviewer` | Auth coverage, deny/allow symmetry, secret leakage in logs, tenant/user boundary mistakes |
| `http-testcase-reliability-reviewer` | Flaky setup, nondeterministic fixture choice, brittle waits, weak cleanup, environment coupling |
| `http-testcase-business-logic-reviewer` | Whether the scenario proves the intended business journey, state transitions, and setup/output correlation |
| `http-testcase-user-experience-reviewer` | Whether the testcase reflects what a user or client actually experiences: status, error semantics, visible outcomes, understandable scenario flow |
| `http-testcase-testing-reviewer` | Oracle strength, false positives, missing negative/boundary/idempotency coverage, determinism, and assertion quality |

Reviewer rule files live under `rules/` in this skill directory. Each subagent should read `rules/shared-review-rules.md` plus its own role file before reviewing.

## Severity Scale

| Level | Meaning |
|---|---|
| `P0` | Testcase would certify a critical security or business behavior incorrectly |
| `P1` | High-confidence gap in auth, lifecycle, or expected outcome verification |
| `P2` | Meaningful weakness that can hide regressions or create flaky/confusing coverage |
| `P3` | Minor clarity, consistency, or maintainability improvement |

## Review Principles

Review testcase artifacts as **behavior proofs**, not as scripts.

For this skill, **testcase quality** means:

- strong oracles for transport status and business outcome
- deterministic setup and fixture selection
- complete enough branch coverage for the claimed scenario
- correct correlation between setup inputs and observed outputs
- explicit verification of side effects and terminal state

Every reviewer should check for these common failure modes:

- success is inferred from `.code == ""` without checking HTTP status or required business fields
- empty result sets still pass, creating false positives
- setup chooses unstable fixtures like `.data[0]`
- cleanup exists but does not prove resource state changed correctly
- auth is tested only for deny or only for allow, not both where the contract needs both
- scenario extracts IDs or URLs but never uses them to prove outcome
- checks verify JSON shape but not business meaning
- scenario logs sensitive values that should not be printed broadly
- scenario description claims more than the steps actually verify

## What Good Findings Look Like

Each finding must be actionable and testcase-specific.

Good findings usually include:

- exact file and line
- what behavior is not actually being proven
- why the testcase can pass incorrectly
- the missing assertion or missing scenario branch

Bad findings to avoid:

- generic shell style comments
- abstract "improve coverage" advice without a concrete missing case
- runtime implementation complaints about `run.sh` or HTTP helpers outside scope

## Review Workflow

### Stage 1: Determine Scope

Default scope is all relevant files under `http/scenario/**`, whether tracked or untracked.

If the user gives a narrower target such as a scenario name or file path, limit review to testcase files relevant to that target.

Read only the testcase artifacts needed to understand:

- scenario names and descriptions
- step order and extracted variables
- check scripts
- cleanup/final verification scripts inside `http/scenario/`

### Stage 2: Build Intent Summary

Write a short intent summary from the testcase files themselves:

```text
Intent: Verify cloud pod lifecycle and access control using scenario-driven HTTP and websocket checks.
Must prove lifecycle transitions, access restrictions, and observable success/failure states.
```

### Stage 3: Dispatch Reviewers

Dispatch the five reviewers in parallel.

Each reviewer gets:

1. the scope
2. the intent summary
3. the relevant testcase files
4. `rules/shared-review-rules.md`
5. its own role file under `rules/`
6. the shared output schema below

Each reviewer returns compact JSON:

```json
{
  "reviewer": "http-testcase-testing-reviewer",
  "findings": [
    {
      "title": "Pagination testcase passes with empty result set",
      "severity": "P2",
      "file": "http/scenario/checks/check_page_size_5.sh",
      "line": 11,
      "confidence": 0.91,
      "autofix_class": "manual",
      "owner": "downstream-resolver",
      "requires_verification": true,
      "pre_existing": false,
      "suggested_fix": "Assert at least one row exists before validating page_size and confirm the returned page size matches the scenario input.",
      "why_it_matters": "The testcase can pass even when the endpoint returns no matching records, so it does not prove pagination behavior.",
      "evidence": [
        "The script checks page_size but does not fail on an empty data array."
      ]
    }
  ],
  "dimension_status": {
    "security": "clear",
    "reliability": "findings",
    "business_logic": "findings",
    "user_experience": "clear",
    "testcase_quality": "findings"
  },
  "residual_risks": [],
  "testing_gaps": []
}
```

## Review Rule Files

Subagents must read:

- `rules/shared-review-rules.md`
- `rules/security-reviewer.md`
- `rules/reliability-reviewer.md`
- `rules/business-logic-reviewer.md`
- `rules/user-experience-reviewer.md`
- `rules/testing-reviewer.md`

Each reviewer reads the shared rules plus its own role file.

## Merge Rules

After reviewer results return:

1. drop malformed findings
2. suppress findings below `0.60` confidence unless `P0`
3. deduplicate by file + nearby line + normalized title
4. resolve conflicts with evidence first, severity second:
   - if reviewers disagree on whether a finding is real, keep it only when at least one reviewer provides concrete testcase evidence
   - do not promote a weakly evidenced finding to `P0` or `P1` only because another reviewer was cautious
   - when reviewers agree on the issue but disagree on urgency, keep the higher severity only if the evidence supports user-visible or security-significant failure
5. group findings by severity
6. keep residual risks and testing gaps as separate sections
7. include a coverage summary for each review dimension: security, reliability, business logic, user experience, testcase quality

## Output Format

Present findings first, ordered by severity.

Use this table format:

| # | File | Issue | Reviewer(s) | Confidence | Route |
|---|---|---|---|---|---|

After findings, include:

- `Coverage Summary`
- `Residual Risks`
- `Testing Gaps`
- `Verdict`

## Verdict Rules

- `Ready` when no findings remain and all five review dimensions were explicitly assessed
- `Ready with fixes` when all five review dimensions were explicitly assessed and only P2/P3 findings remain
- `Not ready` when any P0/P1 finding remains
- `Not ready` when one of the five review dimensions was not actually covered in the final report

## Reviewer Prompt Template

Use this exact structure when spawning each reviewer subagent:

```text
Review HTTP testcase artifacts under http/scenario/** only.

Role: {reviewer_name}
Focus: {reviewer_focus}

Read these rule files before reviewing:
- {skill_dir}/rules/shared-review-rules.md
- {role_rule_file}

Intent:
{intent_summary}

Scope files:
{file_list}

Rules:
- Follow the shared and role-specific rule files.
- Return compact JSON with findings, residual_risks, testing_gaps, and a dimension_status object for security, reliability, business_logic, user_experience, and testcase_quality.
```

## Quality Gate

Before finalizing the review, verify:

- findings are testcase-specific
- line numbers point to testcase artifacts in scope
- no findings are just shell-style nits
- no findings depend on execution logic outside `http/scenario/**`
- the final report explicitly marks security, reliability, business logic, user experience, and testcase quality as either `clear` or `findings`
