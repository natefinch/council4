---
name: ready-for-dev-loop
description: Refine actionable issues in the selected epic or requested scope, respecting dependencies and the authored-Go direction. Use when asked to triage issues, prepare work for implementation, or run the ready-for-dev loop.
---

# Ready For Dev Loop

Prepare work; do not implement it. Follow `.github/copilot-instructions.md`.
The goal is an executable queue, not making every open issue `Ready For Dev`.

## Scope and eligibility

1. Start with the user-selected epic, issue list or workstream. For the card
   migration use #3269's actual dependencies, not parser/lowering rankings.
   Broader backlog triage must stay within the user's request.
2. Read current issue bodies, relevant comments, active assignments and blocking
   relationships. Exclude epics, parked/superseded work and compiler-only
   expansion from ordinary readiness promotion.
3. Classify card work as authored implementation, Engine capability, explicit
   legacy compiler maintenance or diagnostics-only work. Preserve useful runtime
   bugs and research even when their old compiler approach is superseded.
4. Refine a dependent issue if helpful, but do not label it ready until every
   real prerequisite is satisfied. Native blocking links and the issue's
   requirements both matter. Remove a stale readiness label with an explanation.
5. Do not take over active work, close its PR or discard its evidence. Coordinate
   a checkpoint/handoff before changing an active assignment's execution path.

## Make one issue actionable

Inspect the owning code, nearest shared operation/tests and relevant rules.
Preserve the original requirements. Update one `## Implementation guidance`
section with the concrete limitation, owning files, bounded approach,
acceptance scenarios, prerequisites, appropriate validation and non-goals.
Do not write a second epic or a speculative universal abstraction.

For cards, Oracle wording/rulings inform authored Go and runtime scenarios.
Recognition is not a prerequisite. Do not require compiler delta, rendered-source
inspection or full corpus regeneration for an authored card. A capability
missing from the Engine gets one canonical blocker, not many wording issues.

Apply `Ready For Dev` only after the handoff is actionable and prerequisites
are complete. Mark truly deferred work `Parked`; keep legacy compiler
maintenance identifiable with `Legacy Compiler`. Clear parked status only when
the user has selected/reactivated the work and its blockers are resolved.

Duplicates or already completed work can be closed with evidence; superseded
approaches should retain history and point to their replacement. Do not
bulk-close real runtime work because the compiler-first plan changed.

Refresh live state before editing an issue so concurrent requirements are not
overwritten. Record only material decisions/blockers, not routine triage chatter.

## Finish

Stop when eligible requested work has a usable handoff, or remaining work is
blocked/parked and accurately classified. Report the first unblocked next issue
and concrete blockers. Do not keep promoting the backlog to satisfy a
"no unready open issues" metric.
