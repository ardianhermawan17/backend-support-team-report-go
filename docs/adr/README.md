# ADR README

## Purpose
This folder contains Architecture Decision Records (ADRs).

## What an ADR is
An ADR is a short, permanent record of an important technical decision.
It captures:
- the decision that was made
- why it was made
- the alternatives that were considered
- the trade-offs that were accepted
- the impact on the system going forward

## When to create an ADR
Create or update an ADR when a change affects any of the following:
- architecture boundaries
- data modeling decisions
- persistence strategy
- API style or versioning direction
- concurrency or transaction strategy
- security architecture
- background processing or messaging approach
- testing strategy that changes the engineering contract
- deployment or operational design with long-term impact

## ADR format
Each ADR should include:
- title
- status
- date
- context
- decision
- alternatives considered
- trade-offs
- consequences
- related documents

## Writing rules
- Keep the document concise and specific.
- Describe the reasoning, not just the outcome.
- Record rejected options when they matter.
- Avoid implementation code unless a tiny snippet is needed to explain the decision.
- Update or supersede old ADRs rather than silently changing the history.

## Agent behavior
Before changing architecture-sensitive code, the agent should check whether an ADR already defines the rule.
If a new decision is made, the agent should add a new ADR instead of hiding the decision inside code comments.
