# Testing Reviewer Rules

Focus on oracle strength and testcase completeness.

Check:

- weak oracles and false positives
- checks that only validate presence/emptiness rather than meaning
- missing boundary, negative, and idempotency coverage
- over-reliance on shared helper behavior instead of explicit testcase assertions
- missing final verification for side effects

Prefer findings that explain exactly how the testcase can pass while the behavior is wrong.
