# cardgen

Package `cardgen` is the isolated home for **Card Generation** tooling. It turns
Scryfall bulk data and Oracle text into executable `game.CardDef` Go source for
the Card Registry. Runtime game, rules, registry, and simulation behavior live
outside this directory.

There is one generation path:

```text
Scryfall JSON
  -> Oracle recognition
  -> typed game values
  -> CardDef validation
  -> deterministic Go source
```

The compiler is fail-closed. It emits a card only when every face, ability,
semantic element, and meaningful source token is supported. Unsupported cards
receive source-spanned diagnostics; `cardgen` never emits TODOs, partial ability
data, or guessed behavior.

Before compilation, `CorpusPolicy` limits the working corpus to cards that are
legal, restricted, or banned in Standard, Pioneer, Modern, Legacy, Pauper,
Vintage, or Commander. Playable paper token definitions are retained as a
special exception. Alchemy, digital-only identities, memorabilia, illegal
Un-set cards, minigames, art-series records, emblems, planes, schemes, and
Vanguard cards are excluded with explicit report reasons.

## Compiler stages

1. **Recognition (`cardgen/oracle`).** The lexer and parser preserve exact source
   spans. The semantic compiler recognizes costs, targets, triggers, keywords,
   Saga chapter headings, references, and ordered effects conservatively.
