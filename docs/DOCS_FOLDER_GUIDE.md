# DOCS_FOLDER_GUIDE.md

## Purpose

This document defines how the `docs/` folder is organized and how an agent should use it.

## Source-of-truth rule

When a task touches architecture, operations, API behavior, testing, or code generation, the agent must first look for an existing document in `docs/` before creating a new rule.

## Folder intent

- `docs/adr/` stores Architecture Decision Records.
- `docs/playbooks/` stores operational playbooks and incident response steps.
- `docs/api/` stores API behavior, versioning, request/response conventions, and endpoint rules.
- `docs/architecture/` stores broader architecture guidance, system boundaries, and design principles.
- `docs/runbooks/` stores operational procedures if the project keeps them separate from playbooks.
- `docs/` root files define global documentation rules and reading order.

## Writing rules for documentation

- Keep each document focused on one responsibility.
- Write instructions, not implementation code.
- Prefer explicit rules over vague suggestions.
- Use stable naming and keep terminology consistent with the codebase.
- When a new rule changes an old rule, update the old file instead of duplicating conflicting guidance.

## Agent behavior

The agent should:

- read the gateway file first
- find the most specific matching document
- follow the most recent rule if two docs overlap
- avoid inventing process when a doc already defines one
- update documentation together with any schema, API, or codegen change

## Change discipline

If a change affects decision history, it belongs in `docs/adr/`.
If a change affects handling failures, security incidents, deployment, or recovery, it belongs in `docs/runbooks/`.
