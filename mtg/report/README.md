# mtg/report

The report package turns simulation output into an actionable deck performance
report: a human-readable text summary and a detailed JSON file about the deck
under test.

## Decoupled input

A report is produced purely from a `sim.SimulationResult` — the per-game
`rules.GameResult` data, with the event stream and end-state folded in by
`RunGame`. The report layer never touches a live `*game.Game`, so the same input
can be saved, replayed, and re-reported, and the report code stays independent of
the engine.

To support this, `rules.GameResult` carries, in addition to the outcome and
per-turn logs:

- `Events` — the game's full `game.Event` stream.
- `EndState` — each seat's final life, elimination, remaining hand, and library
  size (for end-of-game analysis such as cards stranded in hand).
- `Cards` — every card instance's public name and owner, so event and end-state
  consumers can attribute cards by name and to the deck that owns them.

## Usage

```go
rep := report.Generate(simResult, report.Options{
    TestedSeat: testedSeat,            // 0-based seat of the deck under test
    DeckNames:  [game.NumPlayers]string{...},
})
_ = rep.WriteText(os.Stdout)   // human-readable summary
_ = rep.WriteJSON(jsonFile)    // detailed JSON report
```

`Generate` builds the structured `Report`; `WriteText` and `WriteJSON` render it.
Outcome, per-card, mana/curve, and interaction metrics are layered onto this
envelope by later analysis.