2. **Typed lowering (`lower.go` and `executable.go`).** Recognized semantics
   become typed `game.*` ability values, including chapter-numbered
   `game.ChapterAbility` values and the `game.ReadAheadStaticBody` Saga keyword
   template. `assembleCardDefs` combines
   those values with printed Scryfall fields and calls
   [`game.ValidateCardDef`](../mtg/game/README.md#carddef-structural-validation).
   Parameterized Kicker, Madness, Morph, Disguise, Mutate, and Toxic lines lower
   into their corresponding sealed `game.KeywordAbility` values; unsupported
   parameter forms remain fail-closed. Exact "Whenever this creature mutates"
   triggers lower to `game.EventPermanentMutated`. Exact static power/toughness bonuses may
   also grant supported keywords through separate layer-6 and layer-7
   continuous effects. Standalone keyword grants to supported controlled,
   creature-subtype-filtered, and attached permanent groups lower to layer-6
   continuous effects. Exact
   source-relative keyword grants gated by controlling supported permanent
   types, subtypes, colors, or colorless permanents use condition-gated layer-6
   effects. Exact `X` quantities, supported count/life/opponent/source-power
   formulas, and common target
   restrictions lower into runtime quantities and structured target predicates.
   Ordered effect clauses retain independent target specifications and references.
   Exact fixed, `X`, and supported dynamic placement of recognized named
   counters lowers from supported spell, activated, loyalty, triggered,
   ordered-effect, and Saga chapter bodies into typed `game.AddCounter`
   permanent instructions or `game.AddPlayerCounter` instructions for poison,
   energy, and experience. Counter kinds and target domains are checked
   strictly. Stun and finality placement remain fail-closed until their
   mandatory runtime mechanics are
   implemented ([#222](https://github.com/natefinch/council4/issues/222),
   [#223](https://github.com/natefinch/council4/issues/223)). Self-enter triggers support exact intervening
   conditions for kicked or cast entry and controlling one
   permanent of a named permanent card type. Non-self permanent
   enters-the-battlefield triggers lower for exact single-subject
   (`a`/`an`/`another`, optional `nontoken` qualifier) and `one or more`
   subject forms, with optional permanent type filter (creature, artifact,
   enchantment, land, planeswalker, or unfiltered) and optional you-control or
   opponent-controls controller constraints. Phase and step triggered abilities
   using `At the beginning of …` lower for the ten exact step-trigger phrases:
   your upkeep, each upkeep, each player's upkeep, each opponent's upkeep, your
   end step, each end step, each player's end step, combat on your turn, each
   combat, and your draw step. All other step-trigger phrases and all
   intervening-if conditions on step triggers are fail-closed. Exact
   self-dies triggers support exact `if it had no +1/+1 counters` and
   `if it had no -1/-1 counters` conditions using the departed permanent's
   last-known information. Fixed-damage bodies preserve that permanent as the
   damage source through an event reference. Exact event-card references can
   return the departed card from its owner's graveyard to hand or grant its
   Adventure face a graveyard-cast permission through the end of its
   controller's next turn. Ordinary battlefield activations
   lower exact mana, tap, untap, sacrifice, discard, pay-life, source-exile,
   graveyard-exile, and source-counter-removal costs into typed payment data.
   Exact trailing activation restrictions lower to typed sorcery, combat,
   upkeep, and once-per-turn timing checks.
   Common enters-tapped life, opponent-count, land-count, and
   basic-land-subtype conditions lower into typed replacement predicates.
   Exact optional pay-2-life and reveal-a-land-or-creature-subtype entry
   wordings lower into typed resolution payments for enters-tapped
   replacements.
3. **Rendering (`render.go`).** `Renderer.RenderCardSource` walks only validated
   typed values, derives imports from those values, and emits byte-deterministic,
   gofmt-stable Go source.

The bulk compiler detects distinct Oracle cards that map to the same filename or
Go identifier and appends a stable Scryfall-derived suffix to both generated
identities. Playable tokens always use `mtg/cards/tokens/<letter>` and include
their complete normalized Oracle UUID in both filename and Go identifier.
Printed `CardDef.Name` values remain unchanged.

`mtg/game` owns typed Card Definition data and structural validity;
`mtg/rules` owns behavior; `cardgen` owns recognition, lowering, and rendering.
See [ADR 0008](../docs/adr/0008-typed-ir-lowering.md).

## Usage

Compile the Scryfall Oracle Cards corpus into a temporary Card Registry tree:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/generated-cards \
  -report .cardwork/oracle-compile-report.json
```

After inspecting and validating the temporary tree, use `-out mtg/cards` only
when intentionally updating repository card definitions. The command
regenerates affected letter-package `cards.go` files.

Cards outside compiler coverage remain unsupported by Card Generation. Truly
exceptional mechanics may still use a hand-written **Card Implementation** with
an `ImplementationID`; that escape hatch is independent of this compiler.

## Tooling layout

- `cardgen/oracle`: lexer, parser, semantic compiler, corpus checks, and bulk
  compilation command.
- `cardgen`: typed lowering, Card Definition assembly, deterministic rendering,
  Scryfall data shapes, and source naming helpers.
- `cardgen/cmd/gencardlist`: `go generate` helper that writes each
  `mtg/cards/<letter>/cards.go` Card Registry list.

## Supported layouts

The source generator can represent Scryfall `normal`, `token`, `leveler`,
`saga`, `class`, `case`, `prototype`, `host`, `augment`, `emblem`, `mutate`,
`planar`, `scheme`, `vanguard`, `transform`, `modal_dfc`, `meld`,
`double_faced_token`, and `reversible_card` layouts. Corpus policy is narrower:
it excludes nonstandard game objects such as emblems, planes, schemes, and
Vanguard cards before source generation.

Transform, modal DFC, and double-faced token cards emit front-face fields on
`CardDef` and an optional `Back` face. Meld cards emit their front card with
`LayoutMeld`; complete meld behavior remains rules work. Reversible cards emit
one Card Definition per playable side.

## Key interfaces

- `GenerateExecutableCardSource(card, pkgName)` recognizes, lowers, validates,
  and renders a card, or returns diagnostics without source.
- `ExecutableGenerator` configures source identity disambiguation for bulk
  generation.
- `Renderer.RenderCardSource(card, defs, hints, pkgName)` renders validated typed
  Card Definitions deterministically.
- `ParseTypeLine(typeLine)` splits a type line into supertypes, types, and
  subtypes.
- `GeneratedIdentity` selects a generated card's category, package, filename,
  variable name, and migration path. `CardNameToVarName`,
  `CardNameToFileName`, and `CardNameToPackageLetter` provide its component
  naming rules.

Prepare layouts use `CardFace.EntersPrepared` on the creature face and
`CardDef.Alternate` for the spell face. The generator accepts them only when
both faces and the exact enters-prepared ability are fully lowerable.

Current executable mechanic coverage and the corpus support count live in
[`oracle/README.md`](oracle/README.md). The numbered expansion checklist lives
in [`../docs/oracle-compiler-expansion.md`](../docs/oracle-compiler-expansion.md).
