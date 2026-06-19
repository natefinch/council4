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

- Frequency counts — draws, casts, resolves, discards, and removals (a permanent
  dying) — read from the folded `Events` stream. `ZoneChanges` is the total
  number of zone-change events for the card: a coarse superset that already
  includes its draws, casts, discards, and removals (plus moves like bounce,
  mill, and exile), so it is not additive with those columns.
- `SeenInWins` / `SeenInLosses` — games in which the card was drawn or cast,
  split by outcome, to compare a card's record across wins and losses.
- `Stranded` — how often the card was left in the tested deck's hand at game end
  (read from the folded `EndState`), a sign it rotted rather than being played.

Cards are sorted most-active first (casts, then draws, then name).

## Mana & curve

The report's `ManaCurve` summarises how the deck under test developed and spent
its mana across the completed games:

- **Lands per turn** (played-land actions per tested turn) and **mana spent per
  game** over the number of **spells cast** (from the folded events).
- **Flood rate** (games that played many lands but cast few spells) and **screw
  rate** (games with several turns yet very few lands), plus a **no-land-drop
  rate** (the fraction of the tested deck's turns with no land played — a
  missed-land-drop proxy; distinguishing "held a land" from "had none" would need
  per-turn hand telemetry).
- **Expensive rot** — nonland cards drawn but never cast, weighted by mana value
  (`RotMVPerGame`) and listed most expensive first.

To support this, the folded `CardInfo` carries each card's `ManaValue` and
`Types` (so lands are identified and rot is valued).

## Tempo, commander, and interaction

The report also covers, for the deck under test across completed games:

- **Tempo** — the average turn the deck cast its first spell (comes online),
  combat damage dealt to opponents per game, and per active turn.
- **Commander** — average commander casts per game, the cast-count distribution,
  and a dependency rate (the fraction of the deck's wins in which it had cast its
  commander). The folded `EndState` carries each seat's `CommanderCasts`.
- **Opponent interaction** — how often an opponent's spell or ability targeted the
  tested player or one of its permanents (permanents are attributed to an owner
  via the enter-the-battlefield events, which carry both the permanent and card
  IDs). This is a proxy for targeted removal and disruption aimed at the deck;
  counters and untargeted board wipes are not yet attributed.
