# council4 command

`cmd/council4` is the CLI entry point for running Council4 simulations.

The current command runs a minimal hardcoded test game: four players each get a deck of basic Forests, all players use `agent.FirstLegal`, and the rules engine runs until players deck out and one winner remains.

This is intentionally not the final product CLI. It exists to exercise the current game loop from a terminal while the engine grows.

## Usage

```bash
go run ./cmd/council4
```

Useful flags:

```bash
go run ./cmd/council4 -seed 7 -deck-size 8 -verbose
```

- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many Forests each test deck contains.
- `-verbose` prints the per-turn action log.

## Current behavior

The minimal loop supports:

- Opening hands.
- Untap, draw, main phases, placeholder combat, cleanup.
- Playing one land per turn.
- Passing priority around the table.
- Player elimination when a player tries to draw from an empty library.

The command does not yet parse decklists, load real card data, cast spells, produce reports, or simulate combat.
