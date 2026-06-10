# counter package

`mtg/game/counter` models counters that can be placed on players, permanents, and other game objects.

Permanent counters use `counter.Set` on `game.Permanent`. Poison, energy, and
experience remain named fields on `game.Player`; the typed
`game.AddPlayerCounter` primitive validates and places those player-only kinds.

## Main types

### Kind

`Kind` identifies a counter type, such as:

- `PlusOnePlusOne`
- `MinusOneMinusOne`
- `Loyalty`
- `Charge`
- `Stun`
- `Shield`
- `Poison`
- `Energy`
- `Experience`

Add new counter kinds here when card implementations need them.

The Oracle executable compiler deliberately rejects named Stun and Finality
counter placement until their untap and zone-change mechanics are implemented
(#222 and #223). The kinds remain available for runtime mechanics and manual
card definitions.

### Set

`Set` stores counter counts by kind:

```go
var counters counter.Set
counters.Add(counter.PlusOnePlusOne, 2)
counters.Remove(counter.PlusOnePlusOne, 1)
count := counters.Count(counter.PlusOnePlusOne)
```

The rules engine calls `CancelOpposites` while applying state-based actions.

## Package boundaries

This is a leaf package. It should not import `mtg/game` or `mtg/rules`.
