# CLEAN_CODE_TASK.md

> **Agentic task prompt — deduplication and shared-code extraction.**
> This file is read **once** at session start. Do not re-read it mid-task.

---

## Phase 0 — Initialization (mandatory, execute in order)

1. Read `docs/AGENT_README_FIRST.md` — understand session rules and reading order.
2. Read `docs/DOCS_FOLDER_GUIDE.md` — understand doc structure and source-of-truth rules.
3. Read `docs/architecture/CODE_CONTRACT.md` — understand coding principles and forbidden patterns.
4. Read `docs/architecture/MAIN_GOAL_APP.md` — understand business boundaries before touching any module.
5. Read `docs/architecture/FOLDER_STRUCTURE.md` — understand where shared code belongs.
6. Read `docs/architecture/CODE_GEN.md` — understand mandatory codegen reporting format.

> After reading, pause. Do not write code yet. Move to Phase 1.

---

## Phase 1 — Discovery (read the codebase, produce a findings list)

Search the repository for every category of duplication listed below.
Open each suspect file and read its full content before deciding it is a duplicate.

### 1.1 Snowflake ID generators

Suspect files (check all of these and any others you find):

- `internal/modules/team/infrastructure/id/snowflake_generator.go`
- `internal/modules/player/infrastructure/id/snowflake_generator.go`
- `internal/modules/schedule/infrastructure/id/snowflake_generator.go`
- `internal/modules/report/infrastructure/id/snowflake_generator.go`

For each file, record:
- Is the struct name identical?
- Are the constants (`customEpochMillis`, `nodeBits`, `sequenceBits`, `maxNodeID`, `sequenceMask`) identical?
- Is the `NewID()` logic identical?
- Is the only difference the `package` declaration?

### 1.2 HTTP route files

Suspect pattern: every module has its own `routes.go` under `interfaces/http/`.
Check:
- `internal/modules/team/interfaces/http/routes.go`
- `internal/modules/player/interfaces/http/routes.go`
- `internal/modules/schedule/interfaces/http/routes.go`
- `internal/modules/report/interfaces/http/routes.go`
- `internal/modules/auth/interfaces/http/routes.go`

For each file, record what is module-specific vs what is generic boilerplate (dependency wiring pattern, auth middleware attachment, group creation).

### 1.3 `authenticatedAccount` helper

Check every HTTP handler file for a function named `authenticatedAccount`:
- `internal/modules/team/interfaces/http/handlers/handler.go`
- `internal/modules/player/interfaces/http/handlers/handler.go`
- `internal/modules/schedule/interfaces/http/handlers/handler.go`
- `internal/modules/report/interfaces/http/handlers/handler.go`

Record: is the function body byte-for-byte identical across all of them?

### 1.4 `sendJSONRequest` / `sendRequest` test helpers

Check every integration test helper file:
- `tests/integration/modules/teams/teams_http_helpers_test.go`
- `tests/integration/modules/players/players_http_helpers_test.go`
- `tests/integration/modules/schedules/schedules_http_helpers_test.go`
- `tests/integration/modules/reports/reports_http_helpers_test.go`

Record: which helper functions are identical across packages?

### 1.5 `createAccountAndLogin` test helper

Same helper files as above. Record whether `createAccountAndLogin` is identical across all packages.

### 1.6 `IDGenerator` port interface

Check every module port file:
- `internal/modules/team/application/ports/id_generator.go`
- `internal/modules/player/application/ports/id_generator.go`
- `internal/modules/schedule/application/ports/id_generator.go`
- `internal/modules/report/application/ports/id_generator.go`

Record: is the interface definition identical?

---

## Phase 2 — Planning (produce a TODO list before touching any file)

After completing Phase 1, produce a written plan in this format:

```
FINDINGS
========
[List each duplication group with file paths and a one-line description of what is duplicated]

TODO LIST
=========
[ ] 1. Extract SnowflakeGenerator to internal/platform/idgenerator/snowflake.go
[ ] 2. Extract IDGenerator interface to internal/platform/idgenerator/interface.go
[ ] 3. Update all four module id packages to delegate to the shared generator
[ ] 4. Extract authenticatedAccount helper to internal/shared/middleware/auth_context.go (or similar)
[ ] 5. Extract shared test helpers to tests/integration/testhelpers/ package
[ ] 6. Update all callers of extracted code
[ ] 7. Delete now-empty or now-redundant files
[ ] 8. Write codegen report to tools/codegen/{next-number}-agentic-report.json
```

> Stop after writing the TODO list. Do not proceed until the plan is complete.

---

## Phase 3 — Re-read before acting

Before writing any code, re-read the following to confirm nothing was missed:

