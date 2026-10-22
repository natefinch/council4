---
name: impl-loop
description: Implement a selected epic, eligible Ready For Dev issues, or authorized continuous card production using authored Go and shared Engine prerequisites. Use when asked to run impl-loop, work through an epic, or keep implementing cards/issues. No automatic PR or merge ceremony.
---

# Implementation Loop

Follow the [repository policy](../../../.github/copilot-instructions.md).
Default to authored Go card behavior,
one main writer, proportional checks and direct commits when authorized.
Compiler expansion is not the work-selection strategy.
For card implementation, route to [card-authoring](../card-authoring/SKILL.md)
and its [canonical guide](../../../docs/card-authoring.md), including the
executable scaffold and behavioral evidence.

## Select work

1. Read the requested epic/issue and live blocking relationships. A selected
   epic takes precedence over general backlog selection; use its earliest
   unblocked child and existing runtime prerequisites.
2. For broadly authorized continuous card production, follow the canonical
   guide's [saved queue and continuation scope](../../../docs/card-authoring.md#saved-production-queue).
   Resume the same owner's claim first. If no reviewed batch is executable,
   review and persist a small new family or implement its proven canonical
   Engine prerequisite; no saved batch is not completion or a request for
   another strategy prompt. The coordinator alone mutates the queue.
3. Otherwise, if no epic was selected, choose an actionable open `Ready For Dev` issue in
   the requested scope. Check blocking issues, active ownership, open work and
   supersession notices before claiming it. Do not create an epic merely to
   run an already defined queue.
4. A reference in a closed epic is not a reservation. An open active assignment
   still needs coordination; do not take over another writer's change.
5. Exclude parked/superseded work and Oracle expansion tasks. Only select legacy
   compiler maintenance when the assigned work explicitly requires it.
6. Mark the issue `In Progress`. Promote the next child to `Ready For Dev` only
   after its prerequisites are complete and its requirements remain actionable.

## Implement and checkpoint

1. Inspect the checkout and safely fast-forward clean main from origin. Never
   reset, stash, rebase or discard unrelated work to begin an issue.
2. Read only the relevant current code, rules and requirements. Reuse shared
   Engine operations; keep costs/targets/declarations inspectable and mutable
   state game-owned. If the authoring interface is not ready, follow its
   canonical prerequisite instead of creating a second one.
3. Implement the bounded change and behavioral scenarios. Run scoped tests
   during iteration; new card support does not require parser/lowering work.
4. Follow the [validation and review policy](../../../.github/copilot-instructions.md#verify-proportionally).
   Do not add per-issue gates or report refreshes.
5. Commit only intended changes directly to main when authorized. No mandatory
   branch, PR, approval or merge. Push normally only when authorized; report
   protection/ownership conflicts rather than force-pushing or bypassing them.
6. Record the commit and meaningful outcome once. Close the issue after its
   acceptance criteria are met, noting explicitly whether work is only local
   or has reached origin. Do not report a local commit as deployed support.

## Missing capabilities and next work

Reuse a concrete blocker or file one concise `To Be Triaged` / `Card Support TODO`
issue if distinct work is proven. Link the affected issue to that blocker.
Do not file an issue for every unexamined card or speculative wording, and do
not maintain duplicate future-work sections across epics.

Keep partial cards unsupported. Leave blocked issues open, remove stale
`Ready For Dev`, explain the missing capability, and proceed to another eligible
item only within the user's assigned scope. Do not clear meaningful active
ownership until a handoff exists.

When the run includes both remaining cards and their required Engine support,
a proven blocker routes to a bounded canonical prerequisite, not an automatic
stop. Check live ownership, retain the acceptance probe, implement and review
the shared invariant, then explicitly reassess dependent exact identities
against the actual runtime and remaining printed behavior. Issue closure alone
does not make a batch ready or a card supported. Keep selected Engine work moving
instead of indefinitely substituting easier batches. Independent ready-card
work needs isolated workspaces and serial integration; it does not authorize
concurrent queue writers.

Refresh live status between items. Complete an epic only when all its required
acceptance work is complete, not when no child can currently be selected.
Stop with an actionable summary when the requested work is done, every remaining
item is outside scope or cannot safely proceed after bounded investigation and
escalation, an authorized budget ends, required sources/access are unavailable,
ownership or integration permissions conflict, or an external decision is
required. A bounded issue/batch assignment does not silently expand into corpus
work. Never treat a stopping condition as completion of blocked work.

## Prompt

> Work through the selected epic's unblocked children. Author Go using shared
> Engine operations; do not expand Oracle recognition to add cards. Preserve
> existing behavior, validate proportionally, inspect the diff and commit directly
> to main when authorized. Keep one writer; no mandatory PRs or new epics. Reuse
> precise blockers and honor the run's scope, budget and push permissions.
> For an authorized continuous-production assignment, replenish reviewed small
> batches, work canonical Engine prerequisites, reassess dependents and resume
> from saved ownership/evidence without another strategy prompt. Tune concrete
> workflow friction while delivering; do not start another model tournament,
> scheduler or worker fleet as a prerequisite.
