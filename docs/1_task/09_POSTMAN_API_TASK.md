# POSTMAN_API_TASK.md

> **Agentic task prompt — generate a Postman v2.1 collection from all registered API routes.**
> This file is read **once** at session start. Do not re-read it mid-task.

---

## Phase 0 — Initialization (mandatory, execute in order)

1. Read `docs/AGENT_README_FIRST.md` — understand session rules and reading order.
2. Read `docs/DOCS_FOLDER_GUIDE.md` — understand doc structure and source-of-truth rules.
3. Read `docs/api/API.md` — understand API design rules, resource boundaries, and versioning.
4. Read `docs/api/auth.md` — understand authentication endpoints and token flow.
5. Read `docs/architecture/MAIN_GOAL_APP.md` — understand the business entities and what each endpoint serves.
6. Read `docs/architecture/CODE_GEN.md` — understand mandatory codegen reporting format.

> After reading, pause. Do not write any file yet. Move to Phase 1.

---

## Phase 1 — Route Discovery (read every routes file)

Open and read every route registration file in the codebase. Extract the HTTP method, path, and handler name for each registered endpoint.

### 1.1 Auth routes

Read: `internal/modules/auth/interfaces/http/routes.go`
Read: `internal/modules/auth/interfaces/http/handlers/handler.go`

Record every endpoint with its method, full path, and handler function.

### 1.2 Team routes

Read: `internal/modules/team/interfaces/http/routes.go`
Read: `internal/modules/team/interfaces/http/handlers/handler.go`
Read: `internal/modules/team/interfaces/http/requests/team_requests.go`
Read: `internal/modules/team/interfaces/http/responses/team_response.go`

### 1.3 Player routes

Read: `internal/modules/player/interfaces/http/routes.go`
Read: `internal/modules/player/interfaces/http/handlers/handler.go`
Read: `internal/modules/player/interfaces/http/requests/player_requests.go`
Read: `internal/modules/player/interfaces/http/responses/player_response.go`

### 1.4 Schedule routes

Read: `internal/modules/schedule/interfaces/http/routes.go`
Read: `internal/modules/schedule/interfaces/http/handlers/handler.go`
Read: `internal/modules/schedule/interfaces/http/requests/schedule_requests.go`
Read: `internal/modules/schedule/interfaces/http/responses/schedule_response.go`

### 1.5 Report routes

Read: `internal/modules/report/interfaces/http/routes.go`
Read: `internal/modules/report/interfaces/http/handlers/handler.go`
Read: `internal/modules/report/interfaces/http/requests/report_requests.go`
Read: `internal/modules/report/interfaces/http/responses/report_response.go`

### 1.6 Health route

Read: `internal/api/gin/routes/routes.go`

---

## Phase 2 — Planning (produce a complete route inventory before writing any file)

After completing Phase 1, write a route inventory in this format:

```
ROUTE INVENTORY
===============

[Health]
GET  /api/v1/health

[Auth]
POST /api/v1/auth/login
GET  /api/v1/auth/me          (requires Bearer token)

[Teams]
POST   /api/v1/teams           (requires Bearer token)
GET    /api/v1/teams           (requires Bearer token)
GET    /api/v1/teams/:team_id  (requires Bearer token)
PUT    /api/v1/teams/:team_id  (requires Bearer token)
DELETE /api/v1/teams/:team_id  (requires Bearer token)

[Players]
POST   /api/v1/teams/:team_id/players                    (requires Bearer token)
GET    /api/v1/teams/:team_id/players                    (requires Bearer token)
GET    /api/v1/teams/:team_id/players/:player_id         (requires Bearer token)
PUT    /api/v1/teams/:team_id/players/:player_id         (requires Bearer token)
DELETE /api/v1/teams/:team_id/players/:player_id         (requires Bearer token)

[Schedules]
POST   /api/v1/schedules                (requires Bearer token)
GET    /api/v1/schedules                (requires Bearer token)
GET    /api/v1/schedules/:schedule_id   (requires Bearer token)
PUT    /api/v1/schedules/:schedule_id   (requires Bearer token)
DELETE /api/v1/schedules/:schedule_id   (requires Bearer token)

[Reports]
POST   /api/v1/reports              (requires Bearer token)
GET    /api/v1/reports              (requires Bearer token)
GET    /api/v1/reports/:report_id   (requires Bearer token)
PUT    /api/v1/reports/:report_id   (requires Bearer token)
DELETE /api/v1/reports/:report_id   (requires Bearer token)
```

