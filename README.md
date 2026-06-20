# Council4

Council4 is a Go playtesting engine for Magic: The Gathering Commander decks. The goal is to run many automated 4-player games with AI-controlled agents and produce analytics about deck performance.

## Card support

<!-- card-support:start -->
Council4 currently supports **9,587 of 31,838 cards eligible for paper support (30.1%)**. The Scryfall Oracle Cards corpus contains 6,288 additional digital, special-format, memorabilia, or non-sanctioned-paper records that are excluded from that total. See [`supported.md`](./supported.md) and [`unsupported.md`](./unsupported.md) for the complete lists, and [`unsupported-reasons.md`](./unsupported-reasons.md) for capability-aware blocker planning.
<!-- card-support:end -->

Run `go run github.com/magefile/mage@v1.15.0 cardSupport` after card-support changes. The target reuses the Scryfall Oracle Cards corpus cached outside the repository, runs cardgen in ignored `.cardwork` scratch space, and updates the support documentation. Set `COUNCIL4_ORACLE_CARDS` to use a specific local corpus file; remove the cached file printed by the target to download the latest corpus.

## Parser coverage

[`parser-coverage.md`](./parser-coverage.md) tracks how completely the Oracle parser represents each eligible card's text as typed syntax, measured without running the compiler or lowering. It reports the parser-complete card percentage and exact-effect percentage, plus a ranked work queue of the uncovered grammar that the parser does not yet represent. Run `go run github.com/magefile/mage@v1.15.0 parserCoverage` to regenerate it.

The current implementation is an early rules-engine slice: land-only games still run, spell-mode games can cast simple creatures, draw/life spells, and player-damage spells through the stack, and combat-mode games can attack players with simple creatures.

## Run the CLI

Run the current land-only test game:

```bash
go run ./cmd/council4
```

Useful flags:

```bash
go run ./cmd/council4 -seed 1 -deck-size 8 -verbose -nopass
go run ./cmd/council4 -mode spells -seed 1 -verbose -nopass
go run ./cmd/council4 -mode combat -seed 1 -verbose -nopass
```

- `-mode` chooses `land`, `spells`, or `combat`.
- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many basic Forests each land-mode test deck contains.
- `-verbose` prints the per-turn draw, action, resolve, combat damage, creature damage, death, and loss log.
- `-nopass` omits pass actions from verbose log output.

Example output:

```text
Council4 test game
Seed: 1
Mode: land
Deck size: 8 Forests per player
Turns: 7
Winner: Player 4
Battlefield permanents: 4
```

## Build and test

```bash
go test ./...
go vet ./...
```

The module root is the repository root:

```text
github.com/natefinch/council4
```

## Repository layout

```text
.
├── cmd/
│   └── council4/          # CLI entry point
├── mtg/
│   ├── game/              # Core game state data types
│   │   ├── action/        # Player action data
│   │   ├── counter/       # Counter types and counter sets
│   │   ├── id/            # Unique object identifiers
│   │   └── mana/          # Mana colors, symbols, costs, pools
│   ├── rules/             # Rules engine and minimal game loop
│   └── agent/             # AI agent implementations
├── docs/
│   ├── adr/               # Architecture decision records
│   └── research/          # MTG and AI reference documents
└── CONTEXT.md             # Shared project vocabulary
```

Every Go package directory has its own `README.md` describing that package and how to use it.

## Current engine capabilities

Implemented:

- 4-player Commander game state.
- Deterministic game setup via seeded RNG.
- Opening hands and card drawing.
- Turn progression through beginning, main, combat, ending, and cleanup.
- Multiplayer priority passing.
- Legal actions for playing lands, casting simple spells, and passing.
- Auto-tap payment for normal colored and generic costs using basic lands.
- Stack resolution for creature, instant, and sorcery spells.
- Effect primitives for draw, gain life, lose life, and player damage.
- Runtime player targets for simple targeted spells.
- Combat steps, compact attacker and blocker declarations, combat damage to players and creatures, and lethal creature damage cleanup.
- State-based player elimination for 0 life, poison, commander damage, and failed draws.
- Loss logs with reasons for player eliminations.
- Simple `agent.FirstLegal` and `agent.SimpleCaster` agents for deterministic test games.

Not implemented yet:

- Decklist parsing.
- Card registry and generated card definitions.
- Explicit mana ability actions and advanced costs.
- Advanced combat mechanics such as multi-blocking, evasion, first strike, double strike, trample, deathtouch, indestructible, and prevention.
- Reports and analytics output.

## Documentation

- [`CONTEXT.md`](./CONTEXT.md) defines project vocabulary.
- [`docs/playtest-rule-coverage.md`](./docs/playtest-rule-coverage.md) summarizes supported rule coverage and known limitations for the playtest path.
- [`docs/adr/`](./docs/adr/) records architectural decisions.
- Package-level READMEs document implementation details and usage for each package.
