# Council4

A playtesting engine for Magic: The Gathering Commander decks. It runs automated games between AI-controlled decks and reports detailed analytics on deck performance.

## Language

### Game Concepts

**Game**:
A single simulated 4-player Commander match from setup through one player winning.
_Avoid_: Match, round, session

**Commander**:
The designated legendary creature leading a player's deck, cast from the command zone with escalating tax.
_Avoid_: General

**Decklist**:
A player's 100-card deck specification in standard text format — card names with quantities, one per line.
_Avoid_: Deck list, deck file, deck config

**Card Definition**:
The immutable printed data of a Magic card — name, mana cost, types, oracle text. Sourced from Scryfall.
_Avoid_: Card data, card template

**Card Implementation**:
The declarative behavior description that tells the engine what a card does at runtime — a composition of effect primitives, with hand-written code for cards too complex to express declaratively.
_Avoid_: Card logic, card script, card handler

**Effect Primitive**:
A composable building-block game action — deal damage, destroy permanent, draw cards, create token, etc. Card implementations are composed from these.
_Avoid_: Effect type, action type

**Instruction**:
One ordered resolution step in a spell or ability. It combines exactly one typed Effect Primitive with shared sequencing data such as conditions, optionality, result publication, and result gates.
_Avoid_: Effect struct, opcode

**Static Declaration**:
Continuous-effect or rule-effect data attached directly to a static ability. Static declarations are derived while their source applies; they are not Instructions because they never resolve.
_Avoid_: Static instruction, permanent effect sequence

**Combat**:
The turn phase where creatures attack players or other attackable objects and deal combat damage.
_Avoid_: Battle, fight, attack phase

**Combat Module**:
The in-place rules module (`mtg/rules` `combatEngine`) that owns combat phase orchestration, attack/block declaration, combat damage assignment, and attack-tax payment integration.
_Avoid_: Combat package until a small adapter seam exists

**Attacker**:
A creature declared to attack during combat.
_Avoid_: Combatant, assailant

**Combat Damage**:
Damage dealt by attacking and blocking creatures during the combat damage step.
_Avoid_: Attack damage, battle damage

### Engine Concepts

**Engine**:
The rules engine (`rules.Engine`) that owns the full game loop — phase transitions, priority, state-based actions, effect execution. Receives a seeded RNG for reproducibility.
_Avoid_: Game loop, controller, runtime

**Action**:
A tagged struct describing a single player decision — pass priority, play a land, cast a spell, declare attackers, etc. Produced by the engine via `LegalActions`, chosen by agents, applied by the engine. Lives in `game/action/`.
_Avoid_: Move, choice, command, input

**Payment Choice**:
An engine-mediated decision about how to pay a legal cost, such as which hybrid side to use, whether to pay life for phyrexian mana, or which card/permanent satisfies an additional cost. Payment choices should use the same bounded choice pathway as other non-action decisions.
_Avoid_: Hidden payment heuristic, implicit cost branch

**Payment Plan**:
The concrete rules plan for paying a spell, ability, or attack cost: mana to spend, permanents to tap for mana, life payments, and selected additional costs. Payment plans are built and validated in `mtg/rules` before mutation.
_Avoid_: Cost guess, mana receipt

**Payment Planner**:
The rules module (`mtg/rules/payment`) that builds, validates, and applies **Payment Plans** through a small state adapter supplied by `mtg/rules`.
_Avoid_: Mana helper, cost utility

**Mana Unit**:
A spendable unit of mana with color and minimal provenance such as whether it was produced by a snow source. Mana units let the engine answer restrictions like `{S}` without treating all mana of one color as interchangeable.
_Avoid_: Raw mana count when provenance matters

**Additional Cost**:
A cost paid in addition to a spell or ability's mana cost, such as sacrificing a creature, discarding a card, paying life, or revealing a card. Additional costs should be typed data rather than freeform text when they affect rules behavior.
_Avoid_: Cost text parser, deterministic side effect

**Alternative Cost**:
A cost that replaces a spell or ability's normal mana cost when selected, while still allowing required additional costs. Alternative costs need an explicit cost-selection stage before payment planning.
_Avoid_: Cost reduction, extra cost

**Cost Modifier**:
A runtime cost increase, reduction, set, minimum, or tax applied after normal/alternative cost selection and before payment planning.
_Avoid_: Alternative cost, payment result

**Choice**:
An engine-mediated decision that is not a priority **Action**, such as choosing targets for triggered abilities, ordering simultaneous triggers, or deciding whether to apply an optional effect. In code, `game.ChoiceRequest` is answered by a `rules.ChoiceAgent` when available, with deterministic fallback.
_Avoid_: Action, UI prompt, ad hoc callback

