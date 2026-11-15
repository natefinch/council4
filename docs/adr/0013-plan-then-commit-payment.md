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
An exact payment from present resources is checked before enumerating mana
abilities at all. When contextual effects rule out the ordinary-only fast path, a direct
automatic-source allocation can still supply a candidate history. The search
revalidates every cost, production and effect through the canonical resource
forecast before accepting that history. A rejected candidate falls through to
the complete search; it never proves unpayability. Exact allocation receipts
are revalidated only at their fixed activation count, not after enumerating
shorter histories that cannot match the receipt. Iterative deepening also stops
when no branch reaches its depth: the reachable resource graph is then fully
exhausted, independent of the conservative maximum activation count.
This avoids subset enumeration
for ordinary land payments under unrelated continuous effects while preserving
strategic sources, occurrence choices and opaque-context rejection.

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
their existing entry/replacement decisions. Resolution-time casting remains
pending #3723; priority casting uses the procedure described below.

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
payment components and its bounded output choices. `Partial.ManaAbilities`
retains identity-level use/avoid preferences. The procedure uses
`Partial.ManaOccurrences` with `ManaActivation` on `StepManaAbility` steps to choose or
decline one zero-based occurrence; declining the next occurrence does not
cancel earlier selected uses. `ManaActivationKey` also binds costs and outputs
to each occurrence of that ability. Repeated Altar activations can therefore
sacrifice different creatures and choose different colors; each occurrence's
choices see the resources left by its predecessors. Fixed occurrence choices
cannot disappear from a completion or move to another activation.
Each next reusable activation is offered when it retains a completion, even
when an automatic land could instead supply the remaining mana. The Baleful
Strix scenario chooses Altar twice, sacrificing two Bears for black then blue
while keeping Island untapped; the Crystal cycling scenarios still cover
required repeated activations and their resource-dependent costs.
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

### Priority spell casting (#3722)

Hand, command-zone, graveyard, exile and library-top casts, permission-specific
alternative costs (including foretell, plot, flashback and escape), mutate,
prepared copies and face-down morph/disguise casts use `Engine.costedAction`.
The descriptor captures source permissions before the transition. The shared
procedure moves the source and pushes its `Casting` object at CR 601.2a, before
cost/Phyrexian announcements or any target callback. Timing is checked at the
start, not against the stack after this transition. That same object becomes
cast at 601.2i; permission usage and cast events are committed only then.
Face-down permission filters and cost modifiers use the face-down spell's
characteristics, including when casting a morph land under a creature-spell
permission. `Casting` objects are excluded from stack-target candidates
(CR 115.5). Prepared sources retain their designation until commit.

Availability uses a read-only resource view with the source removed and the
proposed object on the stack. Cost options and binding Phyrexian intentions
are admitted only if some exact target completion has a `Complete` witness.
An effect allowing life for a colored mana cost, such as K'rrik, remains a
separate payment choice, not a fallback that reopens the Phyrexian intention.
Plans retain those original announcements, rather than inferring them from
the final symbol receipt after color-to-life effects.

Numeric actions have a bounded X/kicker policy domain, but exact supplied
values remain valid beyond that domain. Combinatorial numeric target domains
use a witnessed provisional set, not the first 64 target sets. During casting,
the target count is fixed before target objects are chosen (601.2c). Every next
target, including an opponent's target decision, extends the fixed prefix to
a payable whole set. Generation constrains that prefix directly rather than
enumerating unrelated combinations and filtering afterward. Splice selection
also requires a joint host/splice-target witness; all selected cards are publicly
revealed together before any host or spliced target is announced
(601.2b, 702.47a–d). Spliced text uses the host's characteristics.

Payment uses the locked request, occurrence-specific mana-activation decisions
and the shared `Apply` operation. Source sacrifices, life and counters are
reserved jointly, including valid mana effects that enable later nonmana costs.
Spell mana-spend riders consume `ApplyResult.ManaBefore`, which includes planned
production. Normal choices cannot decline a required source into an unpaid
cast. Only a selected opaque activation opens the shared whole-world/RNG/log
checkpoint; canceling it also cancels the source transition or prepared copy
(734.1–2). Priority no longer withholds proposals or exhausted actions.
Resolution-time proposal/probe/reversal machinery remains for #3723; final
dead-helper removal remains #3725.

Folded issue interpretation under the **2026-11-06** rules:

- #3702: registered Dismember with `{C}` and four life observes the topmost
  same object during both Phyrexian announcements and the Grizzly Bears target;
  graveyard/exile/library and mutate paths have equivalent observations.
- #3678: the payable two-Bears pair after eleven opposing Elderwood Scions
  is found on hand and prepared-copy routes without a target sample or empty
  cost proposal.
- #3705: registered Desperate Ritual and Glacial Ray exercise multiple splice
  reveals visible to an opponent before the host's and spliced text's targets.
- #3570: priority Phyrexian symbols, including kicker, are binding before
  targets. The free/resolution-time portion remains #3723.
- #3632/#3633: planned valid mana sources enable X and counter/life additional
  costs. Millikin and Deranged Assistant are **nonmana activated abilities**
  under 605.1a, not payment-time sources. Underworld Breach/Sol Ring escape is
  rejected with Millikin as its proposed payment source; activating and resolving
  Millikin separately makes the resulting escape legal.
- #3628: the old library-moving “Milling Prism” mana-ability example is not
  legal under 605.1a. Existing standalone activation/resolution and enclosing
  library-top window behavior is retained; no illegal nested activation is
  enabled or claimed fixed by that example.

Additional sections read in the official rules cited above: 601.3,
702.47a–e, 702.140a–f, 708.4–6, 702.37a–g, 401.5 and 734.1–2.

Follow the ordered sub-issues of #3718: planner plans and completion queries,
mana-ability schemas, abilities and special actions (proving the procedure
and the opaque fallback), priority spell casting, resolution-time casts,
generic payments, then removal of dead machinery. Behavior is preserved
except where existing tests asserted a rewind on the normal path; those tests
change to the rules-correct outcome, with the rule cited. The absorbed open
issues are labelled `Parked` so production does not fix them piecemeal.
