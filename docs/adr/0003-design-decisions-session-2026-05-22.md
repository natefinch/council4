# Design Decisions — Grilling Session (2026-05-22)

## Project Goal

Build a playtesting engine for Magic: The Gathering Commander decks. The user provides 4 decklists (their own + 3 opponents), the engine runs 1,000+ automated games with AI-controlled players, and produces a detailed deck performance report. No UI — CLI only.

## Decisions

### Card Implementation Architecture
**Decision:** Hybrid declarative + code escape hatch.
- Most cards are declarative compositions of effect primitives (damage, destroy, draw, etc.)
- Cards too complex for declarative expression get hand-written Go code behind the same interface
- Start fully declarative, add code escape hatches as needed
- **ADR:** `docs/adr/0001-hybrid-declarative-card-implementations.md`

### Card Data Pipeline
**Superseded by ADR 0008:** Scryfall bulk JSON → deterministic Oracle compiler at generation time.
- Download Scryfall bulk data as the source of truth for card metadata (name, mana cost, types, oracle text, etc.)
- Parse and compile Oracle text into validated declarative Card Definitions
- Generation happens offline (not at runtime), and selected output is committed to the repo
- The `docs/research/CARD-TEXT-PARSING.md` document guides the oracle text → implementation translation
- Scryfall data doesn't update often; regenerate periodically as needed

### AI Approach
**Decision:** Rule-based heuristic agents first. MCTS and LLM agents later.
- Rule-based agents are fast (milliseconds per decision), enabling 1,000 games in seconds
- The Agent interface (`ChooseAction(obs PlayerObservation) Action`) is the same for all AI types
- MCTS can be added later as a stronger opponent tier (same interface, just slower)
- LLM agents are explicitly planned as a fun future option — Go's goroutine model handles the blocking API calls naturally
- LLMs could also be used to bootstrap heuristic weights for rule-based agents

### Agent Strategy System
**Decision:** Generic "good stuff" scorer behind a Strategy interface.
- Start with one generic strategy that values board presence, card advantage, mana efficiency, and threat removal
- The Strategy interface is encapsulated so archetype-specific strategies (aggro, control, midrange, combo) can be added later without changing the Agent
- Auto-detection of deck archetype from decklist is a future enhancement

### Priority System
**Decision:** Smart priority — skip when no player can respond.
- Only stop for priority when at least one player has a legal instant-speed action
- Functionally equivalent to full priority in outcome
- Critical for simulation speed in 4-player games (>90% of priority passes are trivial)
- **ADR:** `docs/adr/0002-smart-priority-skips-empty-passes.md`

### OpenSpiel
**Decision:** Do not reimplement OpenSpiel. Build directly on the existing game engine.
- OpenSpiel is a general-purpose framework for many games and many algorithms — massive overkill for one game
- The valuable parts (GameState/Action/Observation interface pattern) are already reflected in the existing code
- The `euchre-bot` Go repo is a better reference model — single 4-player card game, no framework overhead

### V1 Mechanic Scope
**Decision:** Core MTG mechanics plus X spells and sacrifice-as-cost.

**Must have (v1):**
- Play lands, tap for mana (including multi-color)
- Cast creatures, instants, sorceries
- Combat: declare attackers (choosing who to attack), declare blockers, deal damage
- Keywords: flying, trample, first strike, double strike, deathtouch, lifelink, haste, vigilance, defender, reach, menace, hexproof, indestructible
- Targeted removal (destroy/exile target creature/permanent)
- Life gain/loss, draw cards, discard cards
- Commander from command zone (with tax), commander damage tracking
- Player elimination
- ETB triggers, death triggers
- Token creation
- +1/+1 and -1/-1 counters
- Equipment and auras (equip/attach)
- Board wipes (destroy all creatures)
- Counterspells
- Flash
- Mana ramp (search library for land)
- X spells
- Sacrifice as cost

**Deferred (v2, but architecture must not block these):**
- Planeswalkers (loyalty abilities)
- Complex triggered ability chains
- Replacement effects
- Continuous effects / layer system
- Modal spells ("Choose one")
- Graveyard recursion

