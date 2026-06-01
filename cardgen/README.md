# cardgen

Package `cardgen` is the isolated home for card-generation tooling. It fetches
card data from the Scryfall API, generates partial `game.CardDef` Go source
files for the council4 card registry, and owns supporting generator commands.
Runtime game, rules, card registry, and simulation code live outside this
directory.

## What it does

Given a Magic: The Gathering card name, the library:

1. Fetches the card's data from Scryfall's `/cards/named` API endpoint.
2. Parses the mechanical fields: name, mana cost, mana value, colors, color identity, types, subtypes, supertypes, power/toughness, loyalty, defense, simple ETB tapped text, and double-faced card faces.
3. Generates a Go source file with a `game.CardDef` literal, leaving the `Abilities` slice empty for LLM completion.
4. Validates generated card definitions against the currently executable rules
   model before batch workflows mark them supported, including newer structured
   object/card references, token-copy specs, and rule-effect primitives.
5. Reports cards that still rely on missing rules/parser functionality, including
   generated-source `Missing primitives` comments and `ImplementationID` escape
   hatches.

## Usage

The library is typically used via the `card-impl` skill's `main.go`:

```bash
go run .agents/skills/card-impl/main.go "Lightning Bolt"
```

This creates `mtg/cards/l/lightning_bolt.go` with the mechanical fields populated.

## Tooling layout

- `cardgen` package: Scryfall fetch and `CardDef` source-generation helpers.
- `cardgen/cmd/cardbatch`: resumable batch workflow for parsing card lists,
  fetching Scryfall oracle data, identifying missing generated card files,
  printing small worklists for `card-impl`, validating attempted cards, and
  reporting unsupported cards.
- `cardgen/cmd/gencardlist`: `go generate` helper that writes each
  `mtg/cards/<letter>/cards.go` list.
- `.agents/skills/card-impl`: agent skill instructions and entrypoint. The skill
  stays outside `cardgen/`, but it imports this package for deterministic
  generation work.

## Double-faced layouts

`cardgen` accepts Scryfall `transform`, `modal_dfc`, `double_faced_token`, `meld`, and `reversible_card` layouts. Transform, modal DFC, and double-faced token cards emit `Layout` plus per-face `[]game.CardFace` data. Meld cards generate the front card with `LayoutMeld`; full melded-permanent behavior is rules/card-implementation work. Reversible cards generate separate `CardDef` variables for each playable side rather than a face-selectable card.

## Key functions

- `FetchCard(name string)` — fetches a card from Scryfall by exact name.
- `GenerateCardSource(card, pkgName)` — generates Go source for a `CardDef`.
- `ValidateCard(card, opts)` / `ValidateCards(cards, opts)` — report static
  support issues in generated `CardDef` values.
- `ParseCardList`, `NewManifestFromItems`, `FetchManifest`, and
  `MarkExistingFiles` — build and update card batch manifests.
- `MissingWorklist` and `ValidateManifestGeneratedCards` — support the
  attempt-then-validate batch workflow.
- `BuildUnsupportedReportWithSource`, `WriteUnsupportedReportMarkdown`, and
  `WriteUnsupportedReportJSON` — turn manifest failures and generated-source
  missing-functionality notes into human- and machine-readable reports.
- `ParseManaCostLiteral(cost)` — converts Scryfall mana cost strings to Go code.
- `ManaValueFromCost(cost)` — computes a face's mana value from a mana-cost string.
- `ParseTypeLine(typeLine)` — splits a type line into supertypes, types, subtypes.
- `SubtypeToLiteral(subtype, types)` — emits central `game.*Subtype*` constants when known.
- `CardNameToVarName(name)` — converts card names to Go exported variable names.
- `CardNameToFileName(name)` — converts card names to snake_case file names.
- `CardNameToPackageLetter(name)` — returns the first letter for sub-package routing.
- `cardgen/cmd/gencardlist` — scans a letter package and regenerates its
  `Cards` slice.
