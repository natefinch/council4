---
name: card-rollout
description: >
  Orchestrate a batch rollout of Magic card implementations from a card-list
  file. Use when the user says "card rollout", "batch implement cards",
  "process this card list", "run card-impl over this list", or asks to turn a
  list of cards into supported council4 CardDef implementations with an
  unsupported-card report.
---

# Card Rollout

Use this skill to take a disk-backed list of Magic cards, fetch oracle text,
attempt supported `CardDef` implementations in small batches, validate the
results, and produce an unsupported-card report for follow-up rules work. The
report must also summarize missing rules/parser functionality by capability and
list which cards would benefit from each capability.

This skill is orchestration. The deterministic Go tooling lives under
`cardgen/`; the per-card oracle parsing lives in the `card-impl` skill.

## When to Use This Skill

Use this skill when the user asks to:

- batch implement cards from a list on disk;
- run `card-impl` over a card list;
- process a Commander staples/test-deck card list;
- produce a report of cards blocked by missing rules support;
- continue a previous `.cardwork` card rollout.

Do not use this skill for one-off card implementation. Use `card-impl` directly
for a single named card.

## Inputs

Required:

- A path to a card-list file, or an existing `.cardwork/.../cards.json`
  manifest.

Optional:

- Batch size. Default to 3 cards per implementation batch.
- Work directory. Default to `.cardwork/<list-name>/`.
- Model tier. Default to a cheap mini-tier model for orchestration and a
  stronger coding model for card implementation subagents.
- Whether to commit after each clean batch. Default: ask before committing unless
  the user explicitly requested commits.

## Model Guidance

The rollout has two different reasoning profiles:

- **Orchestration:** parsing manifests, running `cardbatch`, splitting batches,
  validating, and reporting. Use a cheaper mini-tier model by default. The
  current recommended default is `gpt-5.4-mini`.
- **Card implementation:** editing generated Go files from oracle text. Use a
  stronger coding model by default because mistakes can create plausible-looking
  but invalid card definitions. The current recommended default is
  `claude-sonnet-4.6`.
- **Escalation:** if a mini-tier implementation subagent fails validation, edits
  the wrong files, invents unsupported primitives, or cannot classify a card,
  retry that card with the stronger coding model before calling it unsupported.
- **Avoid for orchestration:** `claude-haiku-4.5` was observed to follow the
  broad workflow but lose important command details in dry-run testing. Do not
  use it as the default rollout orchestrator unless a human is closely reviewing
  every command.

When spawning subagents, set the model explicitly where the tool supports it.
Use mini-tier models for report summarization or manifest-only review; use the
stronger coding model for any subagent expected to edit `mtg/cards/**`.

## Workflow

### Step 1: Check the workspace

1. Run `git status --short`.
2. If there are unrelated dirty files, do not overwrite or revert them. Continue
   only if the rollout can avoid those files; otherwise ask the user how to
   proceed.
3. Confirm the card list or manifest path exists.

### Step 2: Build or refresh the manifest

For a card-list file:

```bash
go run ./cardgen/cmd/cardbatch parse \
  -in <card-list.txt> \
  -out .cardwork/<run>/cards.json

go run ./cardgen/cmd/cardbatch fetch \
  -manifest .cardwork/<run>/cards.json \
  -cache .cardwork/<run>/cache/scryfall
```

For an existing manifest, skip `parse` and run `fetch` only if rows still need
Scryfall data.

Then mark generated source presence:

```bash
go run ./cardgen/cmd/cardbatch missing \
  -manifest .cardwork/<run>/cards.json \
  -repo .
```

### Step 3: Create a small worklist

Use conservative batches so failures stay isolated:

```bash
go run ./cardgen/cmd/cardbatch worklist \
  -manifest .cardwork/<run>/cards.json \
  -repo . \
  -limit 3
```

If the worklist is empty, skip to validation/reporting.

### Step 4: Spawn implementation subagents

Split the worklist into batches of 3-5 cards. Prefer 3 when the cards are
complex, unfamiliar, or have long oracle text.

For each batch, launch a subagent with complete context:

```text
You are implementing council4 Magic card definitions.

Model: use claude-sonnet-4.6 by default for implementation batches. If the user
explicitly requested a cheaper trial, use gpt-5.4-mini for simple cards only and
escalate failed validation to claude-sonnet-4.6.

Cards:
- <Card 1>
- <Card 2>
- <Card 3>

For each card:
1. Invoke/use the project card-impl workflow.
2. Generate the mechanical CardDef source with:
   go run .agents/skills/card-impl/main.go "<Card Name>"
3. Read .agents/skills/card-impl/CARD-IMPLEMENTATION-GUIDE.md.
4. Fill Abilities only using existing game/rules primitives.
5. Use current generated-card conventions: card/super/subtype vocabulary comes
   from `mtg/game/types` (`types.Card`, `types.Super`, `types.Sub`), integer
   comparisons use `mtg/game/compare`, and double-faced cards use front-face
   `CardDef` fields plus optional `Back: opt.Val(game.CardFace{...})`, not a
   `Faces` slice.
6. If the card cannot be represented, leave the safest generated state and
   explain what rules support is missing. Do not invent enum values.
7. Run gofmt on edited card files.

Return changed files, implemented cards, skipped/blocking cards, and any missing
rules primitives. For every missing primitive, use this format so the rollout
report can group repeated gaps:

- Capability: <short reusable rules/parser capability, not a card-specific fix>
- Cards: <card names in this batch that need it>
- Oracle clauses: <short quoted clauses that require it>
- Current workaround: <ImplementationID, partial declarative approximation, or unsupported>

Do not commit.
```

