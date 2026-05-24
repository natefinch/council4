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
go run ./cmd/council4 -mode combat -seed 7 -verbose -nopass
```

- `-mode` chooses `land` for the original land-only game, `spells` for a Phase 2 spell test deck, or `combat` for a simple creature combat deck.
- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many Forests each land-mode test deck contains.
- `-verbose` prints the per-turn draw, action, resolve, combat damage, creature damage, death, and loss log in chronological order.
- `-nopass` omits pass actions from verbose log output.

## Current behavior

Land mode supports:

- Opening hands.
- Untap, draw, main phases, combat, cleanup.
- Playing one land per turn.
- Passing priority around the table.
- Player elimination when a player tries to draw from an empty library.
- Loss reasons in the summary and verbose log.

Spell mode uses Forests plus simple hardcoded creature, draw, life-gain, and player-damage spells. It exercises auto-tap mana payment, cast actions, stack resolution, effect primitives, player targets, and verbose cast/resolve output.

Combat mode uses Forests plus simple hardcoded vanilla, Haste, Vigilance, and Defender creatures. It exercises attacker declarations, blocker declarations, player and creature combat damage, creature death cleanup, combat loss logs, and chronological verbose combat output. The `FirstLegal` agent always chooses the first productive attack and the first offered block, so combat mode intentionally favors deterministic combat over strategic multiplayer choices.

The command does not yet parse decklists, load real card data, produce reports, or simulate advanced combat mechanics like multi-blocking, evasion, first strike, trample, or prevention.
