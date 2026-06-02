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
├── ability.go                 # AbilityDef, Keywords, TriggerConditions, Effects, Modes
├── condition.go               # Reusable Condition and PermanentFilter predicates
├── reference.go               # ObjectReference, PlayerReference, and CardReference effect bindings
├── subtype.go                 # Central artifact/creature/enchantment/land subtype constants
├── choice.go                  # ChoiceRequest/ChoiceDecision for non-action decisions
├── permanent.go               # Permanent — battlefield state for a card or token
├── player.go                  # Player — life, zones, commander tracking, designations
├── zone.go                    # Zone container & ZoneType enum (library, hand, graveyard…)
├── stack.go                   # Stack (LIFO) & StackObject (spells/abilities resolving)
├── target.go                  # Target — runtime target choices for spells/abilities
├── event.go                   # GameEvent — typed rules facts emitted by mtg/rules
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
| `CardDef` | `card.go` | Immutable card template — name, mana cost, types, supertypes/subtypes, abilities, P/T, loyalty, battle defense, ETB tapped/counter/payment data, optional hand-written implementation ID, and optional `CardFace` data for double-faced layouts. Shared across games. |
| `CardInstance` | `card.go` | A specific card in a specific game. Has a unique `id.ID` and an `Owner`. |
| `Permanent` | `permanent.go` | A card or token on the battlefield — tapped, counters, damage, attachments, phased out, face-down origin state, current printed face, etc. |
| `StackObject` | `stack.go` | A spell or ability on the stack — selected/source face, source zone/card, face-down origin state, source ability index, targets, chosen modes, X value, additional costs, and linked resolution results. |
| `Target` | `target.go` | A runtime targeting choice: player, permanent, or stack object. |
| `ChoiceRequest` | `choice.go` | A bounded non-action decision such as trigger target choice, trigger ordering, or optional-effect yes/no. |
| `Condition` | `condition.go` | Reusable data-only predicates for static ability conditions, activation restrictions, intervening-if checks, effect conditions, and replacement effects. |
| `ObjectReference`, `PlayerReference`, `CardReference` | `reference.go` | Reusable resolution-time bindings for effects that need source objects, target-derived controllers/owners, linked objects/cards, damage sources, non-default recipients, or card-condition checks. |

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

`AbilityDef` (`ability.go`) captures the four kinds of MTG abilities:

| Kind | Marker | Example |
|------|--------|---------|
| Spell | (text on instants/sorceries) | "Destroy target creature." |
| Activated | `[Cost]: [Effect]` | "{T}: Add {G}." |
| Triggered | `When` / `Whenever` / `At` | "When this enters the battlefield…" |
| Static | declarative | "Creatures you control get +1/+1." |

50+ keywords are enumerated (flying, haste, deathtouch, lifelink, indestructible, protection, flashback, cascade, discover, eternalize, morph, disguise, etc.). Plain non-parameterized keywords have reusable `AbilityDef` templates such as `FlyingAbility`, `HasteAbility`, `ReachAbility`, and `ExaltedAbility`; treat these exported values as immutable and copy them into `CardDef.Abilities` rather than mutating their fields. Keywords that need card-specific data still use explicit `AbilityDef` values or constructors: `AbilityDef.ProtectionFromColors` parameterizes Protection for color-based damage prevention and targeting restrictions; `MorphCost` and `DisguiseCost` parameterize face-down turn-up costs; `EternalizeAbility` builds the full named keyword activation from a cost plus the source card's creature subtypes. `TargetSpec` supports min/max target counts, legacy natural-language `Constraint` text, broad target categories through `TargetAllow`, announcement-time `Chooser` values, and structured `TargetPredicate` data for common color, type, controller, tapped, combat-state, keyword, mana-value, P/T, and "another" filters. For opponent-chosen target slots, predicates such as `ControllerYou` are evaluated relative to the choosing opponent, so `Chooser: TargetChooserOpponent` plus `Controller: ControllerYou` means "a target controlled by that opponent." `Condition` and `PermanentFilter` model reusable controller-controls and referenced-object predicates for static ability conditions, activation restrictions, trigger intervening-if checks, effect conditions, and replacement effects; event-permanent references may use last-known information. Conditions also cover aggregate controlled-creature total power, class level gates, monstrous-state checks, max-speed checks, and event-permanent name uniqueness for cards such as Guardian Project. Static abilities may carry `Condition` plus `EffectApplyContinuous` templates, `EffectModifyPT` effects, or `EffectApplyRule` rule effects; `mtg/rules` derives those continuous and rule-changing effects dynamically from battlefield and supported non-battlefield zones. `AbilityDef.KickerCost` and `KickerEffects` model the initial Kicker slice. Modal abilities use `Modes` plus `MinModes`, `MaxModes`, and `AllowDuplicateModes` for choose-one, choose-N, up-to-N, one-or-both, all-mode, and duplicate-mode templates. `Effect` stores simple effect primitive data such as `Type`, `Amount`, `DynamicAmount`, `TargetIndex`, object/card references, `Optional`, `ResultCondition`, `Condition`, `CardCondition`, `Selector`, `PowerDelta`, `ToughnessDelta`, `CounterKind`, `CounterSource`, `ManaColor`, resolution `Choice` / `ChoiceLinkID`, resolution `Payment`, runtime `Replacement`, runtime/static `RuleEffects`, library search specs, `Duration`, `UntilEndOfTurn`, `Token`, token-copy specs, `ContinuousEffects`, `DelayedTrigger`, `EmblemAbilities`, `LinkID`, and `Description`; the rules engine owns execution behavior. `EffectApplyContinuous` bridges declarative card text to runtime `ContinuousEffect` layer entries for type, subtype, color, ability, control, and P/T-setting changes. `EffectApplyRule` bridges card text to rule-changing effects such as life-gain prohibitions, attack/block/unblockable prohibitions, generic spell cost modifiers, graveyard cast permissions, and can't-be-countered spell protections. `EffectMoveCounters` moves all counters from a target or triggering event permanent, using last-known information when that permanent has left the battlefield. `EffectChoose` and `EffectPay` model value-producing resolution choices and optional resolution-time payments. `EffectReplace` creates runtime replacement effects for zone-destination changes and simple ETB tapped/counter modifiers, while `CardDef.EntersTappedUnlessPaid` models ETB payment choices such as paying life to enter untapped. `EffectCounter`, `EffectDiscard`, `EffectSearch`, `EffectReveal`, `EffectDiscover`, `EffectShufflePermanentIntoLibrary`, `EffectPutOnBattlefield`, `EffectInvestigate`, `EffectProliferate`, `EffectGoad`, `EffectStartEngines`, `EffectSetClassLevel`, and `EffectMonstrosity` cover reusable keyword-action primitives, including linked reveal/discover-then-act flows. `SearchSpec` supports library-to-hand and library-to-battlefield searches with card-type, supertype, subtype-any, reveal, shuffle, and enters-tapped options. `DynamicAmount` represents effect amounts determined on resolution, such as X, target characteristics, object power, selector counts, counter counts, excess damage, and linked "that much" values. Linked `EffectResolutionResult` data supports success-aware "if you do" / "if you don't" branches during a single stack-object resolution. `AbilityDef.AdditionalCosts`, `AlternativeCosts`, and runtime `CostModifier` data are used by the rules payment planner, while the legacy `AdditionalCost` string remains only as a compatibility bridge. `CardDef.ImplementationID` is pure data that lets `mtg/rules` route spell resolution to a registered hand-written card implementation when effect primitives are not expressive enough. Current combat rules also consult supported keyword counters such as Flying, First Strike, Deathtouch, Lifelink, Trample, Vigilance, and Indestructible.