**Stack Object**:
A spell or ability waiting on the stack to resolve. In code, `game.StackObject` references its source card or permanent, controller, chosen runtime targets, modes, and X value.
_Avoid_: Stack item, pending spell

**Zone**:
A game area identified by `zone.Type`, such as the library, hand, battlefield, graveyard, stack, exile, or command zone. Player-owned card collections use `zone.Zone`; the shared battlefield and stack use richer runtime representations in `game`.
_Avoid_: Card location enum, pile

**Activated Ability**:
An ability with a cost and effect that an **Agent** may choose as an **Action** when legal. Non-mana activated abilities use the **Stack Object** path; mana abilities resolve immediately.
_Avoid_: Manual ability hook, special action

**Runtime Target**:
The concrete target chosen while casting a spell or activating an ability. In code, `game.Target` is separate from `game.TargetSpec`, which only describes what may be targeted.
_Avoid_: Target spec, raw target ID

**Selection**:
Pure, valence-agnostic data describing WHICH game objects share a characteristic predicate — required/excluded types, supertypes, any-of subtypes/colors, controller/player relation, tapped/combat state, keywords, mana value, and power/toughness. It describes WHAT matches, never where candidates come from; counting and candidate-domain concerns stay outside it. In code, `game.Selection` is interpreted by a single matcher in `mtg/rules` that subsumes the legacy `TargetPredicate`, `PermanentFilter`, `TriggerPattern` filters, and `EffectSelector` characteristic logic.
_Avoid_: Predicate, filter, selector, matcher (for the data itself)

