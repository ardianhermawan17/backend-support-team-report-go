# MANAGEMENT_TEAMS.md

## Purpose
This document defines the workflow for implementing and changing the **teams** domain of the application.

It is the task instruction for work related to team management, including administrator access, team lifecycle, and team-level data consistency.

The teams domain is part of the broader soccer-team-report application, where the company administrator manages teams, players, match schedules, and reports. The core model requires that a company can manage many teams, and each team belongs to exactly one company. fileciteturn0file0

## Scope
This instruction focuses only on the **teams** domain.

It applies to work such as:

- create team
- view team
- edit team
- delete team
- team-level validation
- team-level audit behavior
- team-related database changes
- team-related API and application logic

This document does **not** cover task-specific domain files under `docs/1_task/`.
Those files are excluded because they are reserved for task-specific domain rules.

## Business Focus
The teams domain must support the administrator’s ability to manage team information from the application.

The team record must support at least:

- name
- logo
- founded_year
- homebase_address
- cityof_homebase_address

The system must keep the business rule that one company can have many teams, but each team belongs to only one company. The core model also requires traceable, auditable, and scalable team management. fileciteturn0file0

## Single Source of Truth
Before making any team-related change, the agent must treat the documentation as the source of truth.

The first file to read is:

`docs/AGENT_README_FIRST.md`

This is the gateway for the first season of agentic code and must always be the starting point.

After that, the agent must read the rest of the relevant documentation under `docs/`, especially:

- `docs/architecture/`
- the core app goal documentation such as `MAIN_GOAL_APP.md`
- database-related instructions
- API and code-contract instructions

The agent must also read `schema.dbml` to understand the current database structure clearly before making any change.

## Mandatory Reading Order
For every new teams task, the agent must follow this exact sequence:

### 1) Read first
Read the documentation before doing anything else.

Priority reading order:

1. `docs/AGENT_README_FIRST.md`
2. relevant docs under `docs/`
3. `docs/architecture/`
4. `schema.dbml`
5. `MAIN_GOAL_APP.md`

### 2) Think
After reading, the agent must analyze:

- the team domain boundaries
- the company ownership rule
- the current schema constraints
- team lifecycle requirements
- audit and deletion behavior
- how team data affects related modules such as players and schedules

### 3) Plan with TODO
Before coding, the agent must create a short TODO plan.

The plan should include:

- what will be created
- what will be changed
- what will be validated
- what will be kept unchanged

### 4) Read again
After planning, the agent must read the relevant docs again to confirm the plan still matches the documented rules.

This second read is required so the implementation does not drift from the source of truth.

### 5) Code
Only after the second read may the agent begin coding.

## Team Domain Rules
When implementing the teams domain, the agent must respect these rules:

- a team belongs to exactly one company
- team CRUD must stay inside the company boundary
- team data must be consistent across schema, application, and API layers
- team deletion behavior must follow the documented database policy
- team changes must be auditable
- team identifiers must remain compatible with Snowflake-style BIGINT usage

## Schema Awareness
The agent must inspect `schema.dbml` before making team changes.

This is required so the implementation aligns with:

- table ownership
- foreign keys
- uniqueness rules
- soft delete or lifecycle rules if present
- audit behavior
- company-to-team relationship

If the schema does not support the needed team behavior, the agent must identify the mismatch before coding.

## Architecture Awareness
The agent must read `docs/architecture/` before coding team logic.

This is required to keep the implementation aligned with:

- DDD boundaries
- CQRS direction
- module layout
- dependency direction
- service interaction patterns
- audit and reliability expectations

## Implementation Discipline
When implementing teams work:

- do not bypass the documented structure
- do not invent new rules without checking the docs first
- do not mix team logic with unrelated domains
- do not skip the read-first sequence
- do not skip the re-read step after planning
- do not violate company ownership boundaries

## Expected Team Behavior
The team domain should support a clean administrator workflow where the company can manage its teams safely and clearly.

The expected behavior is:

- create team under a company
- view team within the same company context
- update team information
- delete team according to lifecycle policy
- keep traceability for every important mutation

## Output Expectation
Team-related work should result in code and data definitions that are:

- secure
- consistent
- auditable
- easy to test
- easy to extend
- aligned with the app’s main goal

## Final Rule
For teams tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md`, `docs/architecture/`, `schema.dbml`, and `MAIN_GOAL_APP.md`, must be treated as the source of truth before any teams change is made.
