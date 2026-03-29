# auth API

## Purpose

Document the authentication endpoints that issue access tokens and resolve the currently authenticated company administrator.

## Endpoints

### `POST /api/v1/auth/login`

- purpose: authenticate one active company administrator account
- request shape: JSON body with `username` and `password`
- response shape: `200 OK` with `access_token`, `token_type`, `expires_at`, `user`, and `company`
- authorization rule: public endpoint
- validation rule: both fields are required and `username` must map to an active user-company pair
- error cases:
  - `400` when the request body is missing `username` or `password`
  - `401` when the credentials are invalid or the account is soft deleted
  - `500` when the auth flow cannot read the account or sign the token
- audit side effects: no direct write; database-trigger audit behavior remains unchanged because login is a read-only flow

### `GET /api/v1/auth/me`

- purpose: return the authenticated company administrator identity bound to the bearer token
- request shape: `Authorization: Bearer <token>` header
- response shape: `200 OK` with `user` and `company`
- authorization rule: requires a valid bearer token signed by the API and an active underlying account
- validation rule: token must be well formed, signed with the configured secret, unexpired, and still map to an active account row
- error cases:
  - `401` when the header is missing, malformed, expired, invalid, or points to a soft-deleted account
  - `500` when the server cannot resolve the current account
- audit side effects: no direct write; the endpoint re-reads the active account so soft delete invalidates old tokens without adding a session table

## Current scope note

The documented schema has no session or token-revocation table yet, so this auth batch uses stateless JWT access tokens only. Server-side logout or refresh-token rotation should be added in a later auth task when the schema supports session persistence.