**Group Reference**:
Pure data describing WHERE a mass effect finds a group of permanents — a candidate domain (battlefield, the object an Equipment is attached to, the permanents an object's controller controls), a **Selection** that narrows it, and optional anchor/exclusion object references. Group Reference owns the candidate-domain and exclusion concerns that **Selection** deliberately leaves out. In code, `game.GroupReference` is pure data with a closed domain vocabulary; `EffectSelector.GroupReference()` converts each mass-effect selector to its equivalent, and the reference resolver in `mtg/rules` enumerates a group's concrete objects.
_Avoid_: Object reference (for a group), selector, group filter

**Game Event**:
A rules-relevant fact that occurred during a **Game**, such as a spell being cast, a permanent entering the battlefield, damage being dealt, or a creature dying.
In code, `game.Event` values are appended to `game.Game.Events` by `rules.Engine` helpers at mutation boundaries.
_Avoid_: Log entry, action history, report record

**Replacement/Prevention Effect**:
A rules behavior that changes or prevents a pending mutation before it happens, such as preventing damage or replacing destruction with shield-counter removal. In code, the current slice lives in `mtg/rules` before damage and destroy helpers mutate state, while the resulting facts are emitted as `game.Event`s.
_Avoid_: Post-mutation cleanup, log-only prevention

**Continuous Effect**:
A persistent rules effect derived from current game state rather than a one-time mutation, such as an anthem that gives other creatures you control +1/+1. In code, the current slice is recalculated through `rules` effective-value helpers instead of being stored on permanents.
_Avoid_: Permanent mutation, temporary modifier

**Last-Known Information**:
A snapshot of an object's effective characteristics immediately before it changes zones, used by dies/leaves-the-battlefield triggers and linked effects.
_Avoid_: Current card definition, stale battlefield pointer

**Trigger Pattern**:
A structured matcher on a **Game Event** used by a triggered ability. In code, `game.TriggerPattern` hangs off `game.TriggerCondition` and filters by event kind, controller/player relationship, source/self, zones, permanent type, and damage recipient.
_Avoid_: Trigger text parser, string trigger

**Game Result**:
The structured output of a completed game — winner, elimination order, turn count, and per-turn logs of actions taken, mana spent, cards drawn. Produced by `Engine.RunGame()`, consumed by the report package.
_Avoid_: Game log, match result, outcome

**Player Observation**:
A purpose-built fog-of-war view of the game state for a specific player — own hand, public zones, opponent hand sizes, but never hidden cards. Passed to agents when they choose actions.
_Avoid_: Game view, player state, info set, visible state

**Effect Resolver**:
The rules-engine code that executes typed `game.Primitive` values when a spell or ability resolves. Primitive data lives in `game`; behavior lives in `rules` (dispatched via the primitive handler registry).
_Avoid_: Card script, effect data

### AI Concepts

**Agent**:
A decision-making entity that occupies a player seat, receives observations, and chooses actions. All players (including the user's deck) are controlled by agents during simulation.
_Avoid_: Bot, player AI, opponent

**Strategy**:
A scoring function that evaluates legal actions to guide an agent's decisions. Encapsulated behind an interface so different strategies (generic, aggro, control) can be swapped.
_Avoid_: Scorer, heuristic, policy

**Observation**:
The fog-of-war filtered game state visible to a specific player — own hand, public zones, opponent hand sizes, but never hidden cards.
_Avoid_: Player view, visible state, info set

### Simulation Concepts

**Simulation**:
A batch run of many games (default 1,000) with the same decks and agent configurations, producing aggregate statistics.
_Avoid_: Tournament, test run, batch

**Deck Report**:
The analytical output of a simulation — per-card performance, mana curve analysis, tempo metrics, win rate, and game-length distribution.
_Avoid_: Results, stats, output

### Infrastructure Concepts

**Card Registry**:
A lookup table (`cards.Registry`) mapping card names to `CardDef` values. Pure data, no behavior. Used by the decklist parser to resolve card names.
_Avoid_: Card database, card store, card catalog

**Card Generation (cardgen)**:
Isolated tooling that turns Scryfall bulk data and Oracle text into executable `CardDef` Go source. It is a three-stage compiler: **recognition** (Oracle text → conservative semantic IR in `cardgen/oracle`), **lowering** (semantic IR → typed `game.*` ability values and an assembled `game.CardDef` validated by `game.ValidateCardDef`), and **rendering** (typed `CardDef` → deterministic Go source). It emits no partial definitions or TODO scaffolds. `mtg/game` owns typed data and validity; `mtg/rules` owns behavior; `cardgen` owns recognition, lowering, and rendering. Cards it cannot fully recognize are reported as unsupported; exceptional cards may independently use the hand-written **Card Implementation** escape hatch via `ImplementationID`. See [ADR 0008](docs/adr/0008-typed-ir-lowering.md).
_Avoid_: Card compiler (the package is tooling, not part of the runtime engine)

## Relationships

- A **Game** has exactly 4 **Agents**, each piloting a **Decklist**
- A **Decklist** references **Card Definitions**, each backed by a **Card Implementation**
- A **Card Implementation** composes one or more **Effect Primitives**
- An **Agent** uses a **Strategy** to evaluate **Actions** presented in its **Player Observation**
- An **Activated Ability** is represented as an **Action** and usually becomes a **Stack Object** before its **Effect Primitives** resolve
- The **Engine** emits **Game Events** while applying rules; **Trigger Patterns** consume those events to put triggered abilities on the stack
- **Continuous Effects** change effective characteristics while their source remains applicable, without mutating printed **Card Definitions** or battlefield **Permanents**
- A **Simulation** runs many **Games** via the **Engine** and produces a **Deck Report**
- The **Engine** produces **Legal Actions**, agents choose one, the engine applies it
- The **Engine** produces a **Game Result** per game, consumed by the **Deck Report**
- The **Card Registry** maps card names to **Card Definitions**

## Example dialogue

> **Dev:** "When an **Agent** decides what to play, does it see the full game state?"
> **Domain expert:** "No — it only sees its **Observation**, which hides opponents' hands and library order. The engine enforces this boundary."

> **Dev:** "What if a card does something weird that the declarative system can't express?"
> **Domain expert:** "The **Card Implementation** falls back to hand-written Go code. Most cards are just **Effect Primitives** composed together, but the escape hatch exists for the ~5% that need custom logic."

## Reference Documents

- [Card Game AI Research](docs/research/card-game-ai-research.md) — comprehensive guide to AI opponents for multiplayer card games: architecture, MCTS, rule-based agents, LLMs, imperfect information, Go implementation guidance
- [MTG Glossary](docs/research/MTG-GLOSSARY.md) — canonical definitions of Magic: The Gathering terms
- [Commander Strategy](docs/research/COMMANDER-STRATEGY.md) — strategy concepts and patterns specific to Commander
- [Commander Agent Playbook](docs/research/COMMANDER-AGENT-PLAYBOOK.md) — how AI agents should approach Commander decision-making
- [Card Text Parsing](docs/research/CARD-TEXT-PARSING.md) — how Card Generation recognizes MTG Oracle text as structured data
- [MTG General Research](docs/research/MTG-General-Research.md) — broader research on MTG game mechanics
- Magic Comprehensive Rules — official rules referenced externally; local full-text copies are not committed.

## Flagged ambiguities

- "Deck" is used to mean both the **Decklist** (the specification) and the in-game library (the zone). In code, the specification is a **Decklist** and the in-game zone is a **Library**.
