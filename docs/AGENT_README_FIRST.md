# AGENT_README_FIRST.md

## Gateway instruction

This is the first document an agent must read at the beginning of every new session.

## Required reading order

When a new session starts, the agent must read:

1. `docs/AGENT_README_FIRST.md`
2. `docs/DOCS_FOLDER_GUIDE.md`
3. the most relevant document in `docs/` for the task
4. the schema and business model files
5. the task-specific instructions from the user

## Documentation map

Use these documents as the entry points:

- `docs/DOCS_FOLDER_GUIDE.md` for docs organization and routing
- `docs/adr/README.md` for architecture decisions and trade-offs
- `docs/runbooks/README.md` for incidents, recovery, and security operations
- `docs/api/API.md` for API behavior and versioning
- `docs/architecture/MAIN_GOAL_APP.md` for core business rules
- `docs/architecture/CODE_CONTRACT.md` for code-generation and engineering constraints
- `docs/TESTING.md` for test strategy and coverage expectations
- `docs/architecture/CODE_GEN.md` for required codegen reporting format

## Session behavior

The agent must:

- understand the business goal before coding
- check existing docs before making changes
- avoid guessing when a doc already defines the rule
- update the correct document when a rule changes
- keep architecture decisions in ADRs and operational steps in playbooks

## Knowledge routing

Use the docs folder as the source of truth for:

- API design
- business model
- code contract
- testing strategy
- architecture decisions
- operational recovery steps
- code generation reporting

## New session checklist

At the start of a new session, the agent should verify:

- what the business goal is
- which documents apply
- whether the request affects schema, API, testing, codegen, architecture, or operations
- whether the requested change conflicts with existing rules

## Operating principle

Do not improvise over documented rules.
Do not code first and explain later.
Read first, then act.
