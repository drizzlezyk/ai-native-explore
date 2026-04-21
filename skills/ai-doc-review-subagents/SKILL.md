---
name: "ai-doc-review-subagents"
description: "Reviews Markdown design docs with parallel sub-agents and outputs a consolidated report. Invoke when user asks to review requirement or architecture documents for clarity, logic, solution quality, or safety."
---

# AI Document Review Sub-Agents

Review Markdown documents with a coordinator plus five parallel sub-agents:

- `TermStyle`: wording clarity, sentence fluency, terminology consistency
- `LogicFlow`: cross-section consistency, assumptions, data and conclusion alignment
- `SolutionReview`: technical soundness, trade-offs, feasibility, scalability, operability
- `SafeGuard`: security, compliance, privacy, credentials, risky defaults
- `RuleCompliance`: repository-specific documentation rule checks based on embedded workspace review rules

Use this skill when:

- the user asks to review a requirements or architecture document
- the user wants a structured review report for a document owner
- the user wants multi-angle review with parallel reviewer roles
- the user wants a quality score, key issues, and improvement suggestions

Do not use this skill for:

- pure copy-editing with no review intent
- code review of source files
- generating a new design document from scratch

## Workspace-Aware Baseline

Always load repository context before reviewing:

1. Read `AGENTS.md`.
2. Read the target document.
3. Read the matching template as the baseline:
   - `templates/Requirement Analysis/#1 Requirement Analysis Specification.md`
   - `templates/Architecture Desgin/#1 Architecture Design Specification.md`
4. Apply the embedded workspace review rules in this skill, especially for Requirement Analysis and Architecture Design checks.
5. Read relevant rules from `context/team/` when the document involves security, API, credentials, privacy, or compliance.
6. Read relevant writing experience from `context/experience/` when useful for spotting common documentation issues.

For this repository, pay special attention to:

- Requirement Analysis documents must stay concise and should not over-design.
- Acceptance criteria must be measurable and testable.
- Task breakdown should usually stay within 2-4 atomic tasks.
- Relevance analysis checkboxes must include reasons when checked.
- Value evaluation must end with `Accept`, `Reject`, or `Pending`.
- Architecture document directory naming keeps the historical spelling `Architecture Desgin`.
- If `need_security` applies, security design sections must be substantively filled in.

## Coordinator Workflow

### Step 1: Identify review scope

Determine:

- document type: Requirement Analysis or Architecture Design
- document path
- optional owner or audience
- whether the review is template compliance only, design quality only, or full review

If the user does not specify a scope, default to full review.

### Step 2: Build a normalized fact sheet

Before launching sub-agents, extract a compact fact sheet from the document:

- document purpose
- stated goals
- acceptance criteria or design goals
- key metrics and assumptions
- components, interfaces, and data flows
- security-sensitive content
- unresolved placeholders such as `[TODO]`

This fact sheet is shared by all sub-agents to keep findings comparable.

### Step 3: Run five sub-agents in parallel

Each sub-agent reviews independently and produces findings in the same schema:

```text
[AgentName]
Score: x.x/5.0
Severity: critical | major | minor
Section: <document section>
Finding: <what is wrong>
Why it matters: <risk or consequence>
Evidence: <quote or summary from the doc>
Recommendation: <specific fix>
```

### Step 4: Aggregate and de-duplicate

The coordinator merges overlapping findings using these rules:

- keep the most concrete finding when multiple agents flag the same issue
- preserve the highest severity among duplicates
- combine style-only duplicates into a single normalization suggestion
- distinguish document defects from missing upstream business input

### Step 5: Produce the final report

The final report must be concise, actionable, and suitable for the current document owner.

## Sub-Agent Playbooks

### 1. TermStyle

Mission:

- improve readability without nitpicking
- detect inconsistent terminology, abbreviations, naming, and tone
- identify sentences that are hard to parse or potentially ambiguous

Typical checks:

- same concept written with multiple names such as `AI`, `AIGC`, `artificial intelligence`
- role names changing across sections
- metric names not defined on first use
- long sentences with unclear subject or action
- vague statements such as "greatly improve", "quickly", "better experience" without a concrete qualifier

Focus questions:

- Are key terms used consistently across title, sections, tables, and diagrams?
- Can a new reviewer understand each sentence on first read?
- Are placeholders, examples, and final content clearly separated?

Output preference:

- group similar wording problems into one normalization item
- avoid reporting trivial punctuation-only issues unless they impair meaning

### 2. LogicFlow

Mission:

- verify that the document tells one coherent story end to end
- detect contradictions between assumptions, scope, tasks, labels, and conclusions

Typical checks:

