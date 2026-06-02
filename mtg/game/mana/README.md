# mana package

`mtg/game/mana` models produced mana and runtime mana pools.

The package is pure data plus small helpers. It does not decide whether a spell can be cast or spend costs automatically; the rules engine does that.

## Main types

### Color

`Color` represents white, blue, black, red, green, and colorless produced mana.

Use `mtg/game/color.Color` for card colors and color identity. Use `mana.Color`
only for mana production, mana pools, and mana payment.

### Unit

`Unit` is one spendable unit of mana. It records the mana color and whether that
mana was produced by a snow source.

### Pool

`Pool` tracks mana currently available to a player:

```go
pool := mana.NewPool()
pool.Add(mana.G, 1)
pool.AddSnow(mana.C, 1)
ok := pool.Spend(mana.G, 1)
```

Mana pools are emptied by the rules engine as steps and phases end.

Internally, the pool stores mana as `Unit` values so rules can preserve
provenance such as whether mana was produced by a snow source. The simple `Add`,
`Amount`, and `Spend` APIs still work by color; use `AddSnow`, `SnowAmount`, and
`SpendSnow` when a rule specifically cares about snow mana.

The current rules engine can pay colored, true colorless, generic, X, hybrid,
mono-hybrid, phyrexian, and snow costs by consuming pool mana first and then
auto-tapping supported mana sources such as basic lands, mana rocks, and
non-summoning-sick mana dorks. Alternative-cost and typed additional-cost
selection live in `mtg/rules`; full cost-reducer and attack-tax handling remain
future rules-engine work.

## Package boundaries

This is a leaf package. It should remain independent of `mtg/game`, `mtg/rules`, cards, deck parsing, and reporting.
