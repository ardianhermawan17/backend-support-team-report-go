# CHANGE_TO_SOFT_DELETE_DATABASE.md

## Purpose
This document defines the workflow for changing the database design to support **soft delete** across the application.

It applies to updates to the database schema, ERD, and initialization database artifacts such as:

- `schema.dbml`
- `initial.db`

The goal is to keep the database design auditable, safe to evolve, and aligned with the application architecture.

## Scope
This instruction focuses on database changes related to soft delete behavior.

It may include:

- adding `deleted_at` or equivalent soft-delete fields
- updating unique constraints and indexes
- adjusting relationships or audit behavior
- changing table lifecycle rules
- updating schema documentation and initialization data

This instruction does **not** include task-specific domain files under `docs/1_task/`.
Those files are excluded because they are reserved for task-specific domain instructions.

## Single Source of Truth
Before making any change, the agent must read the documentation first.

The first file to read is:

`docs/AGENT_README_FIRST.md`

This is the gateway for the first season of agentic code and must always be treated as the starting point.

After that, the agent must read the rest of the relevant documentation under `docs/`, especially:

- `docs/architecture/`
- database-related instructions
- schema and ERD documentation

The agent must also read `schema.dbml` to understand the current database structure clearly before making any change.

## Mandatory Reading Order
For every soft-delete database change, the agent must follow this exact sequence:

### 1) Read first
Read the documentation before doing anything else.

Priority reading order:

1. `docs/AGENT_README_FIRST.md`
2. relevant docs under `docs/`
3. `docs/architecture/`
4. `schema.dbml`
5. `initial.db`

### 2) Think
After reading, the agent must analyze:

- which tables should support soft delete
- how soft delete affects business rules
- how queries and uniqueness constraints will behave after deletion
- how audit and recovery behavior should work
- whether the ERD must be updated

### 3) Plan with TODO
Before editing files, the agent must write a short TODO plan.

The plan should cover:

- which schema fields will change
- which constraints need revision
- which tables or relationships may need adjustment
- which files will be updated
- what must remain unchanged

### 4) Read again
After planning, the agent must read the relevant documentation again to verify the plan.

This second read is required so the change stays aligned with the documented architecture and database rules.

### 5) Code
Only after the second read may the agent update the database design artifacts.

## Soft Delete Design Rules
When changing the schema to soft delete, the agent should apply the following principles:

- prefer explicit soft delete fields such as `deleted_at` or equivalent
- keep delete history auditable
- avoid hard deletion unless the docs explicitly require it
- make uniqueness constraints compatible with soft delete behavior
- keep foreign key relationships consistent with the new lifecycle rules
- update the ERD and initialization database together when the model changes

## File Update Rules
The agent is allowed to change the documentation and database design files needed to reflect the new database model.

At minimum, the agent may update:

- `schema.dbml`
- `initial.db`
- related docs under `docs/` if the database model changes

If the database and ERD design changes, the documentation must be updated so the new model remains the source of truth.

## Architecture Awareness
The agent must use `docs/architecture/` as the architectural reference before changing any schema.

This is required to ensure the soft-delete design remains compatible with:

- bounded contexts
- DDD boundaries
- CQRS read/write behavior
- audit requirements
- module ownership
- data integrity rules

## Implementation Discipline
When making a soft-delete database change:

- do not change the schema without reading the docs first
- do not skip the planning step
- do not skip the second read
- do not alter files outside the agreed scope
- do not break auditability
- do not violate schema consistency between `schema.dbml` and `initial.db`

## Output Expectation
The result of this workflow should be a database design that is:

- soft-delete aware
- auditable
- consistent
- well documented
- aligned with architecture
- safe for future application growth

## Final Rule
For soft-delete database changes, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md`, `docs/architecture/`, and `schema.dbml`, must be treated as the source of truth before any database change is made.

