# Council4

Council4 is a Go playtesting engine for Magic: The Gathering Commander decks. The goal is to run many automated 4-player games with AI-controlled agents and produce analytics about deck performance.

The current implementation is an early minimal loop: players draw opening hands, untap, draw, play one land when possible, pass priority, and eventually lose when they try to draw from an empty library.

## Run the CLI

Run the current minimal test game:

```bash
go run ./cmd/council4
```

Useful flags:

```bash
go run ./cmd/council4 -seed 1 -deck-size 8 -verbose -nopass
```

- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many basic Forests each hardcoded test deck contains.
- `-verbose` prints the per-turn draw and action log.
- `-nopass` omits pass actions from verbose log output.

Example output:

```text
Council4 minimal test game
Seed: 1
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
- Turn progression through beginning, main, combat placeholder, ending, and cleanup.
- Multiplayer priority passing.
- Legal actions for playing lands and passing.
- State-based player elimination for 0 life, poison, commander damage, and failed draws.
- Loss logs with reasons for player eliminations.
- A simple `agent.FirstLegal` agent that plays a land when possible and otherwise passes.

Not implemented yet:

- Decklist parsing.
- Card registry and generated card definitions.
- Mana production and payment.
- Spell casting and stack resolution.
- Combat damage.
- Reports and analytics output.

## Documentation

- [`CONTEXT.md`](./CONTEXT.md) defines project vocabulary.
- [`docs/adr/`](./docs/adr/) records architectural decisions.
- Package-level READMEs document implementation details and usage for each package.
