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
- `Age`
- `Poison`
- `Energy`
- `Experience`

Add new counter kinds here when card implementations need them.

The Oracle executable compiler supports named Stun counter placement: the untap
step removes one stun counter from a permanent instead of untapping it (CR
122.6f). One or more Finality counters on a permanent create a single
replacement effect (CR 122.1h): if it would be put into a graveyard from the
battlefield, it is exiled instead. `rules.finalityCounterReplacement`
synthesizes it from the counter itself (not an ability, so ability loss does
not remove it); it competes with other replacements under the affected
controller's choice (CR 616.1) and leaves other destinations unchanged. All
kinds remain available for runtime mechanics and manual card definitions.
`Age` is a permanent counter used by the cumulative-upkeep template.

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
