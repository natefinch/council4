# cost package

`mtg/game/cost` models declarative costs printed on spells and abilities. It
is separate from `mtg/game/mana`, which models produced mana and mana pools.

Use this package for mana costs such as `{2}{W}`, `{X}{R}`, `{W/U}`, `{2/G}`,
`{W/P}`, `{C}`, or `{S}`, and for non-mana costs such as tapping, untapping,
sacrificing, discarding, paying life, removing counters, or exiling cards.

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

A battlefield cost printed as a two-type union ("sacrifice an artifact or
creature") sets `MatchPermanentType` with `PermanentType` and the optional
`PermanentTypeAlt`; the cost is payable by any permanent of either type.

For the common case where tapping the source is the only additional cost, use
the predeclared slice directly:

```go
AdditionalCosts: cost.Tap
```

Use `cost.T` as the individual tap entry only when combining it with other
additional costs.

Put spell casting costs on `game.CardFace.AdditionalCosts`. Put activated and
mana ability costs on the corresponding ability body.

`Amount` defaults to one for costs involving objects or cards. `Text` supplies
card-specific display text when the generic text for the kind is insufficient.
`MatchPermanentType` and `PermanentType` constrain battlefield objects;
`MatchCardType` and `CardType` constrain cards in other zones.
`MatchCardColor` and `CardColor` constrain card costs such as revealing blue
cards. `SubtypesAny` requires a selected card to have at least one of the listed
subtypes:

```go
cost.Additional{
	Kind:        cost.AdditionalReveal,
	Amount:      1,
	Source:      zone.Hand,
	SubtypesAny: cost.SubtypeSet{types.Forest, types.Island},
}
```

Use `AmountFromX` for costs whose required count is the announced X value, such
as "Reveal X blue cards from your hand".

`ExcludeSource` keeps the ability's own source permanent out of the candidate
set, modeling "another" in costs such as "Sacrifice another creature". It is
mutually exclusive with the source-only kinds (`AdditionalSacrificeSource`,
`AdditionalExileSource`), which always act on the source itself:

```go
cost.Additional{
	Kind:               cost.AdditionalSacrifice,
	Amount:             1,
	MatchPermanentType: true,
	PermanentType:      types.Creature,
	ExcludeSource:      true,
}
```

`Source` identifies the zone that supplies cards for costs that choose cards
outside the battlefield. `zone.None` leaves the default to the rules module.
The current default for `AdditionalExile` is the graveyard, while
`AdditionalReveal` defaults to the hand. Use an explicit source when another
zone is required:

```go
cost.Additional{
	Kind:   cost.AdditionalDiscard,
	Text:   "Discard this card",
	Source: zone.Hand,
}
```

`Source` uses the shared `zone.Type` vocabulary directly.

`AdditionalUntap` represents `{Q}` and always untaps the source permanent.
`AdditionalRemoveCounter` removes `Amount` counters of `CounterKind` from the
source permanent:

```go
cost.Additional{
	Kind:        cost.AdditionalRemoveCounter,
	Amount:      1,
	CounterKind: counter.Charge,
}
```

### Alternative

`Alternative` replaces the normal mana cost when selected. It may contain a
mana cost, additional costs, or both:

```go
cost.Alternative{
	Label:    "Flashback",
	ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
}
```

Put spell alternatives on `game.CardFace.AlternativeCosts`. Activated ability
alternatives remain on the corresponding ability body.

Required costs on the spell or ability still apply in addition to the selected
alternative's `AdditionalCosts`.

`Alternative.Condition` can restrict when an option is available. The
commander-control condition requires a modeled commander permanent currently
controlled by the caster; it does not treat a commander in another zone as
present.

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
