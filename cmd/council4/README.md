# council4 command

`cmd/council4` is the CLI entry point for running Council4 simulations.

The command can run a game from four real Commander decklists, and also runs hardcoded synthetic test games that exercise the rules engine.

This is intentionally not the final product CLI. It exists to exercise the current game loop from a terminal while the engine grows.

## Usage

```bash
go run ./cmd/council4
```

### Decklist mode

Pass four decklist files (Moxfield/MTGO text format) with repeated `-deck` flags and choose which one is under test:

```bash
go run ./cmd/council4 -deck me.txt -deck opp1.txt -deck opp2.txt -deck opp3.txt -tested 1 -verbose
```

- `-deck` names a decklist file; give it exactly four times for a four-player game. Decklist mode takes precedence over `-mode`.
- `-tested` is the 1-based index of the deck being tested (defaults to 1).
- Card names are resolved against the full committed card registry. Unknown card names and Commander deck-legality problems (deck size, singleton nonbasics, color identity, legendary commander) are reported, and the game does not run, instead of panicking.

### Synthetic test modes

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

The command does not yet produce reports or simulate advanced combat mechanics like multi-blocking, evasion, first strike, trample, or prevention. Batch simulation across many games is handled by the forthcoming simulation harness.
