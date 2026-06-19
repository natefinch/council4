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

## Outcome metrics

The report's `Outcome` covers the deck under test across the completed games
(failed games are excluded):

- **Win rate** and win/loss counts.
- **Finishing position** — a competition ranking per game (the winner is 1st, a
  survivor outranks an eliminated seat, a seat eliminated later outranks one
  eliminated earlier, and draw survivors tie), reported as an average and a
  per-position count.
- **Game length** — the turn-count distribution (min/avg/max plus a histogram),
  and the same split into **turns to win** and **turns to loss**.

## Per-card performance

The report's `Cards` lists per-card metrics for the deck under test, aggregated
by card name across the completed games (only cards the tested deck owns):

- Frequency counts — draws, casts, resolves, discards, removals (a permanent
  dying), and generic zone changes — read from the folded `Events` stream.
- `SeenInWins` / `SeenInLosses` — games in which the card was drawn or cast,
  split by outcome, to compare a card's record across wins and losses.
- `Stranded` — how often the card was left in the tested deck's hand at game end
  (read from the folded `EndState`), a sign it rotted rather than being played.

Cards are sorted most-active first (casts, then draws, then name).

Mana/curve and interaction metrics are layered onto this envelope by later
analysis.
