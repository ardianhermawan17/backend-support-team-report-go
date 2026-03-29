# FINALIZATION_TASK.md

## Goal
Finalize the repository by rewriting the project root `README.md` into a clear tutorial for running the application end-to-end, with and without Makefile usage.

This task is documentation-focused. The output should help a new developer understand:
- what the project is,
- how to run it locally,
- how to run it with Docker Compose,
- how to run it with Makefile shortcuts,
- how to verify the app and tests.

## Required reading order for the agent
Before making changes, the agent must follow this sequence:

1. Read `docs/AGENT_README_FIRST.md`
2. Read `docs/DOCS_FOLDER_GUIDE.md`
3. Read the most relevant docs under `docs/` for this task
4. Read the root `README.md`
5. Read the Makefile and any run-related configuration if needed
6. Read the user instruction again
7. Then work

## Agentic workflow
The prompt must follow this working pattern:

**read first → think → plan with TODO → read again → code**

That means the agent should:
1. Read the documentation and current README.
2. Think about what is missing from the tutorial.
3. Make a short TODO plan before editing.
4. Read the source files again to confirm the exact commands and paths.
5. Update the README.

## Scope
Update the root `README.md` so it becomes a practical tutorial for running the project.

The tutorial should cover both paths:
- **with Makefile**
- **without Makefile**

## Minimum content to include in the README
The updated README should explain, at minimum:

- what the project is
- prerequisites
- how to install dependencies
- how to run the application directly
- how to run it with Docker Compose
- how to run it using Makefile commands
- how to run tests
- how to reset or clean the environment if relevant
- basic health-check / verification steps
- any important notes about migration or seeding flow if those are part of the startup process

## Project-specific facts the agent should preserve
Use the repository’s existing documented commands and structure instead of inventing new ones.

The current docs indicate:
- the project is a soccer-team-report backend
- the application runs via `go run ./cmd/api`
- Docker Compose is used for local development
- tests run with `go test ./...`
- the docs folder is the source of truth
- the root README is the right place for the run tutorial

## Documentation rules
The agent must:
- treat `docs/` as the source of truth
- not invent run instructions that conflict with existing docs
- update the README instead of scattering run instructions across multiple docs
- keep the tutorial simple, structured, and beginner-friendly
- avoid re-reading `docs/1_task` more than once
- rely on the docs folder for context instead of guessing

## Suggested README structure
Use a structure similar to this:

1. Project overview
2. Prerequisites
3. Local development
4. Run with Makefile
5. Run without Makefile
6. Database migration / initialization steps
7. Testing
8. Health check
9. Troubleshooting
10. Notes

## Acceptance criteria
The task is complete when:
- root `README.md` is rewritten as a clear run tutorial
- both Makefile and non-Makefile instructions are present
- commands are accurate and consistent with the repo
- the tutorial is easy for a new developer to follow
- no conflicting instructions are introduced
- the content follows the docs-first rule

## Constraints
- Do not rewrite unrelated documentation.
- Do not add implementation code.
- Do not change business logic.
- Do not re-read `docs/1_task` after the first pass.
- Keep the change minimal and focused on onboarding and execution instructions.

## Output expectation
The final result should make it easy for someone to clone the repository and run the project correctly on their own.
