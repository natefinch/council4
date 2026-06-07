# cost package

`mtg/game/cost` models declarative costs printed on spells and abilities. It
is separate from `mtg/game/mana`, which models produced mana and mana pools.

Use this package for mana costs such as `{2}{W}`, `{X}{R}`, `{W/U}`, `{2/G}`,
`{W/P}`, `{C}`, or `{S}`, and for non-mana costs such as tapping, sacrificing,
discarding, paying life, or exiling cards.

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

### Additional

`Additional` describes one non-mana cost. Set `Kind` to the operation and use
the other fields only when that operation needs them:

```go
AdditionalCosts: []cost.Additional{
	cost.T,
	{
		Kind:               cost.AdditionalSacrifice,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
	},
}
```

For the common case where tapping the source is the only additional cost, use
the predeclared slice directly:

```go
AdditionalCosts: cost.Tap
```

Use `cost.T` as the individual tap entry only when combining it with other
additional costs.

`Amount` defaults to one for costs involving objects or cards. `Text` supplies
card-specific display text when the generic text for the kind is insufficient.
`MatchPermanentType` and `PermanentType` constrain battlefield objects;
`MatchCardType` and `CardType` constrain cards in other zones.

`Source` identifies the zone that supplies cards for costs that choose cards
outside the battlefield. `zone.None` leaves the default to the rules module.
The current default for `AdditionalExile` is the graveyard. Use an explicit
source when another zone is required:

```go
cost.Additional{
	Kind:   cost.AdditionalDiscard,
	Text:   "Discard this card",
	Source: zone.Hand,
}
```

`Source` uses the shared `zone.Type` vocabulary directly.

### Alternative

`Alternative` replaces the normal mana cost when selected. It may contain a
mana cost, additional costs, or both:

```go
cost.Alternative{
	Label:    "Flashback",
	ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
}
```

Required costs on the spell or ability still apply in addition to the selected
alternative's `AdditionalCosts`.

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

This package is a leaf package except for depending on other leaf vocabulary
packages such as `mtg/game/mana`, `mtg/game/types`, and `mtg/game/zone`. It must
not depend on `mtg/game` or `mtg/rules`; runtime payment selections and
mutations belong in those higher-level packages.
