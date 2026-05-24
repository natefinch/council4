# mana package

`mtg/game/mana` models Magic mana concepts: colors, symbols, printed costs, runtime mana pools, and Commander color identity.

The package is pure data plus small helpers. It does not decide whether a spell can be cast or spend costs automatically; the rules engine does that.

## Main types

### Color

`Color` represents white, blue, black, red, green, and colorless mana.

Use `AllColors()` when deterministic color ordering matters.

### Symbol

`Symbol` represents one printed mana symbol, including generic, colored, hybrid, phyrexian, snow, colorless, and variable `X` symbols.

### Cost

`Cost` is an ordered list of `Symbol` values:

```go
cost := mana.Cost{
	mana.GenericMana(2),
	mana.ColoredMana(mana.White),
	mana.ColoredMana(mana.Blue),
}
```

Use `ManaValue()` for mana value and `Colors()` for colors appearing in the printed cost.

### Pool

`Pool` tracks mana currently available to a player:

```go
pool := mana.NewPool()
pool.Add(mana.Green, 1)
ok := pool.Spend(mana.Green, 1)
```

Mana pools are emptied by the rules engine as steps and phases end.

The current rules engine can pay colored, true colorless, generic, and X costs by consuming pool mana first and then auto-tapping supported mana sources such as basic lands, mana rocks, and non-summoning-sick mana dorks. Hybrid, phyrexian, snow, alternate-cost, and cost-reducer handling are intentionally deferred to the rules engine.

### ColorIdentity

`ColorIdentity` represents a Commander deck's color identity. Use it to validate whether a card belongs in a commander's deck.

## Package boundaries

This is a leaf package. It should remain independent of `mtg/game`, `mtg/rules`, cards, deck parsing, and reporting.
