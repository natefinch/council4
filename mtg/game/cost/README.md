# cost package

`mtg/game/cost` models printed mana costs for spells and abilities. It is
separate from `mtg/game/mana`, which models produced mana and mana pools.

Use this package anywhere a card or ability stores a cost such as `{2}{W}`,
`{X}{R}`, `{W/U}`, `{2/G}`, `{W/P}`, `{C}`, or `{S}`.

## Main types

### Mana

`Mana` is an ordered slice of printed cost symbols:

```go
cost := cost.Mana{
	cost.O(2),
	cost.W,
	cost.U,
}
```

Use `ManaValue()` for the cost's mana value outside the stack. Variable `X`
contributes 0 there. Use `Colors()` to inspect mana colors present in the
printed cost symbols.

An absent mana cost is represented by an absent `opt.V[cost.Mana]` on the card
or ability field. An explicit zero cost is `cost.Mana{cost.O(0)}`.

### Symbol

`Symbol` represents one printed cost symbol. Predeclared symbols cover common
single-symbol costs:

```go
cost.W // {W}
cost.U // {U}
cost.B // {B}
cost.R // {R}
cost.G // {G}
cost.C // {C}
cost.X // {X}
cost.S // {S}
```

Helper constructors cover symbols with parameters:

```go
cost.O(3)                         // {3}
cost.HybridMana(mana.W, mana.U)   // {W/U}
cost.Twobrid(mana.W)              // {2/W}
cost.PhyrexianMana(mana.W)        // {W/P}
```

## Package boundaries

This package is a leaf package except for depending on `mtg/game/mana` for mana
color vocabulary in cost symbols. It should not depend on `mtg/game` or
`mtg/rules`.
