# id package

`mtg/game/id` defines unique identifiers for game objects and card instances.

Use this package when a value needs stable identity within a single game: cards in zones, permanents on the battlefield, stack objects, and similar objects tracked by the engine.

## Main types

### ID

`ID` is the identifier type used throughout the game engine.

```go
var cardID id.ID
```

The zero value is reserved as "no ID" in several structs, such as optional planeswalker or battle attack targets.

### Generator

`Generator` creates unique IDs:

```go
var gen id.Generator
next := gen.Next()
```

`game.Game` owns an `IDGen` and uses it when creating `CardInstance`, `Permanent`, and stack object IDs.

## Package boundaries

This is a leaf package. It must not import `mtg/game` or any rules-engine package.
