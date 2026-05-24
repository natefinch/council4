# council4 command

`cmd/council4` is the CLI entry point for running Council4 simulations.

The current command runs hardcoded test games that exercise the rules engine while the real deck parser and card registry are still under construction.

This is intentionally not the final product CLI. It exists to exercise the current game loop from a terminal while the engine grows.

## Usage

```bash
go run ./cmd/council4
```

Useful flags:

```bash
go run ./cmd/council4 -mode land -seed 7 -deck-size 8 -verbose -nopass
go run ./cmd/council4 -mode spells -seed 7 -verbose -nopass
```

- `-mode` chooses `land` for the original land-only game or `spells` for a Phase 2 spell test deck.
- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many Forests each land-mode test deck contains.
- `-verbose` prints the per-turn draw and action log.
- `-nopass` omits pass actions from verbose log output.

## Current behavior

Land mode supports:

- Opening hands.
- Untap, draw, main phases, placeholder combat, cleanup.
- Playing one land per turn.
- Passing priority around the table.
- Player elimination when a player tries to draw from an empty library.
- Loss reasons in the summary and verbose log.

Spell mode uses Forests plus simple hardcoded creature, draw, life-gain, and player-damage spells. It exercises auto-tap mana payment, cast actions, stack resolution, effect primitives, player targets, and verbose cast/resolve output.

The command does not yet parse decklists, load real card data, produce reports, or simulate combat.
