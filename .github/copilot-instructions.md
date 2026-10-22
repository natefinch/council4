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

## Integration workflow

For authorized implementation, use one coherent change at a time and preserve
the current run's commit/push permissions. The user-directed production workflow
uses PRs, with the latest main merged into the task branch and local tests run
before CI-driven merging. Do not treat authorization for code edits as blanket
publication permission for unrelated work.

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
   Publish task branches and open PRs within scope; merge only after local
   validation, applicable review, and all CI succeeds on the exact current head.
   Never force-push, use admin merge, or bypass/disable protections.
5. Record completion once, with the commit and meaningful behavior/blockers.
   A local commit is not a pushed or deployed change. Continue only within the
   user's assigned scope. PRs remain available when explicitly requested.

Successful cards and Engine prerequisites are delivered to **remote `main`**
through PRs, not left indefinitely as local commits. The
[Dynamic Workflow](../docs/dynamic-card-production.md) serializes main integration,
local validation, independent review where required, CI/review repairs and
head-matched squash merging. It confirms the PR merge commit on remote main.
`Race Tests` remains required on GitHub and is run automatically for PRs; agents
run locally, not in Actions. Preflight/reporting do not authorize implementation.

## Author cards, not another compiler

For a new card request or production batch, use the repo-owned
[card-authoring skill](../.agents/skills/card-authoring/SKILL.md) and
[canonical guide/scaffold](../docs/card-authoring.md). They are the entry route;
do not create another authoring checklist.

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
- Use the versioned `game.Mode.Program` and `mtg/cardprogram` context/catalog for
  Go resolving bodies. The old `CardImplementation.ResolveSpell` is retained for
  compatibility, not a slot for arbitrary static/replacement rules. Later
  migration slices supply deeper operations and policy seams. If one is missing,
  reuse its prerequisite rather than inventing another runtime or parser route.
- `cardgen` is retained for focused legacy correctness/compatibility maintenance,
  not new coverage or bulk regeneration. Keep existing generated cards and tests.
  Compiler acceptance and loadability do not certify gameplay.

## Verify proportionally

This is the authoritative validation policy; skills and guides link here.

- Ordinary code work: run affected-package behavioral tests. Shared rules also
  need affected interaction/rules scenarios. Run `go test ./...` locally only for
  broad changes or an explicit request; CI runs it once for code changes.
- Documentation/comment-only work: check links/instructions and
  `git diff --check`; no local game tests. CI skips code checks for Markdown-only
  changes; other changes run the full suite.
- Lint and lint autofixes require an explicit user request. The command is
  `go run github.com/magefile/mage@v1.15.0 lint`; wait for any cardgen round-trip
  tests to finish first. Standalone lint CI is manual maintenance, not a
  default gate. Build and platform checks are needed only when affected or
  explicitly requested. By explicit user request, PR CI has `Tests` and required
  `Race Tests` (`go test -race ./...`); race is not an extra local per-card gate.
  Manual `workflow_dispatch` choices `check=tests|lint|race` remain available.
- Never recompile the Oracle corpus. Source-fingerprint refresh and registry
  fingerprint-baseline workflows are removed. Do not run global
  `supportInventory` as an implementation gate; it is explicit reporting
  maintenance using existing evidence, not pre/post card work.
- Preserve exact card identities, program IDs/versions/dependencies and genuine
  scenario evidence. Test actual card interactions and shared operations;
  structural assertions, generated source and hashes are not behavior evidence.
- Self-inspect every diff. Focused independent review is only for significant
  shared identity, timing, replacement, cloning or hidden-information changes,
  not every nontrivial batch. Fix substantive findings without restarting broad
  architectural review after every edit.
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
