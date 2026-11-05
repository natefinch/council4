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
   logs, and the player keeps priority (CR 733.1–2). No other path rewinds.

## Why this is enough

Under the rules effective 2026-09-25, abilities whose cost or effect moves a
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

## Migration

Follow the ordered sub-issues of #3718: planner plans and completion queries,
mana-ability schemas, abilities and special actions (proving the procedure
and the opaque fallback), priority spell casting, resolution-time casts,
generic payments, then removal of dead machinery. Behavior is preserved
except where existing tests asserted a rewind on the normal path; those tests
change to the rules-correct outcome, with the rule cited. The absorbed open
issues are labelled `Parked` so production does not fix them piecemeal.