Activated abilities can carry mana costs, typed additional costs, timing restrictions, target specs, X values, a zone of function, and an `IsManaAbility` marker. The current rules engine uses those fields for simple tap mana abilities, basic Equip abilities, Cycling from hand, source-exiling graveyard abilities, and general non-mana activated abilities with supported effects. Cycling and graveyard abilities use the same `StackActivatedAbility` shape with the source card preserved after it moves as a cost. `Game.ActivatedAbilitiesThisTurn` tracks once-per-turn activation guards by source object or source card and ability index.

Triggered abilities use `TriggerCondition.Pattern` to match typed `GameEvent` values. The first trigger slice supports exact event-kind matching plus filters for controller, source/self, `ExcludeSelf` for "another" event-source wording, affected player, permanent/card type include/exclude filters, nontoken permanent events, zone transition, damage recipient, and beginning-of-step events with an explicit `Step`. `TriggerCondition.InterveningCondition` is the structured form of an intervening-if predicate and is checked both when the event triggers and when the ability resolves. `TriggerCondition.State` models simple latched state triggers. Optional "you may" triggered abilities set `AbilityDef.Optional`; they still use the stack, and the rules engine asks for the yes/no choice when they resolve. The legacy `TriggerCondition.Event` string is documentation only and is not used for rules behavior.

### Choices

`ChoiceRequest` and `ChoiceDecision` (`choice.go`) describe engine-mediated decisions that are not priority actions. The current rules engine uses choices for triggered-ability target selection, ordering simultaneous triggers controlled by the same player, optional triggered ability resolution, payment choices, resolution-time value choices, commander-color mana choices, and scry/surveil top-card decisions. Choice data lives in `mtg/game` so agents and rules code can share the request shape without moving behavior out of `mtg/rules`.

### Runtime rules data

`ContinuousEffect`, `EffectDuration`, `RuleEffect`, `ReplacementEffect`, `PreventionShield`, `CostModifier`, `AttackTax`, `ObjectSnapshot`, and linked object/card reference types are runtime data owned by `mtg/rules` behavior but stored in `game.Game` so cloned games, agents, logs, and later simulation tooling can observe a complete rules state. The data package defines shapes only; ordering, expiry, replacement, rule-changing, and payment behavior remain in `mtg/rules`.

### Game Events

`GameEvent` (`event.go`) is the shared typed vocabulary for rules-relevant facts such as spell casts/resolutions, permanents entering or dying, damage dealt or prevented, destruction replacement, cards drawn/discarded/revealed, zone changes, face-up turns, and combat declarations. Token events may carry `TokenDef` as last-known definition data because tokens have no `CardInstanceID`.

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
| Placeholder `Effect` type | A full effect/resolution system is a rules-engine concern, not a scaffold concern |
| Rules live outside `game` | This package defines state. The `mtg/rules` package enforces legality, resolves abilities, and processes state-based actions |

## What's Not Here (Yet)

This package is the **data model** used by the rules engine. Future layers will add:

- **Card database** — loading real card data (e.g., from Scryfall) into `CardDef` structs
- **AI agent** — decision-making for automated play (see the reference docs in `Agent Instructions & Rules/`)
- **Richer rules support** — remaining keyword actions beyond Flash/basic Equip/Cycling/Kicker, choice-based discard/search/modal decisions, full day/night and meld behavior, and agent-selected replacement/prevention ordering

### Double-faced cards

`CardDef.Layout` and `CardDef.Back` model transform, modal DFC, and double-faced token layouts. `CardDef` root fields are the front-face/default characteristics, and `Back` is present only when the card has a second printed face. Cast actions, stack objects, permanents, events, and LKI snapshots carry `FaceIndex` so modal DFC faces and transformed permanents use the correct face-specific costs, types, abilities, P/T, and ETB data while they are on the stack or battlefield.
