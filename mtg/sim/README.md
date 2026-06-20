# mtg/sim

The simulation harness runs many games over the same four decks and collects
their results. It exists so a deck can be playtested across many games, not just
inspected one game at a time.

## Determinism

A run is fully reproducible. Every game derives its own seed from the run's
master seed (`GameSeed`), and from that seed builds its own `*rand.Rand`,
`rules.Engine`, and `game.Game`. Games never share mutable RNG or game state, so:

- `Run(cfg)` returns identical results for the same `Config`.
- `RunOne(cfg, i)` reproduces game `i` from a batch on its own.
- The per-game seed mixer (SplitMix64 over a golden-ratio-spaced counter) gives
  neighbouring games well-separated, uncorrelated seeds.

This per-game independence is also what lets games run in parallel without
changing results. `Config.Workers` sets the maximum number of games run
concurrently (zero means `GOMAXPROCS`); each game's result is written to its own
index, so `Run` returns the same slice for any worker count. A `NewAgents`
factory must therefore be safe to call from multiple goroutines.

## Usage

```go
cfg := sim.Config{
    Configs:   fourPlayerConfigs, // [game.NumPlayers]game.PlayerConfig
    Games:     1000,
    Seed:      masterSeed,
    Workers:   0,   // 0 -> GOMAXPROCS
    NewAgents: nil, // defaults to a deterministic FirstLegal agent per seat
}
result := sim.Run(cfg) // sim.SimulationResult
```

`Run` returns a `SimulationResult` — the stable hand-off the report layer
consumes. It holds every game's full `rules.GameResult` in order (`Games[i]` was
played with `Seeds[i]` and is reproducible via `RunOne(cfg, i)`), the master
seed, the game count, and any `Failures`. Convenience aggregations (`WinCounts`,
`DrawCount`, `FailureCount`) summarise outcomes. The harness retains full results
by default; for very large runs a caller can instead keep only `Seeds` and
reconstruct games on demand.

Provide a `NewAgents` factory to seat real agents. It receives each game's
derived seed, so any agent randomness stays reproducible — derive each seat's
RNG from `gameSeed` rather than sharing one source.

## Goldfish

`RunGoldfish` plays one Commander deck without opponents for an exact turn
budget. It uses the normal setup, mulligan, priority, mana, casting, trigger,
commander, and cleanup rules, but opponent targets and combat defenders are
absent. The default agent uses `GenericStrategy`.

```go
result := sim.RunGoldfish(sim.GoldfishConfig{
    Player:   playerConfig,
    Seed:     42,
    MaxTurns: 10,
})
```

`result.Turns` is the chronological turn-by-turn decision log, `result.Events`
contains the complete typed rules event stream, and `result.EndState` captures
the final hand, library, life, and commander-cast count. A successful
turn-limited run has `TurnLimitReached` set and does not report a multiplayer
winner.

## Replay

For debugging a specific game, `RecordGame(cfg, i)` plays game `i` with recording
agents and returns both its `rules.GameResult` and a `ReplayRecord`. The record
stores the per-game seed and, per seat, the agent's decisions as plain integers:
the index (within each call's legal-action list) of the action it took, and the
option selection it returned for each engine-mediated choice. Because it is just
ints, a `ReplayRecord` round-trips through JSON.

`Replay(configs, record)` reconstructs the game: it re-runs the recorded seed (so
the engine RNG reproduces every shuffle) and scripts each seat to repeat its
recorded decisions, returning a `GameResult` identical to the recorded one. The
record notes per seat whether the agent answered engine-mediated choices itself;
a seat that left choices to the engine's deterministic fallback is replayed with
an action-only agent so the engine falls back the same way, which keeps games
with non-choice agents (and choices where an empty selection is valid, such as a
failed search) exactly reproducible.

This is the heavier debug artifact; for ordinary reproduction, `RunOne(cfg, i)`
already replays any game from the master seed alone.

## Failure capture

A single game must not abort a long batch. `Run` plays each game under a recover,
so a panic — an engine bug, an unsupported card, or an illegal action (the engine
panics on an invalid applied action) — is caught, the rest of the batch still
completes, and the panic is recorded in `SimulationResult.Failures` as a
`GameFailure{Index, Seed, Reason, Stack, Unsupported}`. The failed game keeps its slot in
`Games` with the zero result, and `Failures` is ordered by game index regardless
of which worker hit the panic. `Unsupported` is set when the recovered value is a
`rules.UnsupportedError` (the engine reached a mechanic it does not yet support),
so a batch can tell "not implemented yet" failures from genuine defects. Because
every failure carries its seed, the game is reproducible for debugging via
`RunOne(cfg, i)` or replay.

## Benchmarks

`bench_test.go` tracks per-game runtime and batch throughput so a regression that
would push 1,000 games past the 30-minute budget (ADR 0003) shows up as a
measurable slowdown. They run the lightweight land-only smoke decks for a fast,
deterministic measurement; a heavier board-building workload that exercises the
uncached continuous-effect recompute (`mtg/rules/continuous.go`) is tracked
separately as follow-up optimization work.

Run them with:

```
go test ./mtg/sim/ -run '^$' -bench . -benchmem
```

Baseline (Apple M3 Max, GOMAXPROCS=16; indicative and order-of-magnitude —
absolute timings vary by hardware and run, but allocation counts are stable):

| Benchmark                  | time/op  | B/op   | allocs/op |
|----------------------------|----------|--------|-----------|
| `RunOneGame`               | ~0.2 s   | ~42 MB | ~2.0M     |
| `RunBatchSequential` (16)  | ~2.4 s   | ~669 MB | ~32M     |
| `RunBatchParallel` (16)    | ~1.4 s   | ~669 MB | ~32M     |

At roughly 0.2 s per game, 1,000 games run in a few minutes single-threaded and
faster in parallel — comfortably inside the 30-minute budget. The high per-game
allocation count (~2M allocations) comes mainly from the continuous-effect values
being recomputed uncached on every query, which is the primary optimization
target if the budget ever tightens.