Do not have two subagents edit the same card file or the same letter-package
`cards.go` file concurrently. `go generate` should be run centrally after
subagents complete, not inside parallel subagents unless each owns a disjoint
working copy.

### Step 5: Regenerate and validate

After each batch returns:

```bash
go generate ./mtg/cards/...

go run ./cardgen/cmd/cardbatch validate \
  -manifest .cardwork/<run>/cards.json \
  -repo .
```

Then run the smallest relevant tests first:

```bash
go test ./cardgen/... ./mtg/cards/...
```

Escalate to full validation before final completion:

```bash
go test ./... && go vet ./...
```

### Step 6: Report unsupported or pending cards

Generate both report formats:

```bash
go run ./cardgen/cmd/cardbatch report \
  -manifest .cardwork/<run>/cards.json \
  -repo . \
  -md .cardwork/<run>/unsupported.md \
  -json .cardwork/<run>/unsupported.json
```

Read the Markdown report before summarizing. Use it to distinguish:

- fetch/layout/card-list errors;
- cards still missing generated files;
- cards with generated files but pending validation;
- cards with invalid generated definitions;
- likely rules/parser follow-up areas.

Then add a **Missing functionality rollup** to the Markdown report. Use the
subagent batch summaries, validation issues, generated-card comments, and
`ImplementationID` notes to group gaps by reusable capability rather than by
card. Include this rollup even when `cardbatch report` says there are zero
unsupported cards, because cards may validate while still relying on
`ImplementationID` escape hatches or partial declarative approximations.

For each capability, include:

- the concise capability name;
- affected cards;
- the oracle clauses or behavior that require it;
- whether the current implementation is blocked, approximated, or delegated to
  `ImplementationID`;
- likely code area when known, such as `game.Effect`, `TargetSpec`,
  `TriggerPattern`, replacement effects, mana choices, or parser mapping.

Use stable, reusable capability names so future reports can merge repeated
needs. Prefer names like `equipped-creature selector`,
`controller controls land subtype condition`, `commander color identity mana
choice`, or `shuffle permanent into owner's library` over one-off names like
`Basilisk Collar support`.

Append the section after the generated unsupported-card details:

```markdown
## Missing functionality rollup

### Equipped-creature selector

- Cards: Basilisk Collar, Blazing Sunsteel
- Needed for: "Equipped creature has ..."; "Equipped creature gets ..."
- Current state: delegated to `ImplementationID`
- Likely area: effect selectors / attachment-aware continuous effects

### Commander color identity mana choice

- Cards: Command Tower
- Needed for: "{T}: Add one mana of any color in your commander's color identity."
- Current state: approximated as any-color mana plus `ImplementationID`
- Likely area: mana choice resolution / commander metadata
```

### Step 7: Iterate or stop

If the user asked for a full rollout, repeat worklist -> implementation ->
validation -> report until:

- the worklist is empty;
- remaining cards are blocked by missing rules/parser support;
- tests fail in a way that needs user direction;
- the user-specified batch/iteration limit is reached.

## Output Format

End with a concise summary:

```text
Implemented:
- Card A (`mtg/cards/a/card_a.go`)
- Card B (`mtg/cards/b/card_b.go`)

Still unsupported:
- Card C — oracle-without-abilities; needs parser/rules support for ...
- Card D — unsupported SearchSpec; needs richer tutor modeling

Missing functionality:
- Equipped-creature selector — Basilisk Collar, Blazing Sunsteel
- Commander color identity mana choice — Command Tower

Reports:
- .cardwork/<run>/unsupported.md
- .cardwork/<run>/unsupported.json
```

Mention commits only if commits were made or explicitly requested.

## Error Handling

### Scryfall fetch failed

- Keep the manifest row as `fetch-error`.
- Do not try to implement that card from memory.
- Include the fetch error in the final report.

### `card-impl` generated a scaffold but validation failed

- Do not call the card supported.
- Leave the validation issue in the manifest.
- Report the card as unsupported with the validation reason.

### Generated package does not compile

- Stop parallel rollout.
- Fix the compile error if it is local to the current batch.
- Re-run `go generate ./mtg/cards/...` and `cardbatch validate`.
- If the package remains broken, produce a report and stop.

### Worktree has unrelated user changes

- Do not revert or overwrite them.
- Avoid those files if possible.
- Ask the user before proceeding if the rollout must touch the same files.

## Boundaries

This skill will:

- orchestrate `cardbatch` and `card-impl`;
- spawn implementation subagents for small batches;
- keep batch state under `.cardwork/`;
- validate generated card definitions before calling them supported;
- produce unsupported-card reports.

This skill will not:

- invent new rules enum values;
- mark a card supported just because generation compiled;
- directly modify runtime rules to support blocked cards unless the user asks;
- run unbounded parallel agents over a large card list;
- use cheap implementation models after validation failures without escalation;
- treat `.cardwork` as source of truth over generated files and validation.

## Quick Reference

```bash
go run ./cardgen/cmd/cardbatch parse -in cards.txt -out .cardwork/run/cards.json
go run ./cardgen/cmd/cardbatch fetch -manifest .cardwork/run/cards.json
go run ./cardgen/cmd/cardbatch worklist -manifest .cardwork/run/cards.json -repo . -limit 3
go run .agents/skills/card-impl/main.go "Card Name"
go generate ./mtg/cards/...
go run ./cardgen/cmd/cardbatch validate -manifest .cardwork/run/cards.json -repo .
go run ./cardgen/cmd/cardbatch report -manifest .cardwork/run/cards.json -repo .
```
