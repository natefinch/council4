# Payment Planner

`mtg/rules/payment` contains the rules logic for building, validating, and
applying **Payment Plans**. It is a behavior package under `mtg/rules`: card and
game data still live in `mtg/game`, while `payment` owns cost-option selection,
mana-source selection, additional-cost planning, and applying committed plans.

## Interface

Callers construct a `Planner` with a `State` adapter:

```go
planner := payment.New(state)
plan, ok, reason := planner.Complete(payment.ForSpell(payment.SpellRequest{...}), payment.Partial{})
```

The complete-plan API is used by every Engine costed-action route:

```go
request := payment.ForSpell(spellRequest) // also ForAbility and ForGeneric
plan, ok, reason := planner.Complete(request, payment.Partial{})
options := planner.Options(request, partial, payment.Step{
    Kind: payment.StepComponent, Index: componentIndex,
})
result, ok, reason := planner.Apply(plan)
```

`Plan` is an immutable witness containing IDs rather than mutable game
references. Its accessors return independent copies. `Partial` fixes any subset
of cost-option, X, Phyrexian, grouped-cost, component or nonmana-mechanic choices;
missing choices remain undecided. Every returned option carries a witness.
Component steps select one additional object or counter at a time; threshold
steps can finish once sufficient or continue selecting. Component boundaries
preserve every selected object, including selections above a threshold.
`Apply` revalidates the exact selection and allocation against current state,
including in a game clone, and rejects stale witnesses before mutating anything.
It does not substitute another payment.

The completion search plans additional-cost selections and mana together.
It backtracks over finite resource assignments, with search-local memoization
of failed mana resource states, not game clones or speculative activations.
Shared resource matching rejects impossible fixed-count payments before subset
enumeration, and later mandatory source-counter kinds are reserved up front.
For generic mana, an optimistic total of pool mana, source outputs, and
cost mechanics rejects impossible amounts before trying individual land
subsets. It never establishes that a payment is legal. A mana ability printed
on a card in hand requires activation-history planning only if it functions
from hand.
Fixed-count and threshold selections share the existing eligibility predicates.
Costs reserve resources by what they consume: two zone moves cannot share an
object, but a tap followed by a sacrifice/return can (CR 601.2h). Phyrexian life
payments compete with all other life and mana costs. Random-discard components
remain separate batches; random and library-to-public-zone costs are paid last.

The [plan-then-commit design](../../../docs/adr/0013-plan-then-commit-payment.md)
is implemented across priority/resolution spells, activations, costed special
actions and generic rule/effect/attack payments by `Engine.costedAction`.
`Engine.PaymentPlanner` binds the planner to canonical resource evaluators and
mana activation operations. Resource projections are read-only; no ability is
speculatively executed and no normal payment clones a game. Copy-on-write
projections clear every derived read cache; query-local frames exist only during
pure `Complete`, `Options` and pending-target searches, never agent callbacks or
mutation. Life, counter, damage, production and triggered-mana forecasts reuse
canonical evaluators rather than approximating their effects.

Present-pool payments are tried first, then canonically verified automatic-source
histories, then complete resource search. Fixed allocations are checked at their
exact depth; exhaustion means no branch reaches the next depth, not an arbitrary
run budget. Each immutable node can share its cost-paid view across output
alternatives, not across occurrences or player choices.

Unsupported effect, condition, replacement or trigger contexts are explicitly
opaque. The shared procedure checkpoints only an opaque activation actually
selected during payment, never a broad opaque context. Its whole-world, ID,
log and exact random-word restoration follows CR 734.1–2; irreversible library
effects retain legal mana actions rather than replaying them. Invalid source
announcements can still be canceled under CR 601.2/602.2 before payment. There
are no clone probes, inconclusive-payable results or proposal-withholding retries.

`StepManaAbility` and `Partial.ManaAbilities` offer completion-preserving source
choices; `ManaActivationKey` binds payments and outputs to independent
occurrences, including optional repeated Altar activations.
`ApplyResult.ManaBefore` records
the pool after planned production and before the final spend, for exact rider
accounting. The locked total, exact caller choices and receipts are owned by the
immutable witness. Existing compatibility APIs remain for their callers without
introducing a second Engine payment procedure.

The `State` interface is the seam between the planner and the rules engine. It
provides game queries and mutations that require rules knowledge, such as
effective controllers, permanent characteristics, cost modifiers, tapping mana
sources, zone transitions, and discard/sacrifice handling. In production,
`mtg/rules` implements this seam with `rulesPaymentState` and exposes it through
`Engine.PaymentPlanner`; `paymentOrch` retains compatibility entry points.

Rules edition: [official Comprehensive Rules effective 2026-11-06](https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
SHA-256 `524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829`.
Payment sections read: 601.2, 602.2, 605.1a/605.3, 608.2g and 734.1–2
(not historical section 733). Intrinsic library-moving costs/effects are
nonmana under 605.1a, including Millikin and Deranged Assistant; no such
payment window is enabled. The old Milling Prism example in #3628 is not a
valid current-edition reproduction. General standalone/enclosing library-window
rules remain; its residual claim needs valid gameplay evidence, not more
payment scope.

## Requests

- `SpellRequest` checks or pays a spell's selected cost option. It includes the
  casting player, card, source zone, X value, kicker flag, and optional payment
  preferences. Its `SpendAnyMana` field carries a spending permission such as
  "spend mana as though it were mana of any color to cast it" (CR 118.14,
  609.4b). It is not an alternative cost: every cost option keeps its normal,
  alternative, additional, kicker, and splice mana, and only the final mana
  payment lets any mana pay colored requirements. `ManaSpendAsAnyColor` still
  requires colorless mana for `{C}`; `ManaSpendAsAnyType` lets any mana pay
  `{C}` too. Convoke, improvise, and delve are planned from the unchanged cost,
  and the recorded pool spend keeps the actual mana spent.
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
  every declared attacker from automatic mana-source planning. A set `Spell`
  checks mana in that spell's casting context, and `SpendAs` carries the same
  spending permission as `SpellRequest.SpendAnyMana` (legal X bounds for a
  spell cast from exile under such a rider).

`Preferences` records choices that the rules engine already collected from the
agent or deterministic fallback: alternative-cost index, phyrexian mana-vs-life
choices, and selected permanents/cards for additional costs. The planner treats
preferences as hints, preserving legal selections and falling back when a hint
is stale unless `StrictReplay` is set. Unlike this compatibility policy, choices
fixed in `Partial` are constraints and never fall back.

## Implementation

The mutable application details stay private. Callers should not depend on
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