- earlier sections claim a large scale, later sections assume tiny traffic
- acceptance criteria do not match the described scenario
- task list cannot deliver the stated goals
- relevance analysis checkboxes conflict with the summary result
- requirement analysis concludes `need_light` even though earlier boxes indicate security or design impact
- architecture sections say "not involved" while later sections depend on that capability

Repository-specific checks for Requirement Analysis:

- measurable acceptance criteria exist
- task count and granularity stay reasonable
- relevance reasons are present when needed
- summary labels align with all preceding checkbox sections
- value assessment has a clear conclusion and rationale

Repository-specific checks for Architecture Design:

- architecture, data flow, interfaces, and task breakdown align with the design goal
- sections marked "not involved" include a reason
- optional sections are not silently skipped when risk indicates they should exist

### 3. SolutionReview

Mission:

- evaluate whether the proposed solution is technically sound and appropriate
- challenge weak trade-offs, hidden complexity, and missing alternatives

Typical checks:

- component choice does not match durability, consistency, or latency needs
- no explanation of why a new dependency or middleware is justified
- insufficient rollback, observability, or operability planning
- scalability claims without capacity assumptions
- lack of failure handling, idempotency, retries, or degradation design where needed
- a requirement document contains too much low-level design detail and loses decision focus

Focus questions:

- Is the chosen approach the simplest option that can satisfy the goal?
- Are the trade-offs explicit?
- Are there obvious missing considerations such as cost, resilience, observability, or migration path?
- Would another standard component be safer or cheaper?

When making suggestions:

- prefer concrete alternatives
- explain trade-offs, not just preference
- separate "must fix" design flaws from "could improve" design optimizations

### 4. SafeGuard

Mission:

- identify security, privacy, compliance, and operational risk in the document

Typical checks:

- hardcoded credentials, AK/SK, tokens, passwords, or unsafe secret handling
- missing authn/authz design around new APIs or privileged operations
- privacy-sensitive data collected without minimization, masking, retention, or access control
- third-party dependencies or binaries introduced without supply chain consideration
- security-sensitive changes not reflected in relevance analysis
- data flow crosses trust boundaries without encryption, validation, or audit discussion

Repository-specific checks:

- if the content implies `need_security`, verify the document explicitly reflects that impact
- architecture documents with security implications should include threat analysis and security controls
- suggestions should align with `context/team/` security practices

Critical issue examples:

- secrets written directly into docs or design examples as fixed values
- public exposure or permission expansion with no authentication or authorization design
- sensitive personal data processing with no privacy controls

### 5. RuleCompliance

Mission:

- verify the document against the embedded repository review rules in this skill
- catch template, naming, gating, and cross-document compliance issues that generic reviewers may miss

Embedded workspace review rules:

- Skill and command files should have complete and unambiguous steps, clear parameter descriptions, and remain consistent with `AGENTS.md`.
- Repository docs should keep structure aligned with actual repository layout and should not reference missing files or conflicting rules.
- Requirement Analysis docs must satisfy all of the following:
  - filename starts with `#{issueId}`
  - acceptance criteria are measurable and testable
  - task breakdown stays compact, usually 2-4 tasks
  - checked relevance items include reasons where required
  - value evaluation ends with `Accept`, `Reject`, or `Pending`
  - template guide blocks and checkbox lists are preserved
  - label logic is internally consistent:
    - section A/B/C selections align with 5.1 summary
    - any A-section security trigger maps to `need_security`
    - if A triggers security design, B first item is checked accordingly
    - C-section narrative and checkbox selections do not contradict each other
    - `need_light` is checked only when A/B/C are all unchecked
- Architecture Design docs must satisfy all of the following:
  - filename starts with `#{issueId}`
  - directory spelling remains `Architecture Desgin`
  - if `need_security` applies, section 3.1 includes threat analysis, security design implementation, and security task breakdown
  - sections that do not apply are retained with an explicit reason instead of being deleted
  - template guide blocks are preserved
- Test Strategy and Test Report docs should not keep template placeholders in title or basic info, and their scope should match Requirement Analysis labels.
- Release docs should have actual change level, execution steps, and rollback plan instead of template placeholders.
- Experience docs should be specific and actionable, ideally showing problem, cause, and solution with concrete examples.
- Cross-document consistency must hold:
  - Requirement Analysis labels match whether Architecture/Test directories should exist
  - deliverable references point to real files when verifiable
  - the same fact, scope, and conclusion should not conflict across issue documents

Typical checks for Requirement Analysis:
- checked relevance items include reasons where required
- template guide blocks and checkbox lists are preserved
- label logic is internally consistent:
  - section A/B/C selections align with 5.1 summary
  - any A-section security trigger maps to `need_security`
  - if A triggers security design, B first item is checked accordingly
  - `need_light` is checked only when A/B/C are all unchecked
