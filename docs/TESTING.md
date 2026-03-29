# TESTING.md

## Purpose
This document defines how the project must be tested and what quality level is expected before changes are accepted.

## Testing goals
Testing must prove that:
- business rules are preserved
- database constraints are correct
- API contracts are stable
- race conditions are controlled
- audit logs are written
- company boundaries are enforced

## Test layers
### 1. Unit tests
Use unit tests for:
- pure business rules
- validators
- mappers
- domain services
- ID-related helpers
- score and reporting calculations

### 2. Integration tests
Use integration tests for:
- repositories
- transactions
- unique constraints
- foreign key behavior
- logging side effects
- schedule/report persistence flows

### 3. API tests
Use API tests for:
- request validation
- authorization
- multi-tenant access control
- CRUD behavior
- error responses
- idempotency
- backward compatibility

### 4. Concurrency tests
Use concurrency tests for:
- duplicate player number insertion
- simultaneous match report writes
- duplicate schedule creation when business rules forbid it
- race-sensitive updates

## Coverage expectations
Coverage must prioritize critical business logic over raw percentage alone.

Minimum expectation:
- domain and business rule coverage: very high
- repository and transactional flow coverage: high
- API happy-path coverage: high
- error-path coverage: high
- audit and permission coverage: high

## Test strategy
For every change:
1. add or update the smallest test that proves the rule
2. verify the failure case first when relevant
3. verify the successful case
4. verify database-level protection
5. verify audit side effects

## Required assertions
Tests should assert:
- correct response or result
- correct persistence
- correct unique behavior
- correct tenant isolation
- correct log entry
- correct rollback behavior when a transaction fails

## PUT FILES
- When create a test file for "feature/modules" put it on "tests/integration/modules/{feature/module}"

## What to avoid
Do not rely only on manual testing.
Do not test only the happy path.
Do not approve schema or domain changes without proving the invariants.
