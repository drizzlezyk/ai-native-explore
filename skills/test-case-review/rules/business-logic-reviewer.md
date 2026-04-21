# Business Logic Reviewer Rules

Focus on whether the scenario proves the intended business journey.

Check:

- whether the scenario matches the business story claimed in its description
- setup/output correlation, such as subscribe inputs not being verified later
- lifecycle state transitions actually being asserted
- missing negative cases for invalid, duplicate, stale, or already-finished actions
- extracted variables that are never used to prove downstream behavior

Prefer findings about missing business proof, not generic API completeness.
