# PAGINATION_TASK.md

## Purpose
This document defines the workflow for adding **pagination** support to the application APIs that need it.

The pagination work is scoped to the main soccer-team-report domains that expose list endpoints and need consistent result slicing, response metadata, and test coverage.

Repository context reference:
`https://github.com/ardianhermawan17/backend-support-team-report-go/tree/dev/ardian`

## Scope
This instruction focuses on pagination for these domains:

- `team`
- `player`
- `schedule`
- `report`

The goal is to add pagination only where the API truly needs it, without changing unrelated domains.

## Single Source of Truth
Before making any pagination-related change, the agent must treat the documentation as the source of truth.

The first file to read is:

`docs/AGENT_README_FIRST.md`

That file is the initialization gateway for agentic work and must always be read first.

After that, the agent must read the relevant files under `docs/` to understand the current rules for API design, architecture, code contracts, and testing.

Important:
- `docs/1_task/` must not be re-read for this task
- if it was already read once in a prior flow, do not read it again for pagination work

## Mandatory Reading Order
For every pagination task, the agent must follow this exact sequence:

### 1) Read first
Read the documentation before doing anything else.

Priority reading order:

1. `docs/AGENT_README_FIRST.md`
2. relevant docs under `docs/` except `docs/1_task/`
3. `docs/architecture/`
4. API-related and testing-related docs

### 2) Think
After reading, the agent must reason about:

- which API endpoints need pagination
- the shape of the current response contract
- whether the domain uses page/limit, offset/limit, or cursor style pagination
- how pagination affects filters, sorting, and search
- how pagination should behave for empty results and boundary cases
- how to keep the implementation consistent across team, player, schedule, and report APIs

### 3) Plan with TODO
Before coding, the agent must create a concise TODO plan.

The plan should include:

- which endpoints will get pagination
- what request parameters will be supported
- what response metadata will be returned
- what tests will be added or updated
- what files will be changed

### 4) Read again
After planning, the agent must read the relevant docs again to confirm the plan still matches the documented rules.

This second read is required so the implementation does not drift from the source of truth.

### 5) Code
Only after the second read may the agent begin coding.

## Pagination Design Rules
When adding pagination, the agent must follow these principles:

- use one pagination style consistently across the affected APIs
- keep request and response contracts predictable
- include metadata needed by clients to navigate results
- preserve existing filters and sorting behavior
- avoid breaking existing consumers unless the docs explicitly allow it
- keep pagination logic out of thin handlers when it belongs in application or shared layers

## Domain Coverage
The agent must inspect the following domain APIs to decide whether pagination is needed:

- team list endpoints
- player list endpoints
- schedule list endpoints
- report list endpoints

Pagination should be added only to endpoints that return collections and are expected to grow over time.

## Testing Requirements
Pagination work must include testing updates.

The agent must add or update tests for at least:

- first page result
- middle page result
- last page result
- empty result set
- page beyond last page
- invalid page or limit values
- maximum limit enforcement if the system defines one
- sort stability when pagination is used with filters or ordering

Tests must cover both behavior and response structure.

## Implementation Discipline
When implementing pagination:

- do not change endpoint behavior without checking the docs first
- do not skip the read-first sequence
- do not skip the re-read step after planning
- do not add pagination to endpoints that do not need it
- do not break existing API contracts without documenting the change
- do not treat `docs/1_task/` as a repeated source of truth for this task

## Expected Output
Pagination-related work should result in code and tests that are:

- consistent
- predictable
- easy to consume by clients
- easy to test
- aligned with the application architecture
- aligned with the docs-first workflow

## Final Rule
For pagination tasks, the workflow is always:

**read first → think → plan with TODO → read again → code**

The documentation, especially `docs/AGENT_README_FIRST.md` and the relevant docs under `docs/`, must be treated as the source of truth before any pagination change is made.