### Deck Input Format
**Decision:** Standard Moxfield/MTGO text export format.
- One card per line: `1 Card Name`
- Commander designated by section header (`// Commander` or `COMMANDER:`)
- User provides all 4 decklists explicitly

### Output and Metrics
**Decision:** Text summary to stdout + detailed JSON report file.

**Metrics to track:**
- Win rate and average finishing position
- Turns to win or lose (game length distribution)
- Per-card cast frequency and performance (cards in winning vs losing games)
- Mana curve analysis: lands per turn, mana spent vs available, missed land drops
- Land flood detection (excess lands, nothing to cast)
- Expensive cards rotting in hand (drawn but never cast)
- Running out of cards in mid/late game
- Tempo analysis: when does the deck "come online"
- Commander cast count distribution

### Simulation Volume
**Decision:** Default 1,000 games, configurable via CLI flag.
- Target execution time well under 30 minutes with rule-based agents
- Goroutine parallelism across all CPU cores (each game is independent)

### Package Structure
**Decision:** Top-level packages, single Go module at repo root.

```
github.com/natefinch/council4       ← module root (go.mod here)
├── game/                           ← core types + rules engine (existing)
│   ├── id/                         ← unique object identifiers (existing)
│   ├── mana/                       ← mana colors, costs, pools (existing)
│   └── counter/                    ← counter types and tracking (existing)
├── agent/                          ← Agent interface + implementations
├── sim/                            ← game runner, parallel tournament
├── cards/                          ← generated card data + registry
├── deck/                           ← decklist parser
├── report/                         ← analytics + output
├── cmd/council4/                   ← CLI main
└── docs/
    ├── adr/                        ← architecture decision records
    └── research/                   ← reference documents
```

- Delete the existing `game/go.mod` — module moves to repo root
- Dependency direction: `agent` and `sim` depend on `game`; `cmd/council4` depends on everything

## Existing Codebase

The `game/` package already has solid type definitions covering:
- `Game` struct (4 players, battlefield, stack, turn state, combat)
- `CardDef` and `CardInstance` (immutable card data + per-game instances)
- `Player` (life, zones, commander tracking, mana pool, special designations)
- `Permanent` (status flags, counters, attachments, timestamps)
- `TurnState` (phases, steps, priority, land drops, extra turns)
- `Zone` (library, hand, graveyard, exile, command — ordered, with face-down tracking)
- `Stack` (LIFO spell/ability resolution)
- `CombatState` (attackers, blockers, damage assignment)
- Sealed `AbilityBody` variants plus categorized `CardFace` ability fields (activated, triggered, static, spell, mana, loyalty, replacement)
- Leaf packages: `id/` (unique IDs), `mana/` (colors, costs, pools), `counter/` (counter types)

**What's missing (needs to be built):**
- Rules engine: `LegalActions()`, `ApplyAction()`, phase transitions, state-based actions
- `Clone()` method on `Game` (needed for future MCTS, useful for testing now)
- `GetObservation()` / `PlayerObservation` type (fog-of-war filtering)
- `ResampleFromInfostate()` (needed for MCTS determinization — v2)
- Effect primitive execution system
- Card implementation registry
- Agent interface and rule-based implementation
- Simulation harness (game runner, parallel tournament)
- Decklist parser
- Analytics collection and reporting
- CLI

## Reference Documents

All in `docs/research/`:
- `card-game-ai-research.md` — comprehensive AI guide (architecture, MCTS, agents, Go guidance)
- `MTG-GLOSSARY.md` — canonical MTG term definitions
- `COMMANDER-STRATEGY.md` — Commander-specific strategy patterns
- `COMMANDER-AGENT-PLAYBOOK.md` — how AI agents should approach Commander decisions
- `CARD-TEXT-PARSING.md` — parsing oracle text for card implementation generation
- `MTG-General-Research.md` — broader MTG mechanics research
- Magic Comprehensive Rules — official rules referenced externally; local full-text copies are not committed