- `docs/architecture/CODE_CONTRACT.md` — confirm the extraction does not violate any rule.
- `docs/architecture/FOLDER_STRUCTURE.md` — confirm the target paths are correct.
- `docs/TESTING.md` — confirm no test contract is broken.

> Do not re-read this task file (`docs/1_task/CLEAN_CODE_TASK.md` or wherever it lives). It was read once in Phase 0 and is now closed.

---

## Phase 4 — Execute (one TODO item at a time)

Work through the TODO list sequentially. For each item:

1. Create or update the target file.
2. Update every caller of the moved code.
3. Verify the old file can be deleted or kept as a thin wrapper if package boundaries require it.
4. Do not move to the next TODO until the current one compiles cleanly (check imports and package declarations).

### Extraction targets and rules

#### 4.1 Shared Snowflake generator

Target file: `internal/platform/idgenerator/snowflake.go`
Target package: `package idgenerator`

Rules:
- Keep all constants identical to the existing implementations.
- Keep `NewSnowflakeGenerator(nodeID int64)` as the constructor signature.
- Keep the `IDGenerator` interface in the same package or in a sub-file:
  `internal/platform/idgenerator/interface.go`
- Each module's old `infrastructure/id/` package may be replaced with a one-line file
  that re-exports `idgenerator.NewSnowflakeGenerator` under the original function name,
  OR the module's `routes.go` may import `idgenerator` directly.
  Choose the approach that requires the fewest total changes.

#### 4.2 Shared `authenticatedAccount` helper

Target file: `internal/shared/middleware/auth_context.go`
Target package: `package middleware`

Rules:
- The helper reads from the gin context key `"auth.account"`.
- The constant `authenticatedAccountContextKey = "auth.account"` must be defined once in the shared package.
- Each handler that used the local helper must import the shared package.
- The existing `const authenticatedAccountContextKey` declarations in each handler file must be removed.

#### 4.3 Shared integration test helpers

Target package: `tests/integration/testhelpers`

Move these functions (only if they are identical across all packages):
- `sendJSONRequest`
- `sendRequest`
- `createAccountAndLogin`

Rules:
- The new package must be `package testhelpers`.
- Each `*_helpers_test.go` file that previously defined these functions becomes a thin
  file that imports `testhelpers` and re-uses the shared functions — OR the helpers file
  is deleted and callers import `testhelpers` directly.
- Functions that are module-specific (e.g. `createTeamForReportTests`) stay in the
  module's own helpers file.
- Test files in Go cannot be imported across packages unless they are non-`_test` files.
  Export the helpers from a non-test file: `tests/integration/testhelpers/helpers.go`.

#### 4.4 Shared `IDGenerator` interface

Target file: `internal/platform/idgenerator/interface.go`
Each module's `application/ports/id_generator.go` becomes either:
- deleted (module imports `idgenerator.IDGenerator` directly), or
- a type alias: `type IDGenerator = idgenerator.IDGenerator`

---

## Phase 5 — Verify

After all TODO items are complete:

1. Confirm every file that was moved or deleted is no longer referenced by its old import path.
2. Confirm every module still wires its Snowflake node ID correctly in its `routes.go`.
3. Confirm the `authenticatedAccountContextKey` constant value `"auth.account"` is unchanged — the auth middleware sets this key and all handlers must read the same key.
4. Confirm no business logic was altered — only package paths and file locations changed.

---

## Phase 6 — Codegen report

Write a JSON report file at `tools/codegen/{next-number}-agentic-report.json`.

Find the highest existing report number in `tools/codegen/` and increment by one.

The report must include:

```json
{
  "report_number": "NN",
  "timestamp": "<ISO 8601 UTC>",
  "action_type": "REFACTOR",
  "summary": "Extracted duplicated SnowflakeGenerator, IDGenerator interface, authenticatedAccount helper, and shared integration test helpers into shared packages.",
  "files_created": [],
  "files_changed": [],
  "files_deleted": [],
  "reason": "Multiple modules contained byte-for-byte identical implementations. Consolidation reduces maintenance surface and prevents divergence.",
  "risk": "LOW — pure structural refactor, no business logic changed. All callers updated. Constants and interface signatures preserved.",
  "tests": "Existing integration tests verify correctness. No new tests added because no new behaviour was introduced.",
  "notes": ""
}
```

Fill every array completely. Do not leave files_changed or files_deleted empty if changes were made.

---

## Constraints and forbidden actions

- Do not alter any business rule, domain entity, or database query.
- Do not rename any exported symbol — only move it to a new package.
- Do not change the Snowflake node IDs assigned in each module's `routes.go`.
- Do not merge module-specific helpers with the shared helpers.
- Do not skip the codegen report.
- Do not re-read this file after Phase 0.
