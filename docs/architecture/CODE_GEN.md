# CODE_GEN.md

## Purpose
This document defines how every generated or agent-produced code change must be documented in a machine-readable JSON report.

## Mandatory reporting
Every create, change, or delete action must be recorded in a JSON file at:

`tools/codegen/{number}-agentic-report.json`

The `{number}` must:
- start at `01`
- increase steadily
- never be reused for a new report

## Report rules
- Use valid JSON only.
- Include one report per meaningful agentic change batch.
- Keep the report honest and complete.
- Document every file touched.
- Document the intent and the risk.

## Required JSON fields
At minimum, each report should include:
- `report_number`
- `timestamp`
- `action_type`
- `summary`
- `files_created`
- `files_changed`
- `files_deleted`
- `reason`
- `risk`
- `tests`
- `notes`

## Recommended report behavior
When multiple files change in one batch, document them in a single report.
When a follow-up edit is made, create the next sequential report file.
When code is regenerated, the report must explain what was regenerated and why.

## Example intent
The report should let a reviewer answer:
- what changed
- why it changed
- whether it was safe
- whether tests were updated
- whether the change was partial or complete

## Required discipline
No undocumented change is acceptable.
No hidden file edit is acceptable.
No silent deletion is acceptable.