- value evaluation ends with `Accept`, `Reject`, or `Pending`

Typical checks for Architecture Design:

- filename starts with `#{issueId}`
- directory spelling remains `Architecture Desgin`
- if `need_security` applies, section 3.1 is substantively present
- sections that do not apply are retained with an explicit reason instead of being deleted
- template guide blocks are preserved where the repository expects them

Cross-document checks:

- Requirement Analysis labels match whether Architecture/Test directories should exist
- document conclusions do not conflict across issue deliverables
- references to files, stages, and deliverables point to real repository paths when verifiable

Output preference:

- cite the embedded rule category name when possible
- treat rule violations as compliance findings, not style suggestions
- escalate missing gate-triggered sections above generic wording defects

## Scoring Model

Each sub-agent gives a score from `1.0` to `5.0`:

- `5.0`: strong and complete, no material concerns
- `4.0`: solid, only minor gaps
- `3.0`: workable but has notable weaknesses
- `2.0`: risky or incomplete
- `1.0`: fundamentally flawed

Recommended weights:

- `TermStyle`: 20%
- `LogicFlow`: 25%
- `SolutionReview`: 25%
- `SafeGuard`: 15%
- `RuleCompliance`: 15%

Overall score:

```text
overall = weighted average of five agent scores
```

Apply severity overrides:

- any unresolved critical security, logic, or rule-compliance flaw caps the final score at `3.4`
- two or more critical issues cap the final score at `2.9`

## Severity Standard

- `critical`: likely causes wrong design direction, compliance breach, or major delivery risk
- `major`: important gap or inconsistency that should be fixed before approval
- `minor`: improvement suggestion or low-risk wording issue

Do not inflate severity for purely stylistic feedback.

## Final Report Format

Output in the user's language. For this workspace, Chinese is usually the best default unless the user asks otherwise.

Use this structure:

```markdown
## Document Review Report

- Document: `<path>`
- Review scope: `full review`
- Overall score: `3.2 / 5.0`
- Conclusion: `needs revision before approval`

### Executive Summary
- Critical issues (2): database selection lacks sharding scalability analysis; no cache penetration protection is described.
- Improvement suggestions (3): add canary release strategy; define service degradation thresholds; evaluate total cost of the new component.

### Rule Compliance Summary
- Compliance status: `partially compliant`
- Violated rules (2): missing template guide blocks; gate-triggered security section is absent.
- Rule sources: `embedded workspace review rules`, `AGENTS.md`

### Findings by Agent
- TermStyle: score `3.8/5.0`
  - [minor] terminology switches between `AI` and `artificial intelligence` without normalization.
- LogicFlow: score `2.9/5.0`
  - [major] the document claims `QPS 1000` in goals but later capacity assumptions are based on `QPS 10`.
- SolutionReview: score `3.0/5.0`
  - [critical] the persistence strategy is weak for message durability requirements; evaluate a message queue designed for durable delivery.
- SafeGuard: score `3.1/5.0`
  - [critical] credential handling implies fixed AK/SK in configuration, which is non-compliant.
- RuleCompliance: score `2.8/5.0`
  - [major] the document deletes template-required sections instead of keeping them with an explicit "not involved" reason, which violates repository review rules.

### Violated Rules
- `embedded workspace review rules` - template guide blocks and checkbox lists must be preserved.
- `embedded workspace review rules` - if `need_security` applies, section 3.1 must be filled.
- `AGENTS.md` - checked relevance analysis items must include reasons where required.

### Action List
- Must fix: <items>
- Should improve: <items>
- Nice to have: <items>
```

## Review Behavior Rules

- Prioritize correctness, consistency, and risk over personal preference.
- Cite evidence from the document, not vague impressions.
- Keep findings actionable and specific.
- Do not overload the author with dozens of tiny style notes.
- When repository rules and generic best practices differ, follow the embedded workspace review rules and `AGENTS.md` first.
- Always surface rule violations in a dedicated section instead of burying them inside generic findings.
- When a rule is violated, cite the rule source and summarize the expected behavior in one sentence.
- If a section is intentionally out of scope, say so explicitly.
- If the document is still a template draft with many `[TODO]` markers, separate "draft incomplete" from "design incorrect".
- If the user asks for a quick pass, return only critical and major findings plus the score.

## Invocation Examples

- "Review this Requirement Analysis document with sub-agents."
- "Audit this Architecture Design document for logic, solution quality, and security."
- "Give me a consolidated review report for the current Markdown design doc."
- "Check whether this document is clear, internally consistent, and safe to approve."
