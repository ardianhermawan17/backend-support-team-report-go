# AUTH_TASK.md

## Purpose
This document defines the workflow for implementing the **authentication** domain of the application.

It exists to guide the agent through the correct sequence of reading, understanding, planning, and coding before any authentication-related change is made.

## Scope
This document focuses only on the **authentication domain**.

It applies to work such as:

- login and logout
- session or token handling
- password hashing and verification
- authentication guards and access control entrypoints
- identity-related application flow

This document does **not** cover task-specific domain files inside `docs/1_task/`.
Those files are excluded from this instruction because they are reserved for task-specific domain rules.

## Single Source of Truth
Before writing any authentication code, the agent must treat the documentation as the source of truth.

The agent must read:

1. `docs/AGENT_README_FIRST.md`
2. the relevant files under `docs/`
3. `docs/architecture/`
4. `schema.dbml`

The agent must always start from `docs/AGENT_README_FIRST.md` for the first season of agentic code.

## Mandatory Reading Order
For every new authentication task, the agent must follow this order:

### 1) Read first
Read the documentation first, before any implementation work.

Focus on:

- `docs/AGENT_README_FIRST.md`
- the docs that explain application-wide rules
- `docs/architecture/` for architectural direction
- `schema.dbml` for database clarity

### 2) Think
After reading, the agent must reason about:

- the domain boundaries
- the current schema constraints
- the security implications
- the expected authentication flow
- how the module fits into the wider system

### 3) Plan with TODO
Before coding, the agent must create a concise TODO plan.

The plan should include:

- what will be created
- what will be changed
- what will be validated
- what will be left untouched

### 4) Read again
After planning, the agent must read the relevant docs again to confirm the plan still matches the documented rules.

This second read is required so the implementation does not drift from the documented source of truth.

### 5) Code
Only after the second read may the agent begin coding.

## Authentication Design Rules
Authentication work must follow the system’s architecture and schema constraints.

The agent must ensure:

- authentication logic stays inside the auth domain boundaries
- API handlers remain thin and do not hold business logic
- application logic is separated from domain logic
- password data is handled securely
- access tokens or session logic are implemented consistently with the documented architecture
- database writes are safe and auditable

## Schema Awareness
The agent must inspect `schema.dbml` before making authentication changes.

This is required so the implementation aligns with:

- table ownership
- relationships
- constraints
- audit behavior
- existing user/company linkage

If the schema does not support the needed auth behavior, the agent must identify the mismatch before coding.

## Architecture Awareness
The agent must read `docs/architecture/` before coding authentication logic.

This is required to keep the implementation aligned with:

- DDD boundaries
- CQRS direction
- module layout
- dependency direction
- service interaction patterns
- security and concurrency expectations

## Implementation Discipline
When implementing authentication:

- do not bypass the documented structure
- do not invent new rules without checking the docs
- do not mix auth concerns with unrelated domains
- do not skip the read-first sequence
- do not skip the re-read step after planning

## Output Expectation
Authentication work should result in code that is:

- secure
- predictable
- consistent with the docs
- easy to audit
- easy to test
- easy to extend

## Final Rule
For authentication tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

No code should be written before the documentation has been read and re-checked.
