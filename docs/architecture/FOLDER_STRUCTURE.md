# Golang Backend Folder Structure

Designed for a high-concurrency booking platform using **Gin + DDD + CQRS**, with clear boundaries for API, domain logic, testing, reliability, and Dockerized deployment.

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── worker/
│   │   └── main.go
│   └── migrate/
│       └── main.go
│
├── internal/
│   ├── bootstrap/
│   │   ├── app.go
│   │   ├── dependencies.go
│   │   └── lifecycle.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   ├── env.go
│   │   └── loader.go
│   │
│   ├── shared/
│   │   ├── errors/
│   │   ├── logger/
│   │   ├── middleware/
│   │   ├── observability/
│   │   ├── paginator/
│   │   ├── response/
│   │   └── validator/
│   │
│   ├── platform/
│   │   ├── database/
│   │   │   ├── postgres/
│   │   │   ├── redis/
│   │   │   └── migrations/
│   │   ├── cache/
│   │   ├── messaging/
│   │   │   ├── producer/
│   │   │   ├── consumer/
│   │   │   └── channelbus/
│   │   ├── httpserver/
│   │   ├── idgenerator/
│   │   ├── time/
│   │   └── transaction/
│   │
│   ├── modules/
│   │   ├── auth/
│   │   │   ├── domain/
│   │   │   │   ├── entities/
│   │   │   │   ├── valueobjects/
│   │   │   │   ├── events/
│   │   │   │   ├── repositories/
│   │   │   │   └── services/
│   │   │   ├── application/
│   │   │   │   ├── commands/
│   │   │   │   ├── queries/
│   │   │   │   ├── handlers/
│   │   │   │   ├── dtos/
│   │   │   │   └── ports/
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/
│   │   │   │   ├── jwt/
│   │   │   │   ├── cache/
│   │   │   │   └── messaging/
│   │   │   └── interfaces/
│   │   │       └── http/
│   │   │           ├── handlers/
│   │   │           ├── requests/
│   │   │           ├── responses/
│   │   │           └── routes.go
│   │   │
│   │   ├── user/
│   │   │   ├── domain/
│   │   │   │   ├── entities/
│   │   │   │   ├── valueobjects/
│   │   │   │   ├── events/
│   │   │   │   ├── repositories/
│   │   │   │   └── services/
│   │   │   ├── application/
│   │   │   │   ├── commands/
│   │   │   │   ├── queries/
│   │   │   │   ├── handlers/
│   │   │   │   ├── dtos/
│   │   │   │   └── ports/
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/
│   │   │   │   ├── cache/
│   │   │   │   └── messaging/
│   │   │   └── interfaces/
│   │   │       └── http/
│   │   │           ├── handlers/
│   │   │           ├── requests/
│   │   │           ├── responses/
│   │   │           └── routes.go
│   │   │
│   │   ├── "name"/
│   │   │
│   │   ├── room/
│   │   ├── booking/
│   │   ├── payment/
│   │   ├── review/
│   │   └── notification/
│   │
│   ├── cqrs/
│   │   ├── commandbus/
│   │   ├── querybus/
│   │   ├── dispatcher/
│   │   ├── middleware/
│   │   └── projections/
│   │
│   ├── async/
│   │   ├── jobs/
│   │   ├── workers/
│   │   ├── pipelines/
│   │   ├── channel/
│   │   └── scheduler/
│   │
│   └── api/
│       ├── gin/
│       │   ├── router/
│       │   ├── handlers/
│       │   ├── middlewares/
│       │   └── routes/
│       ├── rest/
│       └── docs/
│
├── pkg/
│   ├── clock/
│   ├── crypto/
│   ├── id/
│   └── retry/
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── Dockerfile.dev
│   │   └── docker-compose.yml
│   ├── k8s/
│   └── nginx/
│
├── configs/
│   ├── app.yaml
│   ├── app.local.yaml
│   └── app.test.yaml
│
├── scripts/
│   ├── migrate.sh
│   ├── seed.sh
│   └── test.sh
│
├── tests/
│   ├── unit/
│   ├── integration/
│   ├── contract/
│   ├── e2e/
│   ├── fixtures/
│   └── testdata/
│
├── docs/
│   ├── architecture/
│   ├── api/
│   ├── adr/
│   └── runbooks/
│
├── tools/
│   ├── codegen/
│   └── lint/
│
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── .gitignore
└── README.md
```

## Notes for agentic AI use

- Put **business rules** inside each module’s `domain/` folder.
- Put **CQRS commands and queries** inside `application/commands` and `application/queries`.
- Put **Gin handlers only at the edge** in `interfaces/http`; they should not contain business logic.
- Keep **cross-module communication** through domain events, application ports, or messaging abstractions.
- Use `internal/async/channel/` for controlled concurrent work, background fan-out, and queue-like processing.
- Keep **tests mirrored to the code** so reliability checks are easy to locate and maintain.
- Keep Docker files under `deployments/docker/` so local, dev, and production builds stay separated.

## Suggested bounded contexts

- `auth` — login, sessions, tokens, permissions
- `user` — profiles, preferences, identity data
