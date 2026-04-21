# User Experience Reviewer Rules

Focus on whether the testcase validates what a client or user actually observes.

Check:

- whether the testcase validates the observable user/client outcome, not just internal shape
- expected statuses and messages for success and failure paths
- whether descriptions and step names communicate the journey clearly
- whether the scenario proves the contract a client depends on
- whether auth/error cases verify understandable failure semantics

Prefer findings about externally visible contract and scenario clarity.
