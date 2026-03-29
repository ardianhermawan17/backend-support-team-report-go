# DATABASE_SEEDER_TASK.md

## Goal
Implement application database seeding for the soccer-team-report backend with Snowflake-style `BIGINT` IDs, proper schema alignment, and test coverage for both seeding and core business models.

## Required reading order
Read these files first, in this order:

1. `docs/AGENT_README_FIRST.md`
2. `docs/DOCS_FOLDER_GUIDE.md`
3. `docs/architecture/MAIN_GOAL_APP.md`
4. `docs/architecture/CODE_CONTRACT.md`
5. `docs/TESTING.md`
6. `docs/architecture/schema.dbml`
7. `docs/architecture/initial.db`
8. the task-specific request from the user

Use the `docs/` folder as the source of truth after that. Read `docs/1_task` only once during initialization, and do not re-read it afterward.

## Thinking workflow
Follow this exact agent flow:

**read first → think → plan with TODO → read again → code**

That means:
- read the docs and existing implementation first
- think through schema impact, seeding dependencies, and test strategy
- write a short TODO plan before changing code
- re-read only the relevant source files before implementation
- then code the smallest correct change set

## Scope

### 1) Database seeder
Create a seeder for application bootstrap data.

Seed order should respect foreign keys and business dependencies:
1. `users`
2. `companies`
3. `teams`
4. `images` if needed for team/player relations
5. `players`
6. `schedules`
7. `reports`

The seed data must be deterministic and use the existing Snowflake generator implementation from:
- `internal/platform/idgenerator/snowflake.go`
- `internal/platform/idgenerator/interface.go`

Do not replace Snowflake IDs with random integers.

### 2) Default admin seed
Create at least one admin user with:
- username: `admin`
- password: `password`
- email: `admin@gmail.com`

Password handling must follow the current authentication implementation and repository conventions. Store the password in the correct hashed form used by the project, not plain text.

### 3) Schema change for `USER`
Add `email` to the `USER` table everywhere the model exists:
- database schema
- DBML schema file
- initial SQL / database bootstrap file
- any user entity, repository, DTO, mapper, query, fixture, or test that depends on the `users` table

If the schema uses a `schema.dbml` and `initial.db`, update both so they remain consistent.

### 4) Makefile behavior
Update the Makefile so seeding runs only when the docker compose build flow is executed.

Required behavior:
- `make docker-compose-build` should include seeding as part of the build/start flow
- `make seeding` should exist as a dedicated command to run only the seeding application
- seeding must not run from normal local test or run targets unless explicitly invoked

If the repository currently uses a different build target name, align it so `docker-compose-build` becomes the canonical one or is an explicit alias for the existing build flow.

### 5) Testing
Add or update tests for:

#### Seeding application
Verify that:
- the seeder inserts the expected base records
- admin user exists with the expected username and email
- seeding is repeatable and does not create invalid duplicates
- IDs remain `BIGINT`/Snowflake-compatible
- foreign key order and constraints are respected

#### Core business models
Add or update tests for the core models affected by the schema change:
- user
- company
- team
- player
- schedule
- report
- image if the relation is touched

Tests must cover both happy path and constraint/protection behavior where relevant.

## Implementation notes

### File change handling
Scan the repository for every place affected by the new `email` field and the seeding flow. Do not patch only the schema file and forget the application layer.

### Core consistency rules
- Keep `BIGINT` IDs everywhere business entities require them
- Keep user/company/team/player/schedule/report relationships consistent
- Preserve current business rules from `MAIN_GOAL_APP.md`
- Preserve testing discipline from `TESTING.md`

### Suggested touch points
Depending on the current code layout, expect changes in:
- database migration/bootstrap files
- DBML schema
- seeding command / entrypoint
- seed service / repository layer
- Makefile
- domain model / entity definitions
- auth/user repository tests
- integration tests for seeded data
- any generated or fixture files that reference `users`

## Acceptance criteria
The task is complete only when all of these are true:

- `email` exists on the user model and schema consistently
- default admin user is seeded with the required credentials and email
- Snowflake generator is used for seeded entity IDs
- `make seeding` runs only the seeder
- `make docker-compose-build` includes seeding in the build/start flow
- tests prove seeding behavior and core model correctness
- no repo file that depends on the user schema is left inconsistent

## Notes for the implementer
- Do not guess over documented rules.
- Do not skip database-level coverage.
- Keep the change focused and traceable.
- Update docs only if a rule must change.
