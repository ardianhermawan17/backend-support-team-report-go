# MAIN_GOAL_APP.md

## Core business model
The application is a soccer-team-report platform for company administrators to manage teams, players, match schedules, and post-match reports.

## Main objective
Provide a robust system that:
- manages teams inside a company boundary
- manages players under a team
- schedules matches between teams
- stores match outcomes and reporting data
- stores images polymorphically
- records audit logs for every important change

## Business entities
### Company
Represents the administrator organization.
A company owns teams and is linked to a login user.

### User
Represents authentication credentials for the company administrator.
A company has one user account in the current model.

### Team
Represents a soccer team managed by a company.
A company can have many teams.
Each team belongs to exactly one company.

### Player
Represents a player in a team.
One team can have many players.
One player belongs to exactly one team.
Player number must be unique inside one team.

### Schedule
Represents a match between two teams.
A schedule contains home team, guest team, match date, and match time.

### Match report
Represents the result of a completed match.
It stores final score, winner status, most scoring player, and accumulated win counters.

### Image
Represents a polymorphic media record.
An image can belong to a team or a player using `imageable_id` and `imageable_type`.

### Log
Represents the audit trail for create, update, delete, and important business actions.

## Operational goals
The system must be easy to audit, safe to evolve, and scalable without rewriting the core model.

## Non-negotiable business rules
- Every team belongs to one company.
- Every player belongs to one team.
- A player number cannot duplicate inside the same team.
- A schedule must always reference two teams.
- A match report must belong to one schedule.
- Every important mutation must be logged.
- IDs must use Snowflake-style `BIGINT` values.

## Success definition
The application is successful when administrators can manage the complete soccer workflow from company ownership down to match reporting with clear traceability and no ambiguity in the data model.
