# MANAGEMENT_REPORTS.md

## Purpose
This document defines the workflow for implementing and changing the **reports** domain of the application.

It is the task instruction for work related to match reporting, including final score tracking, match outcome classification, accumulated wins, and report-level audit behavior.

The core application goal is a soccer-team-report platform where the system stores match outcomes and reporting data, along with the rest of the company-managed soccer workflow. fileciteturn0file0

## Scope
This instruction focuses only on the **reports** domain.

It applies to work such as:

- create match report
- view match report
- edit match report
- delete match report
- calculate match result status
- store top scoring player for a match
- track accumulated wins for home and guest teams
- report-level audit behavior
- report-related database changes
- report-related API and application logic

This document does **not** cover task-specific domain files under `docs/1_task/`.
Those files are excluded because they are reserved for task-specific domain rules.

## Business Focus
The reports domain must support the post-match result of each scheduled match.

The report record must support at least:

- match_schedule
- home_team
- guest_team
- final_score
- status_match
- most_scoring_goal_player
- accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule
- accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule

The allowed match status values are:

- home_team_win
- guest_team_win
- draw

This domain exists to make the outcome of each match visible, auditable, and traceable inside the application’s main business flow. fileciteturn0file0

## Single Source of Truth
Before making any report-related change, the agent must treat the documentation as the source of truth.

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
For every new reports task, the agent must follow this exact sequence:

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

- the report domain boundaries
- the match-schedule dependency
- the team result rule
- the current schema constraints
- calculation and lifecycle requirements
- audit and deletion behavior
- how reports affect related modules such as schedules, teams, and players

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

## Report Domain Rules
When implementing the reports domain, the agent must respect these rules:

- each report must belong to one schedule
- report data must remain consistent with the schedule and the two teams involved
- status values must follow the documented match outcome set
- accumulated win counters must be computed and stored according to the business rule
- report changes must be auditable
- report identifiers must remain compatible with Snowflake-style BIGINT usage

## Schema Awareness
The agent must inspect `schema.dbml` before making report changes.

This is required so the implementation aligns with:

- table ownership
- foreign keys
- match relationships
- lifecycle rules
- audit behavior
- schedule-to-report relationship

If the schema does not support the needed report behavior, the agent must identify the mismatch before coding.

## Architecture Awareness
The agent must read `docs/architecture/` before coding report logic.

This is required to keep the implementation aligned with:

- DDD boundaries
- CQRS direction
- module layout
- dependency direction
- service interaction patterns
- audit and reliability expectations

## Implementation Discipline
When implementing reports work:

- do not bypass the documented structure
- do not invent new rules without checking the docs first
- do not mix report logic with unrelated domains
- do not skip the read-first sequence
- do not skip the re-read step after planning
- do not violate schedule ownership boundaries

## Expected Report Behavior
The reports domain should support a clean administrator workflow where the company can manage match outcomes safely and clearly.

The expected behavior is:

- create report after a match is completed
- view report in the correct company context
- update report information when needed
- delete report according to lifecycle policy
- keep traceability for every important mutation

## Output Expectation
Report-related work should result in code and data definitions that are:

- secure
- consistent
- auditable
- easy to test
- easy to extend
- aligned with the app’s main goal

## Final Rule
For reports tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md`, `docs/architecture/`, `schema.dbml`, and the main application goal document, must be treated as the source of truth before any reports change is made.

