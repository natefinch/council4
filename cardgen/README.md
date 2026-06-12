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
semantic element, and meaningful source token is supported. Trigger wording is
recognized here into a source-spanned `oracle.TriggerPattern` with closed
semantic event, relation, Selection, zone, step, combat, batching, and
intervening-condition vocabulary. Exact condition wording is recognized once
into a closed, source-spanned semantic predicate and exact object wording is
bound conservatively to its source, target occurrence, triggering event
subject, or prior instruction result. Ambiguous and unsupported references
remain explicit semantic values. The retained raw text is used only for
diagnostics and exact source consumption. Unsupported cards
receive source-spanned diagnostics; `cardgen` never emits TODOs, partial ability
data, or guessed behavior.

Trigger recognition uses a small registry of exact event-family
templates. Permanent zone-change, spell/ability, combat, phase/step, permanent
state, and player-event wording variants share templates that bind only the
closed semantic slots above; ambiguous or unsupported slot text fails closed.

Before compilation, `CorpusPolicy` limits the working corpus to cards that are
legal, restricted, or banned in Standard, Pioneer, Modern, Legacy, Pauper,
Vintage, or Commander. Playable paper token definitions are retained as a
special exception. Alchemy, digital-only identities, memorabilia, illegal
Un-set cards, minigames, art-series records, emblems, planes, schemes, and
Vanguard cards are excluded with explicit report reasons.

## Compiler stages

1. **Recognition (`cardgen/oracle`).** The lexer and parser preserve exact source
   spans. The semantic compiler recognizes costs, targets, triggers, keywords,
   Saga chapter headings, references, and ordered effects conservatively. Reusable
   body content (targets, conditions, effects, keywords, references, nested modes)
   is grouped into `oracle.AbilityContent`; each `oracle.CompiledAbility` and
   `oracle.CompiledMode` carries one `oracle.AbilityContent` value alongside its
   shell-specific fields (cost, trigger clause, loyalty change, chapter numbers,
   text, span, optional flag). Static wording is recognized separately into one
   or more source-spanned `oracle.StaticDeclaration` values because declarations
   never resolve and are not Instructions. A declaration carries a closed group
   domain plus Selection, optional shared condition, and a typed continuous
   layer operation, rule domain and operation, cost modifier, or non-battlefield
   card-ability grant. Unsupported groups, conditions, durations, operations,
   and shells remain explicit capability blockers.
2. **Typed lowering (`lower.go`, `activation.go`, `static_declaration.go`, `condition.go`,
   `reference.go`, `trigger_pattern.go`, and `executable.go`).**
   `lowerTriggerPattern` is the single mechanical adapter from
   `oracle.TriggerPattern` to `game.TriggerPattern`; trigger shell lowerers never
   interpret raw event-clause text. `lowerAbilityContent`
   is the single entry point that lowers an `oracle.AbilityContent` value into
   `game.AbilityContent`. All supported shells — spell, activated body, triggered
   body, loyalty body, chapter body, modal option, and ordered-effect clauses —
   call `lowerAbilityContent` directly; no shell lowerer constructs a fake spell
   ability to reach body lowering. `condition.go` is the single
   `oracle.CompiledCondition` to `game.Condition` adapter and requires an
   explicit static, activation, replacement, or intervening-trigger context.
   `reference.go` is the single adapter from bound semantic references to typed
   runtime object and card references, including event-permanent LKI and linked
   prior-instruction results. `activation.go` composes the generic activated
   shell from typed cost components, timing, zone of function, activation
   condition, bound references, and shared Ability Content. Mana and non-mana
   activated abilities use that same shell preparation while retaining distinct
   runtime types. Known shell failures report activation cost, timing, zone,
   condition, reference, mode, or structure diagnostics instead of a generic
   activated-ability failure. `static_declaration.go` is the single mechanical
   adapter from semantic Static Declarations to `game.StaticAbility`,
   `game.ContinuousEffect`, `game.RuleEffect`, and `game.CostModifier` values.
   Mixed static paragraphs lower through that adapter as multiple declarations
   sharing one runtime static ability. Recognized semantics
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
   permanent of a named permanent card type. Permanent zone-change triggers
   share one lowering path for self, attached, single-subject, and `one or more`
   enter, die, leave, exile, return-to-hand, and battlefield-to-graveyard
   clauses. Exact patterns may bind controller and owner relations, origin and
   destination zones, self exclusion, face-down state, and event-subject
   Selection predicates for type unions, supertypes, subtypes (including
   Outlaw), colors, token state, tapped state, combat state, keywords, mana
   value, power, and toughness. `Leaves ... without dying` excludes the
   graveyard destination. Phase and step triggered abilities
   using `At the beginning of …` lower for
   exact supported controller-relative upkeep, draw, end, combat, combat-step,
   and main-phase variants, including steps belonging to the controller of an
   enchanted permanent. Combat templates bind named/self/attached and semantic
   Selection subjects, the other blocking combatant, attacked player or
   permanent recipients, damage-source and damage-recipient Selections,
   combat/noncombat qualifiers, and exact player relations. Player-level attack
   wording and `one or more` attack, block, and combat-damage wording lower only
   through declaration/damage batch IDs, with per-attack-target batching where
   Oracle semantics require it. Compound events, temporal qualifiers, and
   unavailable Selection predicates remain fail-closed with missing-event or
   missing-runtime-capability diagnostics. Exact
   permanent-tapped, permanent-untapped, and turned-face-up action triggers
   share the semantic Trigger Pattern path; face-up triggers may bind self,
   attached, controller-relative, and Selection-filtered subjects. Became-target
   patterns bind the targeted subject's controller independently from the
   targeting spell or ability's controller. Player action templates include
   controller-relative and any-player Cycling events. Sacrifice triggers bind
   the sacrificing player independently from the sacrificed permanent's shared
   Selection subject. Scry and surveil use distinct player-action Trigger
   Pattern events. Activated-ability patterns bind the activating player and
   source-permanent Selection, but lower only when they explicitly exclude mana
   abilities; unrestricted forms fail closed until payment-time mana
   activations join the authoritative event stream.
   Supported draw, life-gain/loss, scry, and surveil patterns may also bind the
   affected player's exact event ordinal during the current turn.
   self-dies triggers support exact `if it had no +1/+1 counters` and
   `if it had no -1/-1 counters` conditions using the departed permanent's
   last-known information. Fixed-damage bodies preserve that permanent as the
   damage source through an event reference. Exact event-card references can
   return the departed card from its owner's graveyard to hand or grant its
   Adventure face a graveyard-cast permission through the end of its
   controller's next turn. Spell-cast triggered abilities using `Whenever ...
   casts ...` lower for three exact player prefixes (`you cast`,
   `a player casts`, `an opponent casts`) and seventeen exact spell phrases:
   `a spell` (wildcard), `a noncreature spell`, `a creature spell`,
   `an instant or sorcery spell`, `an instant spell`/`an instant`,
   `a sorcery spell`, `an artifact spell`, `an enchantment spell`,
   `a land spell`, `a planeswalker spell`, `a noncreature, nonland spell`, and single-color forms
   `a white/blue/black/red/green spell`. Self-cast (`when you cast this spell`),
   `TriggerWhen`, unsupported intervening-if conditions, unknown or non-exact
   ability-word forms, modes, and all other spell-phrase forms are fail-closed.
   Exact Threshold, Delirium, Domain, Metalcraft, Hellbent, Ferocious, and Coven
   conditions lower into typed live-state predicates and dynamic amounts.
   Ordinary battlefield activations
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
