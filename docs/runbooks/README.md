# runbooks README

## Purpose

This folder contains operational runbooks for real production situations.
A runbook is a step-by-step guide for handling things that go wrong.

## What belongs here

Store procedures for:

- service outages
- database failures
- queue or worker backlog
- failed deployments
- authentication incidents
- suspicious activity or security events
- data corruption or inconsistent state
- backup and restore workflows
- rollback and recovery workflows
- rate limiting, abuse, or traffic spikes

## What a runbook must answer

Each runbook should clearly answer:

- how to detect the problem
- how to confirm the scope
- what is safe to do first
- what should not be changed immediately
- how to recover service
- how to verify recovery
- when to escalate
- where to record the incident outcome

## runbook format

Each runbook should include:

- title
- purpose
- symptoms
- impact
- prerequisites
- immediate actions
- recovery steps
- validation steps
- rollback steps if needed
- escalation path
- post-incident follow-up

## Security focus

For security-related runbooks, include:

- containment steps
- credential or token rotation guidance
- log locations to inspect
- evidence preservation notes
- user-facing communication rules
- post-incident hardening actions

## Writing rules

- Use direct steps.
- Keep the order operational and easy to follow under pressure.
- Prefer short commands and checks over long explanations.
- Never mix architectural rationale into a runbook.
- Keep runbooks actionable for on-call and incident response use.

## Agent behavior

When the agent changes system behavior that affects incident handling, it should also update the relevant runbook.
