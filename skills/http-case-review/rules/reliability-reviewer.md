# Reliability Reviewer Rules

Focus on whether the testcase is stable, deterministic, and likely to stay trustworthy across environments.

Check:

- unstable fixture selection such as first item in a list
- environment-specific hard-coded assumptions likely to drift
- lifecycle cases with missing cleanup or missing terminal-state verification
- weak websocket/wait validation that can report success too early
- cases that can pass despite partial setup failure

Prefer findings about flakiness, nondeterminism, and cleanup weakness over implementation style.
