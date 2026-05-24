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

**Combat**:
The turn phase where creatures attack players or other attackable objects and deal combat damage.
_Avoid_: Battle, fight, attack phase

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

**Stack Object**:
A spell or ability waiting on the stack to resolve. In code, `game.StackObject` references its source card or permanent, controller, chosen runtime targets, modes, and X value.
_Avoid_: Stack item, pending spell

**Runtime Target**:
The concrete target chosen while casting a spell or activating an ability. In code, `game.Target` is separate from `game.TargetSpec`, which only describes what may be targeted.
_Avoid_: Target spec, raw target ID

**Game Result**:
The structured output of a completed game — winner, elimination order, turn count, and per-turn logs of actions taken, mana spent, cards drawn. Produced by `Engine.RunGame()`, consumed by the report package.
_Avoid_: Game log, match result, outcome

**Player Observation**:
A purpose-built fog-of-war view of the game state for a specific player — own hand, public zones, opponent hand sizes, but never hidden cards. Passed to agents when they choose actions.
_Avoid_: Game view, player state, info set, visible state

**Effect Resolver**:
The rules-engine code that executes `game.Effect` primitives when a spell or ability resolves. Effect data lives in `game`; behavior lives in `rules`.
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

## Relationships

- A **Game** has exactly 4 **Agents**, each piloting a **Decklist**
- A **Decklist** references **Card Definitions**, each backed by a **Card Implementation**
- A **Card Implementation** composes one or more **Effect Primitives**
- An **Agent** uses a **Strategy** to evaluate **Actions** presented in its **Player Observation**
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
- [Card Text Parsing](docs/research/CARD-TEXT-PARSING.md) — how to parse MTG oracle text into structured data, used when generating card implementations
- [MTG General Research](docs/research/MTG-General-Research.md) — broader research on MTG game mechanics
- Magic Comprehensive Rules — official rules referenced externally; local full-text copies are not committed.

## Flagged ambiguities

- "Deck" is used to mean both the **Decklist** (the specification) and the in-game library (the zone). In code, the specification is a **Decklist** and the in-game zone is a **Library**.
