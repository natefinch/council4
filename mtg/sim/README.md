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
changing results (a separate concern, layered on top of this runner).

## Usage

```go
cfg := sim.Config{
    Configs:   fourPlayerConfigs, // [game.NumPlayers]game.PlayerConfig
    Games:     1000,
    Seed:      masterSeed,
    NewAgents: nil, // defaults to a deterministic FirstLegal agent per seat
}
results := sim.Run(cfg) // []rules.GameResult, one per game, in order
```

Provide a `NewAgents` factory to seat real agents. It receives each game's
derived seed, so any agent randomness stays reproducible — derive each seat's
RNG from `gameSeed` rather than sharing one source.
