# Card Rollout

Batch-orchestrates council4 card implementation from a card-list file using
`cardbatch`, `card-impl`, validation, and unsupported-card reports with a
rollup of missing rules/parser functionality by affected card.

## When to Use This Skill

- You have a list of Magic cards on disk and want to implement supported cards.
- You want an agent to run `card-impl` over a worklist in small batches.
- You want a Markdown/JSON report of cards blocked by missing parser or rules
  support.
- You want to continue a previous `.cardwork` card rollout.

## Prerequisites

- Run from the council4 repository root.
- Go tooling must be available.
- Network access is needed for Scryfall fetches unless the cache is already
  populated.
- The `card-impl` project skill must exist at `.agents/skills/card-impl`.

## Model Defaults

- Use `gpt-5.4-mini` for rollout orchestration and manifest/report-only work.
- Use `claude-sonnet-4.6` for implementation subagents that edit generated card
  files.
- A cheaper implementation trial is OK for simple cards, but failed validation
  should be retried with the stronger coding model before calling the card
  unsupported.

## Usage

Ask Copilot:

- "Use card-rollout on `cards.txt`"
- "Batch implement cards from this list"
- "Run card-impl over `.cardwork/staples/cards.json`"
- "Process this Commander staples list and report unsupported cards"

The skill will:

1. Parse/fetch a card manifest with `cardbatch`.
2. Build a small worklist.
3. Spawn subagents to implement card batches through the `card-impl` workflow.
4. Regenerate card registries.
5. Validate generated card definitions.
6. Run `mage lint` and fix every reported issue before considering code complete.
7. Produce Markdown and JSON unsupported-card reports.
8. Add a missing-functionality rollup that groups reusable rules/parser gaps and
   lists which cards need each gap filled.

## Features

- **Manifest-driven workflow** - Resumable `.cardwork/<run>/cards.json` state.
- **Small batch rollout** - Conservative subagent batches keep failures isolated.
- **Attempt then validate** - Cards are not considered supported until
  `cardbatch validate` passes.
- **Current card model** - Implementations use `mtg/game/types` for card,
  supertype, and subtype values; `mtg/game/compare` for integer predicates; and
  optional `CardDef.Back` for double-faced back-face data.
- **Canonical source format** - Card source follows the expanded/raw-string layout
  shown in `mtg/cards/k/karplusan_forest.go`. The generator produces this layout;
  subagents must preserve it and fill categorized ability fields, not the legacy
  `Abilities` slice.
- **Unsupported reports** - Reports fetch errors, missing files, pending
  validation, and static validation failures.
- **Functionality rollup** - Groups missing rules/parser capabilities and lists
  the cards and oracle clauses that would use each capability, including cards
  that validate only because they delegate behavior to `ImplementationID`.

## Examples

See [`examples/basic-usage.md`](examples/basic-usage.md).
