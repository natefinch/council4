# Council4 — Game Engine

A pure Go data-structure scaffold for a 4-player [Commander](https://mtgcommander.net/) (EDH) Magic: The Gathering game engine. No UI, no network, no physical-card concerns — just the in-memory types needed to represent a complete game state.

## Quick Start

```bash
cd /path/to/council4
go test ./...
go vet ./...
```

Requires **Go 1.25.1+**. No external dependencies.

## Project Layout

```
mtg/game/                      # package github.com/natefinch/council4/mtg/game
│
├── game.go                    # Game struct, NewGame() constructor, top-level helpers
├── card.go                    # CardDef (static card template) & CardInstance (in-game card)
├── ability.go                 # Ability-related helpers, keywords, trigger conditions, modes
├── ability_body.go            # Sealed ability body variants and shared body helpers
├── mechanic_template.go       # Complete templates for invariant mechanics
├── condition.go               # Reusable Condition and PermanentFilter predicates
├── selection.go               # Selection — shared valence-agnostic characteristic matcher data
├── instruction.go             # Instruction, InstructionKey, Quantity, sequence validation
├── primitive.go               # Typed Primitive variants (the sole effect model)
├── reference.go               # ObjectReference, PlayerReference, and CardReference effect bindings
├── group_reference.go         # GroupReference — candidate-domain binding for mass effects
├── subtype.go                 # Central artifact/creature/enchantment/land subtype constants
├── choice.go                  # ChoiceRequest/ChoiceDecision for non-action decisions
├── permanent.go               # Permanent — battlefield state for a card or token
├── player.go                  # Player — life, zones, commander tracking, designations
├── zone/                      # Zone vocabulary and ordered card collections
├── stack.go                   # Stack (LIFO) & StackObject (spells/abilities resolving)
├── target.go                  # Target — runtime target choices for spells/abilities
├── event.go                   # Event — typed rules facts emitted by mtg/rules
├── continuous.go              # ContinuousEffect layers, CopyableValues, DynamicValue
├── cost_modifier.go           # CostModifier and AttackTax runtime cost data
├── duration.go                # EffectDuration and delayed-trigger data
├── lki.go                     # ObjectSnapshot and linked-object references
├── replacement.go             # PreventionShield, ReplacementDecision, ETB counter data
├── turn.go                    # Phase, Step, TurnState, TurnOrder (4-player rotation)
├── combat.go                  # AttackDeclaration, BlockDeclaration, CombatState
├── object.go                  # ObjectID, PlayerID — shared identity types
├── doc.go                     # Package-level documentation
│
├── action/                    # Player action data types
│   ├── README.md              #   Package guide
│   └── action.go              #   Action tagged struct, payloads, constructors
│
├── id/                        # Leaf package: unique ID generation
│   ├── README.md              #   Package guide
│   └── id.go                  #   ID type (uint64) + atomic Generator
│
├── mana/                      # Leaf package: produced mana and mana pools
│   ├── README.md              #   Package guide
│   ├── doc.go                 #   Package documentation
│   ├── color.go               #   Color enum (W, U, B, R, G, C)
│   ├── unit.go                #   Unit — spendable mana with provenance such as snow
│   └── pool.go                #   Pool — runtime mana tracking
│
└── counter/                   # Leaf package: counter types
    ├── README.md              #   Package guide
    └── counter.go             #   Kind enum (25 counter types), Set with +1/+1 ↔ -1/-1 cancel
```

### Why this layout?

MTG concepts (cards, permanents, abilities, zones, players, the stack) reference each other heavily. Splitting each into its own Go package would create circular imports. Instead:

- **Root `game` package** holds all interrelated types in well-organized files.
- **Leaf packages** (`id`, `mana`, `counter`) have zero game dependencies and can be imported by anyone without cycles.
- **Action data** lives in `action/` so the rules engine and agents can share one action representation without making `game` depend on rules behavior.

## Core Types

### Card Model (Three Layers)

```
CardDef  ──────▶  CardInstance  ──────▶  Permanent / StackObject
(static            (specific card         (in-play game object
 template)          in a game)             with mutable state)
```

| Type | File | Purpose |
|------|------|---------|
| `CardDef` | `card.go` | Immutable card template shared across games. It embeds `CardFace` for the printed front-face characteristics, keeps `ColorIdentity` and layout metadata on the physical card, and stores optional back-face data for double-faced layouts. |
| `CardInstance` | `card.go` | A specific card in a specific game. Has a unique `id.ID` and an `Owner`. |
| `Permanent` | `permanent.go` | A card or token on the battlefield — tapped, counters, damage, attachments, phased out, face-down origin state, current printed face, etc. |
| `StackObject` | `stack.go` | A spell or ability on the stack — selected/source face, source zone/card, face-down origin state, source ability index, targets, chosen modes, X value, additional costs, and linked resolution results. |
| `Target` | `target.go` | A runtime targeting choice: player, permanent, or stack object. |
| `ChoiceRequest` | `choice.go` | A bounded non-action decision such as trigger target choice, trigger ordering, or optional-effect yes/no. |
| `CardFace` | `card.go` | One printed face's characteristics: name, mana cost, colors, supertypes/types/subtypes, categorized abilities, P/T, loyalty, battle defense, replacement abilities, optional implementation ID, and oracle text. |
| `Condition` | `condition.go` | Reusable data-only predicates for static ability conditions, activation restrictions, intervening-if checks, effect conditions, and replacement effects. |
| `ObjectReference`, `PlayerReference`, `CardReference` | `reference.go` | Reusable resolution-time bindings for effects that need source objects, target-derived controllers/owners, linked objects/cards, damage sources, non-default recipients, or card-condition checks. Named constructors (e.g. `SourcePermanentReference`, `TargetPlayerReference`, `ObjectOwnerReference`) build every valid binding, and `Validate()` reports structural problems for `ValidateCardDef`. |
| `GroupReference` | `group_reference.go` | Pure data describing **where** a mass effect finds a group of permanents: a candidate domain (battlefield, attached object, object-controlled), a `Selection` that narrows it, and optional anchor/exclusion object references. The zero value is invalid. |

### Player

Each `Player` (`player.go`) tracks:

- **Life** (starts at 40), **poison counters**, **commander damage** received (per commander)
- **Commander tax** (cast count from command zone × 2) and Commander mulligan count
- **Five zones**: Library, Hand, Graveyard, Exile, Command Zone
- **Mana pool** (`mana.Pool`)
- Optional **PowerBracket** and **PowerLevel** metadata for later simulations/reports
- **Designations**: monarch, initiative, city's blessing, ring level, energy, experience

### Game

`Game` (`game.go`) ties everything together:

- `[4]*Player` — the four players
- `[]*Permanent` — shared battlefield
- `Stack` — LIFO spell/ability stack
- `CommanderIDs` — original commander card instances for commander damage and command-zone replacement
- `Events` — rules-relevant facts emitted by `mtg/rules` as mutations occur
- `TurnState` / `TurnOrder` — turn structure with eliminated-player handling
- `FailedDraws` — transient per-game flags for players who tried to draw from an empty library
- `*CombatState` — current attack/block declarations and combat assignment data (nil outside combat)
- `map[id.ID]*CardInstance` — registry of all card instances
- `NewGame(configs)` — constructor that sets up 40 life, creates card instances, places commanders, shuffles libraries
- `NewGameWithRand(configs, rng)` — deterministic constructor for tests and simulations

### Turn Structure

`TurnState` and `TurnOrder` (`turn.go`) model the full MTG turn:

```
Beginning ──▶ Precombat Main ──▶ Combat ──▶ Postcombat Main ──▶ Ending
  Untap            (lands,          BoC           (same as          End Step
  Upkeep            spells)         Attackers      precombat)       Cleanup
  Draw                              Blockers
                                    1st Strike Dmg
                                    Combat Dmg
                                    EoC
```

Priority and land-per-turn tracking are built in. `TurnOrder` handles the 4-player clockwise rotation and skips eliminated players.

`CombatState` is populated by the rules engine during combat. The current rules slice stores attack declarations against players, planeswalkers, or battles; block declarations; and blocker order for deterministic multi-block damage assignment. `Permanent.Goaded` stores which players have goaded a creature so the rules engine can enforce attack requirements during declare attackers. Combat damage, creature damage, and permanent death logs live in `mtg/rules.GameResult`.

### Abilities

`CardFace` exposes abilities through eight categorized fields: `SpellAbility opt.V[AbilityContent]`, `ActivatedAbilities []ActivatedAbility`, `ManaAbilities []ManaAbility`, `LoyaltyAbilities []LoyaltyAbility`, `TriggeredAbilities []TriggeredAbility`, `ChapterAbilities []ChapterAbility`, `ReplacementAbilities []ReplacementAbility`, and `StaticAbilities []StaticAbility`. A spell's rules text is already stored in `CardFace.OracleText`, so its resolving content needs no additional wrapper. Card source files populate these fields directly using typed struct literals. Every resolving ability uses `AbilityContent`: ordinary abilities use `game.Mode{...}.Ability()`, while modal abilities define `Modes`, `MinModes`, and `MaxModes` explicitly. Saga chapter abilities additionally record the lore-counter numbers that trigger their content.

Resolving spells and abilities store ordered `Instruction` values in `Mode.Sequence`. `AbilityContent.IsModal()` distinguishes a real mode choice from ordinary content containing one required mode. An `Instruction` combines one sealed, data-only `Primitive` variant with shared sequencing data: conditions, optionality, result publication, result gates, and diagnostics. Primitive structs expose only fields valid for that operation. For example, `Damage` accepts one `DamageRecipient`; `CreateToken` accepts one `TokenSource`; and `PutOnBattlefield` accepts one `BattlefieldSource`. `Quantity` represents either a fixed integer or a resolution-time `DynamicAmount`, preventing both forms from being configured simultaneously.

Resolution data uses separate key namespaces. `ResultKey` identifies instruction outcomes used by `InstructionResultGate` and previous-result quantities, `ChoiceKey` identifies values published by `Choose`, and `LinkedKey` identifies linked cards or objects published by reveal-like operations. `ValidateInstructionSequence` rejects nil primitives, invalid primitive references, forward or duplicate publications, and cross-namespace key use.

Static abilities do not resolve, so they do not use Instructions for continuous or rule-changing behavior. `StaticAbility.ContinuousEffects` and `StaticAbility.RuleEffects` declare that data directly; `mtg/rules` derives the active runtime effects while the static ability functions. The rules engine executes typed primitives directly.

When a card's oracle-text order differs from struct-field order (e.g. static keywords printed before triggered/activated abilities), use an immediately-invoked initializer function and `append` each ability to the appropriate field in oracle order — see `mtg/cards/k/karplusan_forest.go` for the canonical pattern. When consuming abilities in rules or tests, iterate the categorized fields directly or use `CardFace.AbilityCount()` and `CardFace.BodyAt()` for the stable canonical index order: Spell, Activated, Mana, Loyalty, Triggered, Chapter, Replacement, Static. Nested abilities granted by `ContinuousEffect.AddAbilities` also use typed `Ability` values.

| Kind | Categorized field | Marker | Example |
|------|------------------|--------|---------|
| Spell | `SpellAbility opt.V[AbilityContent]` | (text on instants/sorceries) | "Destroy target creature." |
| Activated | `ActivatedAbilities []ActivatedAbility` | `[Cost]: [Effect]` | "{T}: Add {G}." |
| Mana | `ManaAbilities []ManaAbility` | `[Cost]: Add {X}` (no targets) | "{T}: Add {G}." |
| Loyalty | `LoyaltyAbilities []LoyaltyAbility` | `[+N]/[−N]/[0]:` | "+1: …" |
| Triggered | `TriggeredAbilities []TriggeredAbility` | `When` / `Whenever` / `At` | "When this enters the battlefield…" |
| Chapter | `ChapterAbilities []ChapterAbility` | Roman numeral followed by `—` | "II, III — Draw a card." |
| Replacement | `ReplacementAbilities []ReplacementAbility` | "If … would … instead …" | "If ~ would die, instead …" |
| Static | `StaticAbilities []StaticAbility` | declarative | "Creatures you control get +1/+1." |

50+ keywords are enumerated (flying, haste, deathtouch, lifelink, indestructible, protection, flashback, cascade, discover, eternalize, morph, disguise, etc.). Keyword authoring uses sealed `KeywordAbility` variants on ability types that carry keywords (`StaticAbility`, `ActivatedAbility`, and `TriggeredAbility`), such as `SimpleKeyword`, `WardKeyword`, `EnchantKeyword`, `KickerKeyword`, `MorphKeyword`, `DisguiseKeyword`, `ToxicKeyword`, `SuspendKeyword`, and `ProtectionKeyword`. Plain non-parameterized keywords have reusable `StaticAbility` templates such as `DevoidStaticBody`, `FlyingStaticBody`, `HasteStaticBody`, `ReachStaticBody`, and `ExaltedStaticBody`; append them to `StaticAbilities` rather than mutating their fields. A Devoid card face must also have an empty `Colors` slice so its colorless characteristic is represented in every zone; its `ColorIdentity` remains unchanged. Complete invariant mechanics use the canonical templates in `mechanic_template.go`: `WardStaticAbility`, `EnchantStaticAbility`, `ProtectionFromColorsStaticAbility`, `CyclingActivatedAbility`, and `EquipActivatedAbility` coordinate their canonical text, costs, zones, targets, keyword metadata, and instructions; `TapManaAbility` and `TapManaChoiceAbility` coordinate tap costs, canonical text, resolution choices, and mana production; `CantBlockStaticBody`, `CantBeBlockedStaticBody`, `MustAttackStaticBody`, and `CantBeCounteredStaticBody` coordinate exact source-scoped rule declarations with their battlefield and stack zones. Prefer these templates over reconstructing their component fields. Rules code should use the shared helpers in `keyword.go`, such as `BodyHasKeyword`, `BodyKeywordAbility`, `BodyAddKeywordKindsTo`, `BodyWardCost`, `StaticBodyWardCost`, `ActivatedBodyCyclingCost`, `ActivatedBodyEquipCost`, `BodyMadnessCost`, `BodyToxicAmount`, `StaticBodyEnchantTarget`, `ActivatedBodyKicker`, `StaticBodyProtectionColors`, `StaticBodyMorphCost`, `StaticBodyDisguiseCost`, `StaticBodySuspendInfo`, and `EternalizeActivatedBody`.

`TargetSpec` supports min/max target counts, legacy natural-language `Constraint` text, broad target categories through `TargetAllow`, announcement-time `Chooser` values, and structured `TargetPredicate` data for common color, type, controller, tapped, combat-state, keyword, mana-value, P/T, and "another" filters. For opponent-chosen target slots, predicates such as `ControllerYou` are evaluated relative to the choosing opponent, so `Chooser: TargetChooserOpponent` plus `Controller: ControllerYou` means "a target controlled by that opponent." `Condition` and `PermanentFilter` model reusable controller-controls and referenced-object predicates for static ability conditions, activation restrictions, trigger intervening-if checks, instruction conditions, and replacement effects; event-permanent references may use last-known information. Conditions also cover life thresholds, alive-opponent counts, individual or collective opponent permanent counts, aggregate controlled-creature total power, class level gates, monstrous-state checks, max-speed checks, and event-permanent name uniqueness for cards such as Guardian Project.

Modal abilities use `Modes` plus `MinModes`, `MaxModes`, and `AllowDuplicateModes` for choose-one, choose-N, up-to-N, one-or-both, all-mode, and duplicate-mode templates. Exactly one mode with `MinModes: 1` and `MaxModes: 1` is non-modal and executes without recording a choice; `Mode.Ability()` constructs that form. `SearchSpec` supports library-to-hand and library-to-battlefield searches with card-type, supertype, subtype-any, reveal, shuffle, and enters-tapped options. `DynamicAmount` represents quantities determined on resolution, such as X, target characteristics, object power, selector counts, counter counts, excess damage, and previous instruction results. Declarative `cost.Additional` and `cost.Alternative` values live in `mtg/game/cost`. Spell casting costs are authored on `CardFace.AdditionalCosts` and `CardFace.AlternativeCosts`; activated and mana ability costs remain on their categorized ability bodies, and the payment planner reads those categorized fields directly. Runtime selections remain in `game.AdditionalCostSelection`, while the legacy `AdditionalCost` string remains only as a compatibility bridge. `CardDef.ImplementationID` is pure data that lets `mtg/rules` route spell resolution to a registered hand-written Card Implementation when the typed primitives are not expressive enough.

Activated abilities are authored using `ActivatedAbility` (regular), `ManaAbility` (mana), or `LoyaltyAbility` (planeswalker) and carry mana costs, typed additional costs, timing restrictions, target specs, X values, a zone of function, and optional `KeywordAbilities` (e.g. `EquipKeyword`). The rules engine uses those fields for tap mana abilities, Equip, Cycling from hand, source-exiling graveyard abilities, and general activated abilities with supported effects. Cycling and graveyard abilities use the same `StackActivatedAbility` shape with the source card preserved after it moves as a cost. `Game.ActivatedAbilitiesThisTurn` tracks once-per-turn activation guards by source object or source card and ability index.

Triggered abilities use `TriggerCondition.Pattern` to match typed `Event` values. The first trigger slice supports exact event-kind matching plus filters for controller, source/self, `ExcludeSelf` for "another" event-source wording, affected player, permanent/card type include/exclude filters, nontoken permanent events, zone transition, damage recipient, and beginning-of-step events with an explicit `Step`. `TriggerCondition.InterveningCondition` is the structured form of an intervening-if predicate and is checked both when the event triggers and when the ability resolves. Dedicated event predicates cover whether an entering permanent was kicked or was cast. `TriggerCondition.State` models simple latched state triggers. Optional "you may" triggered abilities set `TriggeredAbility.Optional`; they still use the stack, and the rules engine asks for the yes/no choice when they resolve. The legacy `TriggerCondition.Event` string is documentation only and is not used for rules behavior.

### Selection

`Selection` (`selection.go`) is pure, valence-agnostic rules data describing **what** characteristics an object must share, never where candidate objects come from. It is the single matcher description that subsumes the characteristic fields previously duplicated across `TargetPredicate`, `PermanentFilter`, the permanent/card filters of `TriggerPattern`, and the `EffectSelector` mass-effect constants. The zero value is a wildcard that matches anything; `Empty()` reports the wildcard state and `Validate()` reports structural contradictions (a card type both required and excluded, every any-of type/color excluded, a keyword both required and excluded) for `ValidateCardDef`.

Selection separates conjunctive and disjunctive list semantics on purpose, because the legacy types disagreed: `RequiredTypes` and `Supertypes` are all-of (an "artifact creature" type line), while `RequiredTypesAny`, `SubtypesAny`, and `ColorsAny` are any-of ("creature or artifact"). `ExcludedTypes`, `ExcludedColors`, and `ExcludedKeyword` reject when any listed value is present. Numeric `ManaValue`, `Power`, and `Toughness` use `compare.Int`. `Controller`/`Player` are relative to a viewing player resolved by the rules adapter, `ExcludeSource` drops the predicate's own source object, `NonToken` rejects tokens, and `TokenOnly` requires them. Counting, total power, and candidate-domain concerns (controlled, defending, equipped, all permanents) stay **outside** Selection and remain with the legacy types until the later reference phase owns runtime binding.

During this phase Selection is additive: `TargetSpec.Selection`, `Condition.ControlsMatching`, and `TriggerPattern.SubjectSelection`/`CardSelection` are new optional fields, and `TargetPredicate.Selection()`/`PermanentFilter.Selection()` adapt the legacy data to a `Selection` (sharing backing slices, so callers must not mutate the result). `ValidateCardDef` rejects specifying both a legacy filter and its Selection equivalent on the same spec. The single matcher that interprets every `Selection` lives in `mtg/rules`.

### Group Reference

`GroupReference` (`group_reference.go`) is pure rules data that pairs a `Selection` (**what** matches) with the **where**: a closed `GroupReferenceDomain` vocabulary (`GroupDomainBattlefield`, `GroupDomainAttachedObject`, `GroupDomainObjectControlled`), an optional anchor `ObjectReference` the domain is defined relative to, and an optional excluded `ObjectReference`. It expresses every candidate-domain concern that deliberately stays outside Selection — battlefield groups, the object an Equipment is attached to, the creatures a defending player controls, and source/target exclusions. Named constructors (`BattlefieldGroup`, `BattlefieldGroupExcluding`, `AttachedObjectGroup`, `ObjectControlledGroup`, `ObjectControlledGroupExcluding`) build every valid shape, accessors (`Domain`, `Selection`, `Anchor`, `Exclusion`) read it back, `Empty()` identifies an omitted zero-value group, and `Valid()`/`Validate()` reject inconsistent non-empty combinations. `EffectSelector.GroupReference()` converts each mass-effect selector constant to its equivalent group, including `EquippedCreature`, `AllCreaturesExceptTarget`, and `OtherCreaturesDefendingPlayerControls`; the runtime enumeration that resolves a `GroupReference` to concrete objects lives in `mtg/rules`.



`ChoiceRequest` and `ChoiceDecision` (`choice.go`) describe engine-mediated decisions that are not priority actions. The current rules engine uses choices for triggered-ability target selection, ordering simultaneous triggers controlled by the same player, optional triggered ability resolution, payment choices, resolution-time value choices, commander-color mana choices, and scry/surveil top-card decisions. Choice data lives in `mtg/game` so agents and rules code can share the request shape without moving behavior out of `mtg/rules`.

### Runtime rules data

`ContinuousEffect`, `EffectDuration`, `RuleEffect`, `ReplacementEffect`, `PreventionShield`, `CostModifier`, `AttackTax`, `ObjectSnapshot`, and linked object/card reference types are runtime data owned by `mtg/rules` behavior but stored in `game.Game` so cloned games, agents, logs, and later simulation tooling can observe a complete rules state. The data package defines shapes only; ordering, expiry, replacement, rule-changing, and payment behavior remain in `mtg/rules`.

### Game Events

`Event` (`event.go`) is the shared typed vocabulary for rules-relevant facts such as spell casts/resolutions, permanents entering or dying, damage dealt or prevented, destruction replacement, cards drawn/discarded/revealed, zone changes, face-up turns, and combat declarations. Permanent-enter events preserve whether the resolved spell was kicked and whether entry came from a cast, non-copy spell. Token events may carry `TokenDef` as last-known definition data because tokens have no `CardInstanceID`.

Events are not player `Action`s and are not report-oriented `GameResult` logs. `mtg/game` defines the event data so card definitions can refer to event kinds and trigger patterns without importing rules behavior; `mtg/rules` emits and consumes events at mutation boundaries. `Game.TriggerEventCursor` records how far trigger detection has consumed the event stream.

Use `Game.AppendEvent` to add events so event-owned slice fields are copied at the boundary. `EventsForTurn`, `EventsThisTurn`, and `EventsPreviousTurn` return copies; callers may inspect or mutate the returned slices without mutating `Game.Events`. `TokenDef` pointers in token events point at shared immutable card definitions and must not be mutated by event consumers.

### Runtime Targets

`Target` (`target.go`) records the concrete target choices made while casting a spell or activating an ability. It is separate from `TargetSpec`: `TargetSpec` describes what an ability can target, while `Target` records what was actually chosen at runtime.

Use `PlayerTarget`, `PermanentTarget`, and `StackObjectTarget` to construct targets so unused ID fields remain zeroed and equality comparisons stay reliable.

### Mana

The mana-related leaf packages split printed card information from produced
mana:

- **color**: card colors and `Identity` for Commander deck legality
- **cost**: printed mana costs and symbols such as `{3}`, `{W/U}`, `{C}`, and `{S}`
- **mana**: produced mana colors, spendable units, and runtime pools with `Add`/`Spend`/`Empty`

### Deterministic shuffling

`Zone.Shuffle(rng)` requires an explicit `*rand.Rand`. Use `NewGameWithRand` or `rules.Engine.NewGame` for reproducible library order in tests and simulations.

### Counters

The `counter` package provides 25 counter kinds (+1/+1, -1/-1, loyalty, charge, time, shield, stun, keyword counters, etc.) and a `Set` type that tracks counts per kind. Shield counters are consumed by the rules replacement/prevention slice to prevent damage or replace destruction. Includes `CancelOpposites()` for the +1/+1 vs -1/-1 state-based action (CR 704.5r).

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Shared battlefield (not per-player) | MTG battlefield is one shared zone; permanents track `Owner` and `Controller` separately |
| Commander damage keyed by `CardInstance` ID | Survives zone changes — a commander re-cast from the command zone is the same card instance |
| `Owner` ≠ `Controller` on all objects | Control-changing effects are fundamental to MTG |
| Token support via `Permanent.Token` + `TokenDef` | Tokens aren't backed by card instances; they need their own `CardDef` |
| Attachments stored on permanents | `Permanent.AttachedTo` and `Permanent.Attachments` let rules maintain Aura/Equipment relationships without making `game` depend on attachment legality |
| Typed game events live in `game` | Card definitions need the event vocabulary, while rules behavior remains in `mtg/rules` |
| Continuous effect data lives in `game` | Card definitions need selectors/effect primitives; `mtg/rules` derives effective values without mutating permanents |
| Hand-written implementation IDs live in `game` | Card definitions can name an escape-hatch implementation without importing behavior; `mtg/rules` owns the registry and mutation helpers |
| Typed `Primitive` effect model | A full effect/resolution system is a rules-engine concern; `game` defines the sealed primitive data, `mtg/rules` owns execution |
| Rules live outside `game` | This package defines state. The `mtg/rules` package enforces legality, resolves abilities, and processes state-based actions |

## What's Not Here (Yet)

This package is the **data model** used by the rules engine. Future layers will add:

- **Card database** — loading real card data (e.g., from Scryfall) into `CardDef` structs
- **AI agent** — decision-making for automated play (see the reference docs in `Agent Instructions & Rules/`)
- **Richer rules support** — remaining keyword actions beyond Flash/basic Equip/Cycling/Kicker, choice-based discard/search/modal decisions, full day/night and meld behavior, and agent-selected replacement/prevention ordering

### Double-faced cards

`CardDef.Layout` and `CardDef.Back` model transform, modal DFC, and double-faced token layouts. `CardDef` root fields are the front-face/default characteristics, and `Back` is present only when the card has a second printed face. Cast actions, stack objects, permanents, events, and LKI snapshots carry `FaceIndex` so modal DFC faces and transformed permanents use the correct face-specific costs, types, abilities, P/T, and replacement abilities while they are on the stack or battlefield.

## CardDef Structural Validation

`ValidateCardDef(card *CardDef) []CardDefIssue` performs deep pure-data structural validation of a card definition and is the authoritative owner of all structural checks that depend only on game data:

- **Nil card** — reports `CardDefIssueNilCard` without dereferencing the pointer.
- **Missing name** — reports `CardDefIssueMissingName` for blank names.
- **Oracle text without abilities** — reports `CardDefIssueOracleWithoutAbilities` when `OracleText` is non-empty but no abilities and no `ImplementationID` are present.
- **TargetSpec validity** — reports `CardDefIssueInvalidTargetSpec` for invalid min/max target counts and unsupported chooser constraints.
- **Target index bounds** — reports `CardDefIssueTargetIndexOutOfRange` when an instruction or condition references a target index that has no matching `TargetSpec`.
- **Reference structure** — reports `CardDefIssueInvalidReference` for `ObjectReference`/`PlayerReference` bindings whose kind and fields are inconsistent (an unknown kind, a linked-object reference with no link ID, an object controller/owner reference with no object, and so on), delegating to each reference's `Validate()` method.
- **Keyword variants** — reports `CardDefIssueInvalidKeywordAbility` for unknown/nil keywords, keywords with missing costs, and `SuspendKeyword` with non-positive `TimeCounters`.
- **Conditions** — validates structured `ObjectReference` bindings in `ActivationCondition`, trigger `InterveningCondition`, and `StaticAbility.Condition`.
- **Continuous effects** — recursively validates `AddAbilities` on `ContinuousEffect` values.
- **Instruction sequences** — delegates to `ValidateInstructionSequence` for nil primitives, invalid primitive references, forward result gates, and duplicate key publications.
- **Nested abilities and replacements** — walks all face ability fields (SpellAbility, ActivatedAbilities, ManaAbilities, LoyaltyAbilities, TriggeredAbilities, ReplacementAbilities, StaticAbilities) plus Back and Alternate faces.

`ValidateCardDef` is a package function (not a method) so that nil `*CardDef` values can be diagnosed without a valid receiver. Issues are returned as `[]CardDefIssue`; each issue records a `FaceName`, `Path`, `Code`, and `Message`.

Runtime policy concerns, such as whether a hand-written `ImplementationID` is registered, remain outside structural Card Definition validation. No Oracle-text parsing or compiler concepts are present in this package.

### Typed data as a compiler target

Because `mtg/game` owns the canonical typed card model and its structural validation, the `cardgen` executable backend uses these same types as a typed **intermediate representation (IR)**. `cardgen` lowers Oracle text directly into `game.StaticAbility`, `game.ActivatedAbility`, `game.ManaAbility`, `game.TriggeredAbility`, `game.ReplacementAbility`, and `game.AbilityContent` values, assembles a `game.CardDef`, and calls `ValidateCardDef` before rendering Go source. This keeps the boundary clean: `mtg/game` owns the typed data and what makes it structurally valid, `mtg/rules` owns behavior, and `cardgen` owns recognition (Oracle text → typed values) and rendering (typed values → Go source). See [`cardgen/README.md`](../../cardgen/README.md#compiler-stages) and [ADR 0008](../../docs/adr/0008-typed-ir-lowering.md).
