# MANAGEMENT_SCHEDULES.md

## Purpose
This document defines the workflow for implementing and changing the **schedules** domain of the application.

It is the task instruction for work related to match scheduling, including home/guest team pairing, schedule lifecycle, validation, and schedule data consistency.

## Scope
This instruction focuses only on the **schedules** domain.

It applies to work such as:

- create schedule
- view schedule
- edit schedule
- delete schedule
- assign home team and guest team
- validate match date and match time
- schedule-level audit behavior
- schedule-related database changes
- schedule-related API and application logic

This document does **not** cover task-specific domain files under `docs/1_task/`.
Those files are excluded because they are reserved for task-specific domain rules.

## Business Focus
The schedules domain must support the rule that the administrator company can create a match schedule for each team pairing.

The schedule record must support at least:

- match_date
- match_time
- home_team
- guest_team

This is a many-to-many style relationship between teams through match scheduling, because teams can appear in many schedules over time.

## Single Source of Truth
Before making any schedule-related change, the agent must treat the documentation as the source of truth.

The first file to read is:

`docs/AGENT_README_FIRST.md`

This is the gateway for the first season of agentic code and must always be the starting point.

After that, the agent must read the rest of the relevant documentation under `docs/`, especially:

- `docs/architecture/`
- the core app goal documentation
- database-related instructions
- API and code-contract instructions

The agent must also read `schema.dbml` to understand the current database structure clearly before making any change.

## Mandatory Reading Order
For every new schedules task, the agent must follow this exact sequence:

### 1) Read first
Read the documentation before doing anything else.

Priority reading order:

1. `docs/AGENT_README_FIRST.md`
2. relevant docs under `docs/`
3. `docs/architecture/`
4. `schema.dbml`
5. the main application goal document

### 2) Think
After reading, the agent must analyze:

- the schedule domain boundaries
- the team participation rule
- the current schema constraints
- match lifecycle requirements
- date and time validation rules
- audit and deletion behavior
- how schedules affect related modules such as teams and match reports

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

## Schedule Domain Rules
When implementing the schedules domain, the agent must respect these rules:

- each schedule must reference a home team and a guest team
- schedule CRUD must stay inside the company boundary
- schedule data must be consistent across schema, application, and API layers
- schedule changes must be auditable
- schedule identifiers must remain compatible with Snowflake-style BIGINT usage

## Schema Awareness
The agent must inspect `schema.dbml` before making schedule changes.

This is required so the implementation aligns with:

- table ownership
- foreign keys
- match relationships
- lifecycle rules
- audit behavior
- team-to-schedule relationship

If the schema does not support the needed schedule behavior, the agent must identify the mismatch before coding.

## Architecture Awareness
The agent must read `docs/architecture/` before coding schedule logic.

This is required to keep the implementation aligned with:

- DDD boundaries
- CQRS direction
- module layout
- dependency direction
- service interaction patterns
- audit and reliability expectations

## Implementation Discipline
When implementing schedules work:

- do not bypass the documented structure
- do not invent new rules without checking the docs first
- do not mix schedule logic with unrelated domains
- do not skip the read-first sequence
- do not skip the re-read step after planning
- do not violate team ownership boundaries

## Expected Schedule Behavior
The schedule domain should support a clean administrator workflow where the company can manage match schedules safely and clearly.

The expected behavior is:

- create schedule with home and guest teams
- view schedule in the correct company context
- update schedule information
- delete schedule according to lifecycle policy
- keep traceability for every important mutation

## Output Expectation
Schedule-related work should result in code and data definitions that are:

- secure
- consistent
- auditable
- easy to test
- easy to extend
- aligned with the app’s main goal

## Final Rule
For schedules tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md`, `docs/architecture/`, `schema.dbml`, and the main application goal document, must be treated as the source of truth before any schedules change is made.

