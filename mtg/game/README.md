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
├── choice.go                  # ChoiceRequest/ChoiceDecision for non-action decisions
├── permanent.go               # Permanent — battlefield state for a card or token
├── player.go                  # Player — life, zones, commander tracking, designations
├── zone.go                    # Zone container & ZoneType enum (library, hand, graveyard…)
├── stack.go                   # Stack (LIFO) & StackObject (spells/abilities resolving)
├── target.go                  # Target — runtime target choices for spells/abilities
├── event.go                   # GameEvent — typed rules facts emitted by mtg/rules
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
├── mana/                      # Leaf package: mana system
│   ├── README.md              #   Package guide
│   ├── doc.go                 #   Package documentation
│   ├── color.go               #   Color enum (W, U, B, R, G, C)
│   ├── symbol.go              #   Symbol — colored, colorless, generic, hybrid, phyrexian, snow, X
│   ├── cost.go                #   Cost — ordered list of symbols, ManaValue(), Colors()
│   └── pool.go                #   Pool — runtime mana tracking; ColorIdentity for Commander
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
| `CardDef` | `card.go` | Immutable card template — name, mana cost, types, supertypes/subtypes, abilities, P/T, loyalty, battle defense, and optional hand-written implementation ID. Shared across games. |
| `CardInstance` | `card.go` | A specific card in a specific game. Has a unique `id.ID` and an `Owner`. |
| `Permanent` | `permanent.go` | A card or token on the battlefield — tapped, counters, damage, attachments, phased out, face-down, etc. |
| `StackObject` | `stack.go` | A spell or ability on the stack — source ability index, targets, chosen modes, X value, additional costs. |
| `Target` | `target.go` | A runtime targeting choice: player, permanent, or stack object. |
| `ChoiceRequest` | `choice.go` | A bounded non-action decision such as trigger target choice, trigger ordering, or optional-effect yes/no. |

### Player

Each `Player` (`player.go`) tracks:

- **Life** (starts at 40), **poison counters**, **commander damage** received (per commander)
- **Commander tax** (cast count from command zone × 2)
- **Five zones**: Library, Hand, Graveyard, Exile, Command Zone
- **Mana pool** (`mana.Pool`)
- **Designations**: monarch, initiative, city's blessing, ring level, energy, experience

### Game

`Game` (`game.go`) ties everything together:

- `[4]*Player` — the four players
- `[]*Permanent` — shared battlefield
- `Stack` — LIFO spell/ability stack
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

50+ keywords are enumerated (flying, haste, deathtouch, lifelink, indestructible, protection, flashback, cascade, etc.). `AbilityDef.ProtectionFromColors` parameterizes the initial Protection slice for color-based damage prevention and targeting restrictions. Static abilities may carry `EffectModifyPT` effects with selectors such as `EffectSelectorCreaturesYouControl` or `EffectSelectorOtherCreaturesYouControl`; `mtg/rules` derives those continuous P/T effects dynamically from the battlefield. `Effect` stores simple effect primitive data such as `Type`, `Amount`, `TargetIndex`, `Selector`, `PowerDelta`, `ToughnessDelta`, `ManaColor`, `UntilEndOfTurn`, `Token`, and `Description`; the rules engine owns execution behavior. `CardDef.ImplementationID` is pure data that lets `mtg/rules` route spell resolution to a registered hand-written card implementation when effect primitives are not expressive enough. Current combat rules also consult supported keyword counters such as Flying, First Strike, Deathtouch, Lifelink, Trample, Vigilance, and Indestructible.

Activated abilities can carry mana costs, an additional-cost string, timing restrictions, target specs, X values, and an `IsManaAbility` marker. The current rules engine uses those fields for simple tap mana abilities, basic Equip abilities, Cycling from hand, and general non-mana activated abilities with supported effects. Cycling abilities use the same `StackActivatedAbility` shape with the source card preserved after it is discarded. `Game.ActivatedAbilitiesThisTurn` tracks once-per-turn activation guards by source object and ability index.

Triggered abilities use `TriggerCondition.Pattern` to match typed `GameEvent` values. The first trigger slice supports exact event-kind matching plus filters for controller, source/self, affected player, permanent type, zone transition, and damage recipient. Optional "you may" triggered abilities set `AbilityDef.Optional`; they still use the stack, and the rules engine asks for the yes/no choice when they resolve. The legacy `TriggerCondition.Event` string is documentation only and is not used for rules behavior.

### Choices

`ChoiceRequest` and `ChoiceDecision` (`choice.go`) describe engine-mediated decisions that are not priority actions. The current rules engine uses choices for triggered-ability target selection, ordering simultaneous triggers controlled by the same player, and optional triggered ability resolution. Choice data lives in `mtg/game` so agents and rules code can share the request shape without moving behavior out of `mtg/rules`.

### Game Events

`GameEvent` (`event.go`) is the shared typed vocabulary for rules-relevant facts such as spell casts/resolutions, permanents entering or dying, damage dealt or prevented, destruction replacement, cards drawn/discarded, zone changes, and combat declarations. Token events may carry `TokenDef` as last-known definition data because tokens have no `CardInstanceID`.

Events are not player `Action`s and are not report-oriented `GameResult` logs. `mtg/game` defines the event data so card definitions can refer to event kinds and trigger patterns without importing rules behavior; `mtg/rules` emits and consumes events at mutation boundaries. `Game.TriggerEventCursor` records how far trigger detection has consumed the event stream.

### Runtime Targets

`Target` (`target.go`) records the concrete target choices made while casting a spell or activating an ability. It is separate from `TargetSpec`: `TargetSpec` describes what an ability can target, while `Target` records what was actually chosen at runtime.

Use `PlayerTarget`, `PermanentTarget`, and `StackObjectTarget` to construct targets so unused ID fields remain zeroed and equality comparisons stay reliable.

### Mana

The `mana` package models the full MTG mana system:

- **Colors**: W, U, B, R, G, C (colorless)
- **Symbols**: colored, colorless (`{C}`), generic (`{3}`), variable (`{X}`), hybrid (`{W/U}`), mono-hybrid (`{2/W}`), phyrexian (`{W/P}`), snow (`{S}`)
- **Cost**: ordered symbol list with `ManaValue()` and `Colors()`
- **Pool**: runtime mana tracking with `Add`/`Spend`/`Empty`
- **ColorIdentity**: set of colors for Commander deck legality

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
- **Richer rules support** — advanced payment choices, keyword actions beyond Flash/basic Equip/Cycling, choice-based decisions, the full layer system for continuous effects, and richer replacement/prevention ordering
