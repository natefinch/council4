# zone package

`mtg/game/zone` defines Magic game-zone vocabulary and the ordered card
collections used for player-owned zones. It is a leaf package that depends only
on `mtg/game/id`.

## Zone types

Use `Type` and its constants wherever data identifies a game zone:

```go
var from zone.Type = zone.Graveyard

if from.IsPublic() {
	// All players can inspect cards in this zone.
}
```

`None` represents the absence of a zone. The remaining constants correspond to
the zones defined by CR 400: `Library`, `Hand`, `Battlefield`, `Graveyard`,
`Stack`, `Exile`, and `Command`.

## Card collections

`Zone` stores card instance IDs for player-owned zones:

```go
library := zone.New(zone.Library)
library.AddToBottom(cardID)
top, ok := library.Top()
```

The top is index zero. `Add` places a card on top, `AddToBottom` appends it,
and `All` returns a copy in zone order. `Shuffle` requires an explicit
`*rand.Rand` so callers control determinism.

`Zone` also tracks face-down cards for zones such as exile. The shared
battlefield and stack use their richer runtime representations in `mtg/game`
rather than `zone.Zone`.
