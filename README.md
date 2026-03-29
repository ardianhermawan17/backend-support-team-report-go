# backend-sport-team-report-go

Backend API for a soccer team reporting application. It is built with Gin and PostgreSQL, and exposes its endpoints under `http://localhost:8080/api/v1`.

If you only need to get the project running, use one of these paths:

1. Makefile
2. Docker Compose
3. Local manual Go commands

## Before You Start

You will need:

- Go `1.25.0` or newer
- Docker and Docker Compose
- GNU Make if you want to use the Makefile commands
- PostgreSQL if you want to run the app locally without Docker Compose

Configuration defaults live in `configs/app.yaml` and `configs/app.local.yaml`.

Environment overrides supported by the app:

- `APP_NAME`
- `APP_ENV`
- `APP_HOST`
- `APP_PORT`
- `DATABASE_DSN`
- `AUTH_JWT_SECRET`

Use `.env.example` as a reference. The app does not load `.env` automatically.

## 1. Run With Makefile

This is the easiest option if you already have `make` installed.

### Full project startup

```bash
make docker-compose-build
```

This flow:

- builds the Docker images
- starts PostgreSQL
- runs migrations
- runs seeding
- starts the API on `localhost:8080`

Useful commands after that:

```bash
make docker-up
make docker-logs
make docker-down
```

Use `make docker-up` when the images are already built and you just want to start the containers again.

### Run against an existing local database

If PostgreSQL is already running outside Docker, use:

```bash
make tidy
make migrate
make seeding
make run
```

### Run tests

```bash
make test
```

## 2. Run With Docker Compose

Use this if you do not want to use `make`, but still want the full container-based setup.

### First-time startup

```bash
docker compose -f deployments/docker/docker-compose.yml build migrate seed api
docker compose -f deployments/docker/docker-compose.yml up -d postgres
docker compose -f deployments/docker/docker-compose.yml run --rm migrate
docker compose -f deployments/docker/docker-compose.yml run --rm seed
docker compose -f deployments/docker/docker-compose.yml up -d api
```

### Start again later

```bash
docker compose -f deployments/docker/docker-compose.yml up -d
```

### View logs

```bash
docker compose -f deployments/docker/docker-compose.yml logs -f
```

### Stop and remove the local database volume

```bash
docker compose -f deployments/docker/docker-compose.yml down -v
```

## 3. Run Locally With Go Commands

Use this when you want to run the API directly on your machine and connect it to a PostgreSQL instance you manage yourself.

### Install dependencies

```bash
go mod tidy
```

### Set environment variables if needed

Default local database DSN:

```text
postgres://postgres:postgres@localhost:5432/soccer_team_report?sslmode=disable
```

PowerShell example:

```powershell
$env:APP_ENV = "local"
$env:DATABASE_DSN = "postgres://postgres:postgres@localhost:5432/soccer_team_report?sslmode=disable"
$env:AUTH_JWT_SECRET = "local-dev-only-change-me"
```

POSIX shell example:

```bash
export APP_ENV=local
export DATABASE_DSN=postgres://postgres:postgres@localhost:5432/soccer_team_report?sslmode=disable
export AUTH_JWT_SECRET=local-dev-only-change-me
```

### Run migrations

```bash
go run ./cmd/migrate
```

### Seed the database

```bash
go run ./cmd/seeding
```

### Start the API

```bash
go run ./cmd/api
```

## Seeded Local Login

After seeding, you can log in with:

- username: `admin`
- password: `password`
- email: `admin@gmail.com`

## Verify The App

Health check:

```bash
curl http://localhost:8080/api/v1/health
```

Login check:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

## Tests

Run the full suite with:

```bash
go test ./...
```

Some integration tests start a real PostgreSQL container through Docker, so Docker needs to be available even if you are not using Docker for normal local development.

## Notes

- Main app entrypoint: `go run ./cmd/api`
- Migration entrypoint: `go run ./cmd/migrate`
- Seed entrypoint: `go run ./cmd/seeding`
- Migrations live in `internal/platform/database/migrations`
- The health endpoint is `GET /api/v1/health`
- The `docs/` folder is still the source of truth for project rules and architecture
