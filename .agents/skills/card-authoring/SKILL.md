---
name: card-authoring
description: Add or author complete Magic cards using native Go and shared Engine operations. Use when asked to add card support, author a card, or continue the authored-card production queue. No Oracle compiler expansion or runtime services.
---

# Card authoring

Use the [canonical guide](../../../docs/card-authoring.md). It owns the worker
loop, scaffold command, behavioral evidence and stopping conditions;
do not invent another checklist. Use the
[capability lookup](../../../mtg/cardprogram/README.md#capability-lookup)
to map rules wording to declarations, shared operations and behavioral prior art.
Prerequisites: Go, `gh` for live issue/rulings
coordination, and the pinned Oracle corpus. No credentials or service APIs are
needed for gameplay or scaffolding.

Follow the [repository policy](../../../.github/copilot-instructions.md),
then read the assigned cards' current Oracle identities, text/rulings and
applicable Comprehensive Rules. **Before designing or implementing the card,
open the official rules document or a verified local copy and read the relevant
sections.** Prior knowledge, rulings, repository comments and another card's
citations do not replace this step. Record the document edition, URL or local
path, sections actually read and the behavior they require in the existing
source record. If the document cannot be accessed or verified, report the
source-review blocker; do not claim a fully verified admission. Use normal Go
functions and inspectable declarations with the existing `cardprogram.Context`.
Keep one writer. Do not expand the parser/lowerer, mutate raw Game state, create
a DSL, or omit abilities to gain coverage.

For example, a request to add Generous Gift routes here: resolve its pinned
identity, inspect Terastodon's destruction/token operations, use `cmd/cardauthor`
for an unregistered draft, and execute named-card scenarios before registration.
It must still create the Elephant when an indestructible legal target survives.

If scaffold inputs are missing/ambiguous, stop that draft and resolve the exact
identity. Before declaring a rules operation absent, follow the canonical guide's
capability investigation: read complete relevant declarations, trace the Engine
implementation and inspect behavioral prior art rather than guessing API names.
If a rules operation is absent, search for a precise shared blocker,
leave the whole card unsupported and continue only other assigned eligible
cards. Report failed checks, partial work and local/unpushed commits explicitly.
Do not auto-enable recurring work or create a nested epic.
