# init_project.md

## Purpose
This document defines how to initialize a new backend project scaffold for the soccer-team-report system using **Go** and **Gin**.

The goal is to create a clean, scalable, audit-friendly starting point that follows the repository conventions, architecture boundaries, and documentation-first workflow.

## Single Source of Truth
Before creating or changing any code, the agent must read the documentation under `docs/` first.

The first file to read is:

`docs/AGENT_README_FIRST.md`

That file is the gateway for the agentic workflow and must be treated as the starting point for every new session.

After that, the agent must continue reading the rest of the relevant docs before writing any code.

## Initialization Order
When starting a new project or a new feature branch, follow this order:

1. Read `docs/AGENT_README_FIRST.md`
2. Read the rest of the relevant files under `docs/`
3. Confirm the project goal and bounded context
4. Scaffold the folder structure
5. Initialize the Go module
6. Add Gin as the HTTP framework
7. Create the application bootstrap and routing entrypoints
8. Add configuration, logging, and environment loading
9. Prepare database, migration, and test folders
10. Only after the scaffold is ready, begin feature code

## Project Scaffolding Rules
The scaffold must be created with these principles:

- Use **Go** as the backend language.
- Use **Gin** as the HTTP framework.
- Keep the codebase modular and ready for DDD + CQRS.
- Separate API edge, application logic, domain logic, and infrastructure concerns.
- Keep shared utilities under a dedicated shared or common layer.
- Prepare the project so it can later support PostgreSQL, audits, workers, and code generation.
- Keep tests and docs alongside the structure from the beginning.

## Required Top-Level Areas
The initial scaffold should include the following areas at minimum:

- `cmd/` for application entrypoints
- `internal/` for application source code
- `docs/` for instructions, ADRs, and playbooks
- `tests/` for unit, integration, and e2e coverage
- `configs/` for environment and runtime configuration
- `deployments/` for Docker and deployment assets
- `scripts/` for migration, seed, and test helpers
- `tools/` for codegen and other internal tooling

## Application Bootstrap Expectations
The first implementation should focus on the minimum structure needed to run the service:

- a Go module
- a Gin HTTP server bootstrap
- configuration loading from environment or config files
- a health or readiness endpoint
- structured logging
- database connection wiring points
- clear route registration boundaries

## Documentation-First Behavior
The agent must not jump directly into code generation from memory.

It must first inspect the docs and use them as the source of truth for:

- folder structure
- naming conventions
- business boundaries
- API behavior
- code contracts
- testing expectations
- code generation reporting

## Code Creation Rule
Code may only be written after the documentation context has been read and understood.

The agent should always:

- verify the docs first
- then create or modify scaffold code
- then update the relevant documentation if needed
- then record any generated or changed artifact according to the codegen workflow

## Output Expectation
The result of `init_project.md` is a project scaffold that is:

- consistent
- maintainable
- audit-friendly
- modular
- ready for Gin-based development
- aligned with the docs-first workflow

## Hard Requirement
Do not treat code as the first source of truth.

For every new session or project initialization, the agent must read the `docs/` folder first, starting from `AGENT_README_FIRST.md`, and only then proceed to code.

