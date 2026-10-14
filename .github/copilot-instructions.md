# Working in Council4

Council4 builds a Go Magic library for reproducible Commander playtests.
New card support is **authored Go using shared Engine operations**, not Oracle
compiler expansion. [ADR 0012](../docs/adr/0012-authored-go-card-implementations.md)
records this decision; [epic #3269](https://github.com/natefinch/council4/issues/3269)
is the dependency-ordered migration.

## Choose the next work

- Follow the user's selected issue or epic. Read its current requirements,
  relevant comments and actual blocking dependencies, not every historical plan.
- For an epic, implement the earliest unblocked child. Complete/reuse its
  canonical runtime prerequisites before starting a dependent child. Do not
  create another epic to reserve work already assigned to one.
- `Ready For Dev` means actionable and prerequisites satisfied. Do not promote
  parked, superseded, dependency-blocked or compiler-expansion work by default.
  A closed epic reference is not a permanent reservation.
- Mark an issue `In Progress` only while working on it. Check existing ownership
  before taking over active work; preserve unmerged changes and handoff evidence.
- For a direct user request without an issue, do the requested work; do not
  require an issue-filing conversation as an implementation prerequisite.

## Direct-to-main workflow

The default for an authorized issue/epic implementation is one coherent local
commit per completed change on `main`, without mandatory branches, PRs, approvals
or merge choreography. Respect the current run's commit/push permissions; this
policy does not authorize unsolicited commits or pushes for arbitrary edits.

1. Inspect the checkout, existing edits and HEAD. With a clean `main`, fetch
   `origin` and fast-forward only (`git merge --ff-only origin/main`). Do not
   reset, stash, rebase or switch the user's checkout to enforce this default.
   Stop if divergent history or conflicting work needs a decision.
2. Keep **one writer to main**. Parallel research is fine; independent code
   work needs isolated workspaces and serialized integration. `In Progress`
   labels help coordination but are not an atomic writer lock.
3. Implement a bounded change, run the relevant checks below, and inspect the
   diff. Stage only intended files, not all tracked or untracked modifications.
4. Commit when authorized, recording the issue and required commit trailers.
   Push normally only when authorized; never force-push or bypass protections.
   If a protection requires a PR, report it rather than disabling it.
5. Record completion once, with the commit and meaningful behavior/blockers.
   A local commit is not a pushed or deployed change. Continue only within the
   user's assigned scope. PRs remain available when explicitly requested.

## Author cards, not another compiler

- Read current Oracle text/rulings and the relevant Comprehensive Rules. Keep
  costs, targets, modes, triggers and static/replacement declarations inspectable.
  Use ordinary Go control flow where it is simpler than declarative Instructions.
- Reuse engine-owned operations for mutation, batch semantics, payment,
  replacement order, targeting, priority and events. Card code must not access
  raw mutable Game state or approximate a missing rule.
- Keep programs immutable and per-game references, links, usage and delayed
  state game-owned and clone-safe. Preserve reproducibility and fog of war.
- Preserve existing generated cards, useful Instructions and behavioral tests.
  Both old and authored cards use one Engine; no wholesale rewrite or new DSL.
- The full authoring interface is being built in #3272. Today's
  `CardImplementation.ResolveSpell` is only a narrow spell escape hatch, not
  general support for custom triggers, static abilities or replacements. If
  infrastructure is missing, reuse its migration prerequisite; do not invent
  a parallel implementation or expand the parser to avoid the blocker.
- `cardgen` is retained for legacy correctness/compatibility maintenance and
  explicit regeneration, not the default route to new coverage. Parser-complete
  and compiler-supported counts do not certify gameplay or deployed availability.

## Verify proportionally

- During development, run the smallest tests covering the change.
- For code changes, run `go test ./...` and
  `go run github.com/magefile/mage@v1.15.0 lint` once as the final local gate.
  Add relevant build, race or cross-platform checks when their behavior is
  affected. Do not repeat the same gates solely because a skill also requests them.
- Documentation/comment-only changes need link/instruction checks and
  `git diff --check`, not the full game suite, linter or corpus regeneration.
- Run pinned, same-input corpus comparisons only for relevant legacy compiler
  changes or an explicit corpus refresh. Never run `cardSupport` after every
  issue. Keep existing CI until #3270 changes it; do not claim its PR-only
  support-doc auto-commits have already been removed.
- Test actual card interactions as well as shared operations. Named-card
  scenarios are welcome; structural CardDef assertions test translation, not
  executed Go or rules behavior. Source fingerprints are review signals, not a
  substitute for runtime regression coverage.
- Inspect every diff yourself. Routine changes need no mandatory independent
  review; substantial shared-rule changes warrant focused independent review
  (identity, replacements, cloning, timing, hidden information). Fix substantive
  findings without restarting broad architectural review after every edit.
- Do not hide a failing check, partial card or incomplete game behind a success
  result. Distinguish rules draws from budget exhaustion and unsupported behavior.

## Keep the work small

Prefer existing helpers, type-safe Go and focused edits. Use code intelligence
and ranged reads to navigate. Do not split files to meet an arbitrary line count,
fan out agents per file, build wrappers for hypothetical variation, or refactor
unrelated code during review. The thermo-nuclear review skill is opt-in only.

For a concrete missing capability, briefly search for an existing issue and link
it as the blocker. One shared blocker may cover many cards. File new work only
when its scope and completion condition are known; do not enumerate speculative
neighboring wording or mirror routine notes across epics. Park affected cards
whole and continue other eligible work when the assigned scope allows it.

## Agent startup

> Read the selected epic and first unblocked child. Author Go card behavior using
> shared Engine operations; Oracle recognition is not a prerequisite. Preserve
> existing cards and runtime scenarios. Do not expand parser/lowering to add
> cards, invent a DSL or mutate raw Game state. Reuse a precise blocker when a
> shared capability is missing. Implement, test/lint, inspect, and commit directly
> to main when authorized; push only with permission. Keep one main writer and
> continue only within the assigned work.
