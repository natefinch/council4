# cardgen

Package `cardgen` is the isolated home for card-generation tooling. It fetches
card data from the Scryfall API, generates partial `game.CardDef` Go source
files for the council4 card registry, and owns supporting generator commands.
Runtime game, rules, card registry, and simulation code live outside this
directory.

## Validation ownership

Structural CardDef validation — nil card, missing name, oracle text without abilities, TargetSpec bounds, target index range, keyword variant checks, condition references, continuous effects, instruction sequences, and nested ability walks — is owned by [`game.ValidateCardDef`](../mtg/game/README.md#carddef-structural-validation) in the `mtg/game` package.

`cardgen.ValidateCard` and `cardgen.ValidateCards` are thin adapters: they call `game.ValidateCardDef`, map each `game.CardDefIssue` to a `ValidationIssue` with the card name added, and then apply the policy-only checks that belong to the tooling layer:

| Code | Owned by |
|------|----------|
| `nil-card` through `invalid-ability-body` | `game.ValidateCardDef` (structural) |
| `unregistered-implementation` | `cardgen` — depends on `ValidationOptions.KnownImplementationIDs` |
| `implementation-required` | `cardgen` — depends on `ValidationOptions.ReportImplementationIDs` |
| `generated-card-not-found` | `cardgen` — runtime/tooling policy |
| `validation-run-failed` | `cardgen` — tooling error reporting |

`ValidateCard(card, opts)` and `ValidateCards(cards, opts)` preserve the exact `CardName`, `FaceName`, `Path`, `Code`, and `Message` fields expected by existing tests and batch reports.

## What it does

Given a Magic: The Gathering card name, the library:

1. Fetches the card's data from Scryfall's `/cards/named` API endpoint.
2. Parses the mechanical fields: name, mana cost, colors, color identity, types, subtypes, supertypes, power/toughness, loyalty, defense, simple ETB tapped text, and double-faced card back faces. `game.CardDef` derives mana value from the generated mana cost.
3. Generates a Go source file with a `game.CardDef` literal in the canonical expanded/raw-string layout, leaving the categorized ability fields empty for LLM completion. See `mtg/cards/k/karplusan_forest.go` for the canonical format.
4. Validates generated card definitions against the currently executable rules
   model before batch workflows mark them supported, including structured
   object/card references, token-copy specs, library-to-battlefield searches,
   object-power dynamic amounts, and rule-effect primitives.
5. Reports cards that still rely on missing rules/parser functionality, including
   generated-source `Missing primitives` comments and `ImplementationID` escape
   hatches.

## Executable lowering pipeline (typed intermediate representation)

`GenerateExecutableCardSource` does not build Go source by concatenating
strings. It lowers Oracle text into a typed **intermediate representation
(IR)** of `game.*` ability values, validates an assembled `game.CardDef`, and
only then renders deterministic Go source. The pipeline has three distinct
stages, each owned by a different layer:

1. **Recognition / lowering (`cardgen/lower.go`).** `lowerExecutableFaces`
   compiles Oracle text and dispatches each recognized ability to a `lowerXxx`
   helper that returns a typed `game.*` value (for example
   `lowerTapManaAbility` returns a `game.ManaAbility`, `lowerSpell` returns a
   `game.AbilityContent`). The per-face result is a `loweredFaceAbilities`
   holding the categorized typed values in Oracle order. Unsupported text
   produces a source-spanned `oracle.Diagnostic` instead of a value.
2. **Assembly + validation (`cardgen/executable.go`).** `assembleCardDefs`
   combines parsed Scryfall fields (mana cost, colors, types, P/T, oracle text)
   with the lowered typed abilities into one or more `game.CardDef` values, then
   calls [`game.ValidateCardDef`](../mtg/game/README.md#carddef-structural-validation).
   Any structural issue is converted to a diagnostic and the card is failed
   before any source is emitted.
3. **Deterministic rendering (`cardgen/render.go`).** `Renderer.RenderCardSource`
   walks the validated typed values and emits Go source. It never iterates maps
   for ordering, sorts imports, detects needed packages from the rendered
   values, and produces byte-identical output for identical input. Sealed
   interfaces are rendered by switching on the value's `Kind()` and performing a
   single type assertion per case — never a Go type switch.

The handwritten **Card Implementation escape hatch** is preserved: cards whose
mechanics the lowering layer does not recognize are reported as unsupported and
fall back to a hand-written `game.CardDef` with an `ImplementationID`, exactly
as before. The typed pipeline only owns cards it can fully recognize, assemble,
and validate.

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

`cardgen` accepts Scryfall `transform`, `modal_dfc`, `double_faced_token`, `meld`, and `reversible_card` layouts. Transform, modal DFC, and double-faced token cards emit front-face fields on `CardDef`, `Layout`, and an optional `Back` `game.CardFace`. Meld cards generate the front card with `LayoutMeld`; full melded-permanent behavior is rules/card-implementation work. Reversible cards generate separate `CardDef` variables for each playable side rather than a face-selectable card.

## Key functions

- `FetchCard(name string)` — fetches a card from Scryfall by exact name.
- `GenerateCardSource(card, pkgName)` — generates Go source for a `CardDef`.
- `GenerateExecutableCardSource(card, pkgName)` — lowers each face to a typed
  intermediate representation, assembles and `game.ValidateCardDef`-validates a
  `game.CardDef`, and renders deterministic source via `Renderer`. It emits
  source only when every face is fully supported by the strict executable
  backend; otherwise it returns source-spanned diagnostics identifying the
  unsupported ability kind, keyword, parameter, mixed rules text, or structural
  validation failure. The renderer emits canonical `mtg/game` mechanic
  templates for exact Ward, Cycling, Equip, Enchant, color-based Protection,
  and tap-for-mana abilities rather
  than expanding their coordinated costs, zones, targets, keyword metadata,
  choices, and instructions.
  Supported executable mechanics currently include plain keywords, mana-cost
  Ward, Cycling, Equip, base-type Enchant, color-based Protection, supported
  tap mana choices, unconditional enters-tapped replacements, fixed single-target
  damage, destruction, exile, return-to-hand, and power/toughness changes,
  narrow mass destruction, fixed draw and life changes, fixed controller scry,
  fixed controller or target-player discard and mill, one-target tap and untap,
  and simple self-enter or self-dies triggers containing exactly one supported
  effect. Spells may contain an ordered sequence of independently supported
  sentence-sized effects, with at most one targeted clause until target-index
  remapping is available.
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
- `ParseTypeLine(typeLine)` — splits a type line into supertypes, types, subtypes.
- `SubtypeToLiteral(subtype, types)` — emits central `game.*Subtype*` constants when known.
- `CardNameToVarName(name)` — converts card names to Go exported variable names.
- `CardNameToFileName(name)` — converts card names to snake_case file names.
- `CardNameToPackageLetter(name)` — returns the first letter for sub-package routing.
- `cardgen/cmd/gencardlist` — scans a letter package and regenerates its
  `Cards` slice.
