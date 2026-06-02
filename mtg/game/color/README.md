# color package

`mtg/game/color` models Magic card colors: white, blue, black, red, and green.
It deliberately excludes colorless because colorless is a mana quality, not a
card color.

Use this package for card characteristics, color identity, protection from
colors, target predicates, continuous effects that change colors, and any other
rule text that refers to an object's color.

## Main types

### Color

`Color` is the five-color vocabulary for cards and objects:

```go
colors := []color.Color{color.Green, color.Red}
```

Use `Abbreviation()` when rendering Magic's one-letter symbols (`W`, `U`, `B`,
`R`, `G`). Use `AllColors()` when deterministic five-color ordering matters.

### Identity

`Identity` represents a Commander deck's color identity (CR 903.4). It stores
card colors, not produced mana colors:

```go
identity := color.NewIdentity(color.Green, color.Red)
ok := identity.Contains(color.Green)
```

Use it to validate whether a card belongs in a commander's deck.

## Package boundaries

This package is a leaf package. It should remain independent of `mtg/game`,
`mtg/rules`, cards, deck parsing, and reporting.

Use `mtg/game/mana.Color` instead when modeling produced mana or mana in a pool;
that type includes colorless mana.
