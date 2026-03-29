# RATE_LIMITING_AND_ABUSE_RESPONSE.md

## Purpose

Describe how to confirm, contain, and recover from request abuse, credential stuffing, or accidental traffic spikes that trigger the API rate limiter.

## Symptoms

- repeated `429 rate_limited` responses on `POST /api/v1/auth/login`
- repeated `429 rate_limited` responses on authenticated write endpoints
- increased authentication failures from one remote address
- sudden spike in create, update, or delete traffic without a matching business event

## Impact

- legitimate clients may be throttled temporarily
- credential stuffing and brute-force attempts are slowed but not fully blocked by default
- write throughput for one caller is reduced while the limiter window is active

## Prerequisites

- access to application logs
- access to the deployed `security.rate_limit` configuration values
- ability to redeploy or roll configuration safely if limits must change

## Immediate actions

1. Confirm the affected path from logs and identify whether the traffic is targeting login or authenticated writes.
2. Check whether the requests originate from one remote address or many addresses.
3. Do not disable the limiter first. Confirm whether the traffic matches normal business activity.
4. If authentication abuse is confirmed, preserve the relevant logs before making configuration changes.

## Recovery steps

1. If the traffic is malicious, keep the limiter enabled and block the source upstream if edge controls are available.
2. If legitimate traffic is being throttled, increase only the specific affected limit in `security.rate_limit` rather than relaxing all limits.
3. Redeploy the configuration change and verify that `429` volume drops to the expected level.
4. If the deployment runs multiple API instances, remember that the current limiter is process-local. Apply any urgent upstream gateway or load balancer protections as well.

## Validation steps

1. Verify that normal login requests succeed again.
2. Verify that normal authenticated writes succeed again.
3. Verify that abusive repeated requests still receive `429` with `Retry-After`.
4. Verify that no internal error responses increased during the incident window.

## Rollback steps

1. Restore the previous `security.rate_limit` values.
2. Redeploy the previous configuration.
3. Confirm that the rollback removed only the intended limiter change.

## Escalation path

1. Escalate to the API owner if legitimate traffic patterns no longer fit the configured limits.
2. Escalate to security operations if the traffic pattern indicates credential stuffing, password spraying, or coordinated abuse.

## Post-incident follow-up

1. Record the affected endpoints, remote addresses, and limiter values used during the incident.
2. Decide whether the limiter thresholds should be tuned permanently.
3. If abuse targeted login across many instances, plan shared-store or edge rate limiting as a follow-up hardening step.