Extend or correct this inventory based on what you actually read in Phase 1.

> Stop after writing the inventory. Do not produce output files until the inventory is verified complete.

---

## Phase 3 — Re-read before acting

Before generating JSON, re-read the following to capture request/response shapes:

- `docs/api/API.md` — confirm authorization rules, error contracts, and response conventions.
- `docs/api/auth.md` — confirm exact request/response fields for login and me endpoints.

Read the relevant `requests/` and `responses/` Go files for each module to extract exact JSON field names. The Go struct field tags (`json:"..."`) are the canonical field names.

> Do not re-read this task file. It was read once in Phase 0 and is now closed.

---

## Phase 4 — Generate Postman Collection JSON

Produce a single Postman Collection v2.1 JSON file at:

`tools/postman/soccer-team-report.postman_collection.json`

Create the `tools/postman/` directory if it does not exist.

### Collection structure

```
Soccer Team Report API
├── Health
│   └── Health Check
├── Auth
│   ├── Login
│   └── Me (Current Account)
├── Teams
│   ├── Create Team
│   ├── List Teams
│   ├── Get Team
│   ├── Update Team
│   └── Delete Team
├── Players
│   ├── Create Player
│   ├── List Players
│   ├── Get Player
│   ├── Update Player
│   └── Delete Player
├── Schedules
│   ├── Create Schedule
│   ├── List Schedules
│   ├── Get Schedule
│   ├── Update Schedule
│   └── Delete Schedule
└── Reports
    ├── Create Report
    ├── List Reports
    ├── Get Report
    ├── Update Report
    └── Delete Report
```

### Collection-level variables

Define these as collection variables so requests can reference them with `{{variable_name}}`:

| Variable | Initial value | Description |
|---|---|---|
| `base_url` | `http://localhost:8080` | API base URL |
| `access_token` | *(empty)* | Filled automatically by the Login request test script |
| `team_id` | `1` | Example team ID |
| `player_id` | `1` | Example player ID |
| `schedule_id` | `1` | Example schedule ID |
| `report_id` | `1` | Example report ID |

### Authorization

- The collection-level auth must be set to `Bearer Token` using `{{access_token}}`.
- The `Login` request and `Health Check` request must override auth to `No Auth`.
- All other requests inherit the collection-level Bearer token.

### Login request — test script

The Login request must include a Postman test script that saves the token automatically:

```javascript
if (pm.response.code === 200) {
    const body = pm.response.json();
    pm.collectionVariables.set("access_token", body.access_token);
}
```

### Request body rules

Every POST and PUT request must include a JSON body with all fields from the corresponding Go request struct. Use realistic placeholder values, not empty strings.

#### POST /api/v1/auth/login
```json
{
    "username": "admin",
    "password": "password"
}
```

#### POST /api/v1/teams / PUT /api/v1/teams/:team_id
```json
{
    "name": "Thunder FC",
    "logo_image_id": null,
    "founded_year": 2014,
    "homebase_address": "Jalan Stadion Nomor 1",
    "city_of_homebase_address": "Bandung"
}
```

#### POST /api/v1/teams/:team_id/players / PUT /api/v1/teams/:team_id/players/:player_id
```json
{
    "name": "Rizky Pratama",
    "height": 180.5,
    "weight": 74.2,
    "position": "striker",
    "player_number": 9,
    "profile_image_id": null
}
```

