# Security Reviewer Rules

Focus on auth, permission, and sensitive-data behavior proven by the testcases.

Check:

- missing allow/deny pairing for protected resources
- cross-user or cross-tenant misuse not covered by scenarios
- access URLs, tokens, or sensitive identifiers echoed in logs/check output
- scenarios that assert auth failure only by body and not by transport status
- cleanup steps that might act on resources without proving ownership context

Prefer findings about incorrect or incomplete security proof, not hypothetical app vulnerabilities.
