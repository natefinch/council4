---
name: card-authoring
description: Add or author complete Magic cards using native Go and shared Engine operations. Use when asked to add card support, author a card, or continue the authored-card production queue. No Oracle compiler expansion or runtime services.
---

# Card authoring

Use the [canonical guide](../../../docs/card-authoring.md). It owns the worker
loop, scaffold command, rules review, promotion gate and stopping conditions;
do not invent another checklist. Prerequisites: Go, `gh` for live issue/rulings
coordination, and the pinned Oracle corpus. No credentials or service APIs are
needed for gameplay or scaffolding.

Read `.github/copilot-instructions.md`, then the assigned cards' current Oracle
identities, text/rulings and applicable Comprehensive Rules. Use normal Go
functions and inspectable declarations with the existing `cardprogram.Context`.
Keep one writer. Do not expand the parser/lowerer, mutate raw Game state, create
a DSL, or omit abilities to gain coverage.

For example, a request to add Generous Gift routes here: resolve its pinned
identity, inspect Terastodon's destruction/token operations, use `cmd/cardauthor`
for an unregistered draft, and execute named-card scenarios before registration.
It must still create the Elephant when an indestructible legal target survives.

If scaffold inputs are missing/ambiguous, stop that draft and resolve the exact
identity. If a rules operation is absent, search for a precise shared blocker,
leave the whole card unsupported and continue only other assigned eligible
cards. Report failed checks, partial work and local/unpushed commits explicitly.
Do not auto-enable recurring work or create a nested epic.
