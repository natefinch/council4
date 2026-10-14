---
name: impl-loop
description: Implement a selected epic's unblocked children or eligible Ready For Dev issues using authored Go and a direct-to-main workflow. Use when asked to run impl-loop, work through an epic, or keep implementing issues. No automatic PR or merge ceremony.
---

# Implementation Loop

Follow `.github/copilot-instructions.md`. Default to authored Go card behavior,
one main writer, proportional checks and direct commits when authorized.
Compiler expansion is not the work-selection strategy.
For card implementation, route to [card-authoring](../card-authoring/SKILL.md)
and its [canonical guide](../../../docs/card-authoring.md), including the
executable scaffold and bounded rules-review/promotion gate.

## Select work

1. Read the requested epic/issue and live blocking relationships. A selected
   epic takes precedence over general backlog selection; use its earliest
   unblocked child and existing runtime prerequisites.
2. If no epic was selected, choose an actionable open `Ready For Dev` issue in
   the requested scope. Check blocking issues, active ownership, open work and
   supersession notices before claiming it. Do not create an epic merely to
   run an already defined queue.
3. A reference in a closed epic is not a reservation. An open active assignment
   still needs coordination; do not take over another writer's change.
4. Exclude parked/superseded work and Oracle expansion tasks. Only select legacy
   compiler maintenance when the assigned work explicitly requires it.
5. Mark the issue `In Progress`. Promote the next child to `Ready For Dev` only
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
4. Use the repository validation matrix: final test suite and linter once for
   code, documentation checks for docs-only work, and additional affected
   platform/race checks when justified. No unconditional `cardSupport` or
   full-corpus regeneration. Inspect the diff yourself; focused independent
   review is for substantial shared-rule changes, not a gate on every issue.
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

Refresh live status between items. Complete an epic only when all its required
acceptance work is complete, not when no child can currently be selected.
Stop with an actionable summary when the requested work is done, every remaining
item is blocked/parked, an authorized budget ends, or an external decision is
required. Never treat that stopping condition as completion of blocked work.

## Prompt

> Work through the selected epic's unblocked children. Author Go using shared
> Engine operations; do not expand Oracle recognition to add cards. Preserve
> existing behavior, validate proportionally, inspect the diff and commit directly
> to main when authorized. Keep one writer; no mandatory PRs or new epics. Reuse
> precise blockers and honor the run's scope, budget and push permissions.
