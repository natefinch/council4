# counter package

`mtg/game/counter` models counters that can be placed on players, permanents, and other game objects.

The package currently focuses on permanent counters used by `game.Permanent`. Player counters such as poison, energy, and experience are represented directly on `game.Player` for now.

## Main types

### Kind

`Kind` identifies a counter type, such as:

- `PlusOnePlusOne`
- `MinusOneMinusOne`
- `Loyalty`
- `Charge`
- `Stun`
- `Shield`

Add new counter kinds here when card implementations need them.

### Set

`Set` stores counter counts by kind:

```go
var counters counter.Set
counters.Add(counter.PlusOnePlusOne, 2)
counters.Remove(counter.PlusOnePlusOne, 1)
count := counters.Count(counter.PlusOnePlusOne)
```

`Add` automatically cancels `+1/+1` and `-1/-1` counters against each other according to Magic rules.

## Package boundaries

This is a leaf package. It should not import `mtg/game` or `mtg/rules`.