#### POST /api/v1/schedules / PUT /api/v1/schedules/:schedule_id
```json
{
    "match_date": "2026-07-10",
    "match_time": "15:30:00",
    "home_team_id": "{{team_id}}",
    "guest_team_id": "{{team_id}}"
}
```

#### POST /api/v1/reports / PUT /api/v1/reports/:report_id
```json
{
    "match_schedule_id": "{{schedule_id}}",
    "final_score_home": 3,
    "final_score_guest": 1,
    "most_scoring_goal_player_id": "{{player_id}}"
}
```

### Path variable handling

For endpoints that include path parameters (`:team_id`, `:player_id`, `:schedule_id`, `:report_id`), use Postman path variables in the URL that reference the collection variables:

```
{{base_url}}/api/v1/teams/{{team_id}}
{{base_url}}/api/v1/teams/{{team_id}}/players/{{player_id}}
```

### Expected response documentation

Each request must include a description that documents:
- the successful HTTP status code
- the top-level response shape (field names)
- common error codes (400, 401, 404, 409, 500)

Use the following error shape for all error responses:
```json
{
    "error": "<error_code>",
    "message": "<human readable message>"
}
```

### Postman v2.1 JSON schema requirements

The output file must be valid Postman Collection v2.1. Required top-level fields:

```json
{
    "info": {
        "_postman_id": "<any UUID>",
        "name": "Soccer Team Report API",
        "description": "Complete API collection for the soccer-team-report platform. Covers authentication, team management, player management, match scheduling, and match reports.",
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    },
    "variable": [ ... ],
    "auth": { "type": "bearer", "bearer": [{ "key": "token", "value": "{{access_token}}", "type": "string" }] },
    "item": [ ... ]
}
```

Each item (folder) must have:
```json
{
    "name": "Teams",
    "item": [ ... ]
}
```

Each request item must have:
```json
{
    "name": "Create Team",
    "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": { "mode": "raw", "raw": "...", "options": { "raw": { "language": "json" } } },
        "url": {
            "raw": "{{base_url}}/api/v1/teams",
            "host": ["{{base_url}}"],
            "path": ["api", "v1", "teams"]
        },
        "description": "..."
    },
    "response": []
}
```

---

## Phase 5 — Verify

After generating the file:

1. Confirm the JSON is syntactically valid (no trailing commas, all brackets closed).
2. Confirm every route from the Phase 2 inventory has exactly one corresponding request item.
3. Confirm the `Login` request has the test script to save `access_token`.
4. Confirm GET and DELETE requests have no `body` field.
5. Confirm the collection variable `base_url` defaults to `http://localhost:8080`.
6. Confirm the file path is `tools/postman/soccer-team-report.postman_collection.json`.

---

## Phase 6 — Codegen report

Write a JSON report file at `tools/codegen/{next-number}-agentic-report.json`.

Find the highest existing report number in `tools/codegen/` and increment by one.

```json
{
  "report_number": "NN",
  "timestamp": "<ISO 8601 UTC>",
  "action_type": "CREATE",
  "summary": "Generated Postman Collection v2.1 covering all API routes: health, auth, teams, players, schedules, and reports.",
  "files_created": [
    "tools/postman/soccer-team-report.postman_collection.json"
  ],
  "files_changed": [],
  "files_deleted": [],
  "reason": "No machine-readable API documentation existed. The Postman collection allows developers and testers to exercise all endpoints without manual configuration.",
  "risk": "NONE — this task creates a documentation artifact only. No source code or schema was modified.",
  "tests": "No test changes required. Collection can be used with Postman or Newman for manual or automated API verification.",
  "notes": "Import the collection into Postman, set the base_url variable if needed, and run Login first to populate the access_token variable."
}
```

---

## Constraints and forbidden actions

- Do not modify any Go source file.
- Do not modify any migration SQL file.
- Do not invent endpoints that do not exist in the codebase.
- Do not omit any endpoint that is registered in a `routes.go` file.
- Do not use hardcoded token values — always use `{{access_token}}`.
- Do not skip the codegen report.
- Do not re-read this file after Phase 0.
