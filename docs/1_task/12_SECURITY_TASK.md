# SECURITY_TASK.md

## Goal
Perform a full security-focused review of the codebase and implement the required hardening work without breaking business behavior.

This task is not limited to one package. It must review the application end-to-end: API layer, service layer, repository/database layer, configuration, tests, and any supporting utilities.

## Required reading order for every agent session
Follow this exact flow before making changes:

1. Read `docs/AGENT_README_FIRST.md`
2. Read `docs/DOCS_FOLDER_GUIDE.md`
3. Read the most relevant documents in `docs/` for the security change
4. Read the affected code, schema, and business model files
5. Read the user task again
6. Think
7. Plan with TODO
8. Read again
9. Code

Do not re-read `docs/1_task` after the first pass. Treat it as read-once input only.

## Security scope
The implementation must harden the application against the following classes of issues:

- rate limiting for abusive or repeated requests
- SQL injection prevention across all database access paths
- race conditions, especially around schedule creation and updates
- unsafe transaction handling
- missing validation and authorization gaps
- accidental data leakage through logs or errors
- missing security regression tests

## Primary work items

### 1. Codebase-wide security review
Audit the codebase for:
- raw SQL string concatenation
- unsafe query building
- unbounded request handlers
- missing input validation
- missing transaction boundaries
- duplicate schedule writes
- shared mutable state without protection
- weak error handling that could expose internals

### 2. Rate limiting
Add rate limiting at the API boundary.

The implementation should:
- protect high-risk endpoints first
- return a clear throttling response
- be configurable
- avoid breaking legitimate traffic patterns
- be covered by tests

### 3. SQL injection prevention
Ensure all database access uses safe patterns.

Requirements:
- use parameterized queries everywhere
- avoid string interpolation in SQL
- review repository functions for query construction
- add tests that prove malicious input is not executed as SQL
- fail closed if a query path is not safe

### 4. Race condition prevention for schedules
This is a major risk area.

The schedule flow must be reviewed for:
- duplicate creation under concurrent requests
- lost updates
- inconsistent reads
- non-atomic validation followed by insert/update
- unsafe uniqueness assumptions done only in application code

Preferred protections:
- database constraints
- transactions
- row locking when needed
- idempotent create/update behavior where appropriate
- concurrency tests that simulate simultaneous schedule requests

### 5. Tests
Testing must prove the security fixes are real, not just documented.

At minimum, add or update tests for:
- security-focused unit tests for pure logic
- repository tests for safe SQL behavior
- integration tests for schedule persistence flows
- concurrency tests for race-sensitive schedule operations
- API tests for rate limiting and request rejection
- regression tests for any issue discovered during the audit

Follow the project testing expectations in `docs/TESTING.md`.

## Documentation rules
Any change that affects security behavior, operational handling, or engineering rules must be reflected in the correct `docs/` file instead of only being described in code comments.

Use the docs folder as the source of truth.

## Execution plan
The implementation should follow this pattern:

- read first
- think
- plan with TODO
- read again
- code

Use that flow before touching code, and repeat it when the scope changes.

## Acceptance criteria
The task is complete only when all of the following are true:

- rate limiting exists and is tested
- SQL injection vectors are removed or blocked
- schedule creation/update is safe under concurrency
- the relevant business/core paths are still working
- tests cover the security changes
- no undocumented security rule is introduced
- code review notes explain what was audited and what changed

## Output expected from the agent
When the work is finished, the agent should provide:
- a summary of the security findings
- the files changed
- the tests executed
- any remaining risks or follow-up items

## Notes
Do not treat this as a small patch. This is a whole-codebase security review with implementation and regression testing.
