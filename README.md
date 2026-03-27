# backend-sport-team-report-go

Gin-based backend scaffold for the soccer-team-report platform.

## Run

```bash
go run ./cmd/api
```

## Docker

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

For subsequent runs (without rebuilding images):

```bash
docker compose -f deployments/docker/docker-compose.yml up
```

This starts:

- PostgreSQL on `localhost:5432`
- the API on `localhost:8080`
- schema initialization from `deployments/docker/postgres/init/001_initial.sql`

## Health Check

`GET /api/v1/health`

## Tests

```bash
go test ./...
```

The integration tests start a real PostgreSQL container through Docker, apply the SQL migrations, verify the auth repository, and then verify the health endpoint against that live database.
