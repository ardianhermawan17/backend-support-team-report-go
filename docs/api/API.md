# API.md

## Purpose

This document defines how the API must be designed, when a new endpoint is allowed to be created, and how the API should evolve without breaking existing clients.

## API design rules

The API must be resource-oriented, predictable, and audit-friendly. Prefer stable nouns over action-heavy names.

Use these principles:

- Version all public endpoints.
- Keep endpoints small and single-purpose.
- Separate read models from write models when the business case benefits from it.
- Prefer explicit request/response contracts.
- Never expose internal database assumptions directly to clients.

## When to create an API

Create a new API only when one of these is true:

- A new business capability is needed by the application.
- Existing endpoints cannot express the required workflow safely.
- A read model needs to be optimized for reporting or dashboard use.
- A write flow needs a dedicated command to preserve integrity, validation, or concurrency rules.

Do not create an endpoint just because a table exists.

## API shape

Prefer these patterns:

- `GET` for reading resources
- `POST` for creating resources
- `PUT` or `PATCH` for updating resources
- `DELETE` for soft delete or controlled removal

Use nested resources only when the parent-child relationship is clear and stable. Avoid deep nesting that makes the API hard to audit.

## Business-driven API mapping

The soccer-team-report application should expose APIs around:

- companies
- teams
- players
- schedules
- match reports
- images
- logs

Each API must respect company boundaries. A user from one company must not access another company’s teams, players, schedules, reports, or images.

## Required API behavior

- Validate ownership and tenancy on every write operation.
- Enforce uniqueness rules at the API layer and again at the database layer.
- Return deterministic error messages for business rule violations.
- Make write endpoints idempotent whenever reasonable.
- Include audit context such as actor, company, timestamp, and request correlation id.

## API lifecycle

Before an API is added:

1. Confirm the business rule.
2. Confirm the resource boundary.
3. Confirm the database impact.
4. Confirm the test coverage.
5. Confirm the audit/logging impact.

Before an API is changed:

1. Check backward compatibility.
2. Check query and command contracts.
3. Check migration impact.
4. Check test updates.
5. Check whether old clients still work.

Before an API is deleted:

1. Check whether it is still used.
2. Provide a deprecation path if necessary.
3. Preserve historical audit data.
4. Remove only when safe.

## API documentation standard

Every endpoint must document:

- purpose
- request shape
- response shape
- authorization rule
- validation rule
- error cases
- audit side effects

## Teams management endpoints

The teams domain exposes CRUD endpoints inside the authenticated company boundary:

- `POST /api/v1/teams`
- `GET /api/v1/teams`
- `GET /api/v1/teams/{team_id}`
- `PUT /api/v1/teams/{team_id}`
- `DELETE /api/v1/teams/{team_id}`

Rules for these endpoints:

- Require authenticated bearer token.
- Resolve company context from the authenticated account, never from request body.
- Allow access only to teams owned by the authenticated company.
- Use soft delete for removal.
- Return deterministic errors for invalid payload, not found, and uniqueness conflicts.
