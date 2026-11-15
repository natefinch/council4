# Plan-then-commit cost payment

Status: accepted; implementation tracked by epic
[#3718](https://github.com/natefinch/council4/issues/3718).

## Context

Casting a spell (CR 601.2a–i), activating an ability (CR 602.2a–b, which
makes the rest of activation identical to casting), taking a special action
with a cost (CR 116.2b/f/h/k), and paying for an effect (CR 605.3a, 608.2g)
all follow one procedure: announce choices, choose targets, determine the
total cost, optionally activate mana abilities, then pay in full.

The payment planner in `mtg/rules/payment` is already shared for cost math,
but the procedure around it is not. It is implemented separately for priority
casts, mutate, prepared copies, face-down casts, free casts, paid resolution
casts, madness, activated abilities, special actions and generic payments.
Each copy has its own pre-offer feasibility check (`activationFeasible`, or
`manualManaMayPay`, which searches mana-ability orders on clones up to 128
runs and offers the action when inconclusive), its own payment-time mana
window, and its own rewind (`castProposals` withholding, `castAttempts`,
`restoreCastSourceCard`, partial checkpoints in `payment_mana_rollback.go`).

Agents answer one question at a time and can commit to choices with no legal
completion. The engine then rewinds under CR 733 and tracks withheld choices
so the agent can't repeat them. From 2026-10-24 to 2026-11-05, production
filed about 47 payment and cost issues, many the same gap found one route at
a time (#3627 → #3629 → #3630 → #3691–#3695) or artifacts of partial rewind
(#3694 leaked object IDs, #3576 and #3695 lost retry choices, #3633 an
inconclusive search treated as payable).

## Decision

1. **One costed-action procedure.** Every route supplies a small descriptor:
   what is moved or created at CR 601.2a/602.2a, its announcements, its target
   specification, its payment request and its commit step (CR 601.2i). The
   procedure owns step order, choice prompts, payment and events.
2. **Complete plans.** The planner returns an immutable, clone-safe
   `payment.Plan` covering cost option, X, Phyrexian payments, every
   additional-cost selection, convoke/delve/improvise and mana-source
   allocation. It plans all components jointly.
3. **Every offered choice has a completion.** `Planner.Complete(request,
   partial)` returns a witness plan or a reason. Each prompt offers only
   options that keep a witness, and legal-action enumeration offers an action
   only when the empty partial has one. After the last choice the procedure
   applies the plan atomically in CR 601.2h order.
4. **Side-effecting mana abilities are planned.** Sacrifice, life, damage,
   counters, filters, bounded choices and deterministic triggered mana
   abilities (CR 605.1b) are planner-visible activation schemas whose
   resources are reserved. When choosing among them matters, the agent
   chooses through the same filtered prompts.
5. **Rewind is the rare fallback.** A mana ability the planner can't model is
   *opaque*. If a player activates one during payment (CR 605.3a), the
   procedure takes one whole-game snapshot before it. If the payment then
   can't complete, the snapshot is restored, including object IDs, RNG and
   logs, and the player keeps priority (current CR 734.1–2). No other path rewinds.

## Why this is enough

Under the rules effective 2026-11-06, abilities whose cost or effect moves a
card to or from a library are not mana abilities (CR 605.1a). Research for
#3718 found no engine mana ability with random or hidden output. Opaque
abilities are therefore an engine-modeling limit, not a common rules case.
Costs with random elements, such as a random discard, are still deterministic
to plan: the planner knows whether they can be paid, and only their outcome is
random.

## Consequences

- Fixes land once, in the procedure or planner, instead of per route.
- Agents and search (ADR 0011) never explore doomed branches, and choice
  points stay at stable seams.
- `castProposals`/`castAttempts` withholding, `manualManaMayPay`, the
  64-sample withholding interaction, partial checkpoints and per-route
  rollback closures are removed as routes migrate (#3725).
- Generic payments (ward, pay-or, attack taxes, cumulative upkeep) gain
  planned mana abilities, which they currently lack.
- The planner gains complexity: joint planning and activation schemas. It
  must stay deterministic and fail closed: an unrecognized shape is opaque,
  never approximated.

## Mana-schema implementation

`Engine.PaymentPlanner(game)` binds `Complete` and `Apply` to the executing
Engine. Planning uses structural activation schemas and a resource-only
copy-on-write evaluation view, not `Game.Clone` or speculative activation.
The view resets transient evaluation/choice frames; the payment adapter
rejects mutations and translates permanent references by ObjectID.

Plans retain exact activation costs, bounded choices, production units and
spend riders. The shared canonical production, life, counter-replacement and
trigger-matching evaluators are used for forecasting. Ordinary tap sources
retain a fast path; other supported activations reserve their input mana,
life, counters and source jointly, including effects that make a later
nonmana cost payable. Unrecognized effects, selectable replacement ordering,
finite damage shields/redirection, optional mana-trigger bodies and
non-deterministic mana-trigger choices remain opaque. Conditions outside
battlefield/resource predicates, bounded trigger-use policies and cyclic
mana-trigger graphs also remain explicit modeling boundaries rather than
assumed successes.

`Partial.ManaAbilities` and `StepManaAbility` preserve the choice to use or
avoid an identified source; `Options` enumerates component choices across
projected resources and retains only alternatives with a joint completion.
`Apply` revalidates the exact witness before activating through the normal
Engine mana operation. CR 605.1b mana triggers resolve immediately, once,
without a residual stack object, using normal triggered-body optional decisions
and resolution-time intervening-condition checks. Library-moving abilities are
excluded under CR 605.1a (official rules effective 2026-11-06).
An optional `TurnLog` supplied to `PaymentPlanner` records each planned
activation through the normal source-labelled payment-mana log.

`ApplyResult.ManaBefore` is the pool immediately before the final payment,
including planned production. Spell-commit rider consumers must use this
receipt with `Spell.PoolSpend`, rather than a snapshot captured before
activation. Input/filter and non-spell rider consumption is handled during
application. Existing payment callers retain their APIs pending migration.

## Migration

### Activated abilities and costed special actions (#3721)

`Engine.costedAction` owns proposal setup, announcement choices, targets, the
locked `payment.Request`, completion-preserving cost and mana decisions,
application, and commit. `describeActivation` supplies source identity and the
rules-specific commit data for permanent/general/equip/loyalty, hand,
graveyard, cycling, ninjutsu and hand-mana activations. Suspend, foretell, plot
and turn-face-up special actions use the same descriptor/procedure. The existing
action kinds contain no other mana-costed special action; land plays retain
their existing entry/replacement decisions. Casting procedures are unchanged
pending #3722.

Nonmana abilities are on the stack before any announcement/payment callback
(CR 602.2a). Hidden-zone activation sources are publicly revealed first;
`PlayerObservation.AnnouncedCards` exposes only these disclosures, not a hidden
zone. Manifest/cloak creature-card payments disclose the underlying creature
and cost before payment. Morph/disguise payments disclose only their cost.
`CostDisclosures` exposes this public information separately from battlefield
characteristics, which remain face down until the commit.

Action/X enumeration requires `Complete`, not a manual-mana heuristic or clone
probe. `Options` supplies a complete witness for every normal payment decision,
including sequential object/counter selections, a selected mana source's
payment components and its bounded output choices. Source-use choices remain
identity-based, while `ManaActivationKey` binds costs and outputs to each
zero-based occurrence of that ability. Repeated Altar activations can therefore
sacrifice different creatures and choose different colors; each occurrence's
choices see the resources left by its predecessors. Fixed occurrence choices
cannot disappear from a completion or move to another activation.
`Plan.Request` retains the
locked total across resource changes; channel does not recalculate its reducer
after sacrificing a legendary creature. Special-action source reservations
participate in the joint search, so mana payment cannot consume the object
whose special action it is completing. Generic sacrifice mana costs (including
Phyrexian Altar) backtrack through their subjects as part of that same search.

Only selecting an opaque mana activation opens a whole-world checkpoint.
Opacity includes both structural classification and a failed context forecast,
including an otherwise automatic tap source with an optional mana trigger.
Intrinsic library-moving activated abilities remain nonmana, not opaque.
The checkpoint repairs live player/card/permanent/stack references by ID,
restores the ID counter with `Restore`, and preserves synchronous caller scopes.
A game-owned `rand.Source` delegates to the caller's arbitrary `Rand.Uint64`
and journals words only during that checkpoint; restoration replays those
words before continuing the original upstream stream. Normal `Game.Clone`
still forks an independent search RNG. No random-source serialization,
reflection, reseeding, or normal-path world snapshot is used.

Failed reversible opaque payments cancel the announcement/stack object as well
as the mana effects and logs. CR 734.1's irreversible library boundaries retain
the legal mana actions and their surviving identities rather than replaying
them; the failed proposal is canceled. Repeated failures do not withhold an
action: legally distinct retries remain available (CR 734.2), while a
nonprogressing policy terminates under the execution decision budget.

Rules read: official Comprehensive Rules effective **2026-11-06**,
[Wizards source](https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
SHA-256 `524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829`.
Sections: 601.2a–i, 602.2a–b, 605.1a–b/605.3a, 116.2b/f/h/k/116.3,
701.40b/701.58b, 702.37e/702.168d, 702.49a–d, 702.29a–f, and 734.1–2
(the current reversal section is 734, not the historical 733).

Follow the ordered sub-issues of #3718: planner plans and completion queries,
mana-ability schemas, abilities and special actions (proving the procedure
and the opaque fallback), priority spell casting, resolution-time casts,
generic payments, then removal of dead machinery. Behavior is preserved
except where existing tests asserted a rewind on the normal path; those tests
change to the rules-correct outcome, with the rule cited. The absorbed open
issues are labelled `Parked` so production does not fix them piecemeal.
