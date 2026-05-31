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
- `AbilityRequest` checks or pays an activated ability. It includes the source
  permanent, ability definition, X value, and optional payment preferences.
- `GenericRequest` checks or pays standalone costs such as attack taxes, Ward,
  Cycling, Madness, Suspend, and resolution-payment effects.

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
  alternative costs.
- `modifiers.go` applies cost increases, reductions, set values, and minimums.
- `additional.go` plans and applies sacrifice, discard, and life additional
  costs.
- `plan.go` builds spell, ability, and generic payment plans.
- `sources.go` discovers and orders mana sources, including Convoke and Delve.
- `apply.go` mutates mana pools, permanents, graveyards, exile, and life totals
  when a validated plan is committed.

## Testing

Most behavioral coverage remains in `mtg/rules` integration and
characterization tests because payment depends on effective game state,
continuous effects, zone mutations, and event emission. Add package-local tests
here only for planner behavior that can be exercised through a small fake
`State` without re-creating the whole rules engine.
