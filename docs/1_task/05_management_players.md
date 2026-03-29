# MANAGEMENT_PLAYERS.md

## Purpose
This document defines the workflow for implementing and changing the **players** domain of the application.

It is the task instruction for work related to player management, including team assignment, player lifecycle, player-level validation, and player data consistency.

## Scope
This instruction focuses only on the **players** domain.

It applies to work such as:

- create player
- view player
- edit player
- delete player
- assign player to a team
- enforce player-number uniqueness within a team
- player-level audit behavior
- player-related database changes
- player-related API and application logic

This document does **not** cover task-specific domain files under `docs/1_task/`.
Those files are excluded because they are reserved for task-specific domain rules.

## Business Focus
The players domain must support the rule that one player belongs to exactly one team, while one team can have many players.

The player record must support at least:

- name
- height
- weight
- position
- player_number

The allowed positions are:

- striker
- midfielder
- defender
- goalkeeper

The player number must be unique inside a single team.

## Single Source of Truth
Before making any player-related change, the agent must treat the documentation as the source of truth.

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
For every new players task, the agent must follow this exact sequence:

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

- the player domain boundaries
- the team ownership rule
- the current schema constraints
- player lifecycle requirements
- player-number uniqueness rules
- audit and deletion behavior
- how player data affects related modules such as teams and schedules

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

## Player Domain Rules
When implementing the players domain, the agent must respect these rules:

- each player belongs to exactly one team
- each team can have many players
- player CRUD must stay inside the team boundary
- player number must be unique within the team
- player changes must be auditable
- player identifiers must remain compatible with Snowflake-style BIGINT usage

## Schema Awareness
The agent must inspect `schema.dbml` before making player changes.

This is required so the implementation aligns with:

- table ownership
- foreign keys
- uniqueness rules
- lifecycle rules
- audit behavior
- team-to-player relationship

If the schema does not support the needed player behavior, the agent must identify the mismatch before coding.

## Architecture Awareness
The agent must read `docs/architecture/` before coding player logic.

This is required to keep the implementation aligned with:

- DDD boundaries
- CQRS direction
- module layout
- dependency direction
- service interaction patterns
- audit and reliability expectations

## Implementation Discipline
When implementing players work:

- do not bypass the documented structure
- do not invent new rules without checking the docs first
- do not mix player logic with unrelated domains
- do not skip the read-first sequence
- do not skip the re-read step after planning
- do not violate team ownership boundaries

## Expected Player Behavior
The player domain should support a clean administrator workflow where the company can manage its players safely and clearly through team ownership.

The expected behavior is:

- create player under a team
- view player within the correct team context
- update player information
- delete player according to lifecycle policy
- keep traceability for every important mutation

## Output Expectation
Player-related work should result in code and data definitions that are:

- secure
- consistent
- auditable
- easy to test
- easy to extend
- aligned with the app’s main goal

## Final Rule
For players tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md`, `docs/architecture/`, `schema.dbml`, and the main application goal document, must be treated as the source of truth before any players change is made.
