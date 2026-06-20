# Payment Planner

`mtg/rules/payment` contains the rules logic for building, validating, and
applying **Payment Plans**. It is a behavior package under `mtg/rules`: card and
game data still live in `mtg/game`, while `payment` owns cost-option selection,
mana-source selection, additional-cost planning, and applying committed plans.

## Interface

Callers construct a `Planner` with a `State` adapter:

```go
planner := payment.New(state)
ok := planner.CanPaySpellCosts(payment.SpellRequest{...})
```

The `State` interface is the seam between the planner and the rules engine. It
provides game queries and mutations that require rules knowledge, such as
effective controllers, permanent characteristics, cost modifiers, tapping mana
sources, zone transitions, and discard/sacrifice handling. In production,
`mtg/rules` implements this seam with `rulesPaymentState` and exposes it through
the package-local `paymentOrch` adapter.

## Requests

- `SpellRequest` checks or pays a spell's selected cost option. It includes the
  casting player, card, source zone, X value, kicker flag, and optional payment
  preferences.
- `AbilityRequest` checks or pays an activated ability. It includes either the
  source permanent or a source card/zone for non-battlefield abilities, the
  ability definition, X value, and optional payment preferences.
- `GenericRequest` checks or pays standalone costs such as attack taxes, Ward,
  Cycling, Madness, Suspend, and resolution-payment effects, including mana
  taxes offered to the player identified by a triggering event.
  Resolution mechanics such as cumulative upkeep materialize dynamic exact-mana
  multipliers before sending the fixed result through this same planner,
  agent-choice, and application path. Combat combines
  the independent per-attacker attack-tax amounts into one request and excludes
  every declared attacker from automatic mana-source planning.

`Preferences` records choices that the rules engine already collected from the
agent or deterministic fallback: alternative-cost index, phyrexian mana-vs-life
choices, and selected permanents/cards for additional costs. The planner treats
preferences as hints and rejects them if they no longer describe a legal plan.

## Implementation

The package keeps the concrete plan types private. Callers should not depend on
mana-tap ordering, source enumeration, or additional-cost implementation
details except through behavior covered by tests.

Files are split by responsibility:

- `options.go` enumerates spell cost options such as normal cost, kicker, and
  alternative costs, filtering conditional alternatives against current game
  state each time legality or payment is checked. Source-zone cast permission is
  checked separately. Each option records the permission it uses: flashback costs
  require flashback permission, while an independent graveyard permission exposes
  normal and other alternative costs without flashback's exile semantics.
- `modifiers.go` applies cost increases, reductions, set values, and minimums.
- `additional.go` plans and applies sacrifice, source-only sacrifice, discard,
  exile, source-card exile, reveal, tap, untap, counter-removal, and life
  additional costs. Sacrifice costs use the rules-owned authoritative sacrifice
  seam so both sacrifice and zone-change events are emitted. A sacrifice cost
  with `ExcludeSource` (the "another" wording) drops the ability's own source
  permanent from the candidate set. Reveal costs
  validate the selected card at commit time, leave it in its source zone, and
  emit a card-revealed event.
- `plan.go` builds spell, ability, and generic payment plans. `paySpellCosts`,
  `payAbilityCosts`, and `payGenericCost` each also report the exact per-unit
  pool mana the plan consumed (`clonePoolSpend` over the plan's per-unit pool
  spend) so the rules engine can resolve mana-spend riders against the precise
  units spent on every payment path rather than a gross before/after delta.
  Before planning, restricted tagged units are removed from the spendable pool
  unless the spell satisfies their closed condition; nonspell payments cannot
  spend them. Same-color units retain independent rider state, including a
  captured chosen creature subtype.
- `sources.go` discovers and orders mana sources, including timing-restricted
  tap and untap mana abilities, faithful Treasure-style
  tap-plus-sacrifice/color-choice mana abilities, Convoke, and Delve. The
  Treasure shape is recognized structurally and fails closed on extra costs,
  targets, dynamic choices, riders, or non-mana effects. `IsAutomaticManaAbility`
  remains limited to fixed-output tap/untap abilities, so choice-bearing sources
  remain available as standalone agent actions even though the payment planner
  can activate the safe Treasure shape while committing a cost.
- `apply.go` mutates mana pools, permanents, graveyards, exile, and life totals
  when a validated plan is committed.

## Testing

Most behavioral coverage remains in `mtg/rules` integration and
characterization tests because payment depends on effective game state,
continuous effects, zone mutations, and event emission. Add package-local tests
here only for planner behavior that can be exercised through a small fake
`State` without re-creating the whole rules engine.
