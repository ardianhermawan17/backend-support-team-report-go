# CODE_CONTRACT.md

## Purpose
This document defines how agentic AI must code for this project, what constraints must be respected, and what coding styles are preferred.

## Primary architecture preferences
Prefer:
- Domain-Driven Design
- Command Query Responsibility Segregation
- explicit domain boundaries
- small units of change
- business rules enforced close to the domain

## Required coding principles
- Keep business logic out of controllers when possible.
- Keep database access isolated from domain logic.
- Keep validation explicit and deterministic.
- Prefer pure functions where practical.
- Prefer composition over inheritance unless the domain requires it.
- Always preserve company tenancy boundaries.

## Concurrency and race conditions
The implementation must prevent race conditions for critical operations such as:
- creating players with the same jersey number in one team
- inserting match reports
- generating sequential business side effects
- updating ownership-sensitive records

Use database constraints, transactional boundaries, and atomic writes. Never rely only on in-memory checks.

## ID strategy
All application entities that require business IDs must use Snowflake-style `BIGINT` identifiers.
Do not replace them with UUIDs unless the architecture is explicitly redesigned.

## Data access rules
- Write operations must be transactional.
- Read models may be denormalized when it improves reporting.
- Unique constraints must exist in the database, not only in code.
- Foreign keys must be used where relational integrity matters.
- Soft delete is preferred over hard delete for auditable entities when appropriate.

## AI coding rules
Agentic AI must:
- read the docs in `docs/` before editing code
- preserve existing contracts unless the task explicitly changes them
- explain the impact of schema changes before implementing them
- keep changes minimal and traceable
- avoid speculative refactors
- avoid rewriting the whole system when a local change is enough

## Forbidden behavior
Do not:
- bypass business rules for convenience
- skip audit logging
- weaken tenancy rules
- remove constraints that protect integrity
- add hidden side effects without documentation
- introduce unnecessary abstractions
- code without updating the relevant docs

## Output expectations
Every coding action should be explainable in terms of:
- what changed
- why it changed
- which rule it follows
- which risk it reduces
