# Repository Work Methodology

These repository-wide rules are always active. Follow them together with any
applicable skill or user instruction. When multiple instructions request the same
review or workflow action, perform it once rather than duplicating it. When a
narrower skill conflicts with this methodology, this methodology takes precedence.

## Route Work Through Issues

Before changing the repository for work that is not already tied to an issue or
epic, ask the user whether an issue exists for the work.

- If the user provides an issue, use it.
- If the user is unsure or says no issue exists, search for an existing issue that
  covers the work. If a likely match exists, show it to the user and ask whether it
  applies.
- If no matching issue exists, ask whether the user wants a new issue created.
  Do not create one without the user's approval.

Read the active issue's body and relevant comments before implementation. When an
issue becomes active, including a child issue of an epic, add the exact
`In Progress` label.

## Start Work From `origin/main`

Fetch `origin` before starting a new task, then create its working branch from the
current `origin/main`, never from a stale local `main` or an unrelated branch.

When starting an epic, use the pull request strategy selected for that epic:

- **Individual pull requests:** create each child issue's branch from the current
  `origin/main`.
- **Stacked pull requests:** create the first branch from `origin/main`, then create
  each subsequent branch from the preceding branch in the stack and target each
  pull request at its correct parent branch.
- **One pull request:** create one epic branch from `origin/main` and implement all
  child issues on that branch.

## Start an Epic

When the user asks to work on an epic:

1. Add the exact `In Progress` label to the epic.
2. Determine whether the user already specified individual pull requests, stacked
   pull requests, or one pull request for all child issues.
3. If no strategy was specified, ask the user to choose among those three options
   before implementation.
4. Add the exact `In Progress` label to each child issue only when work on that
   child starts.

Do not ask again when the user's request already states the pull request strategy.

## Record Decisions

Record important implementation or scope decisions as comments on the active
issue. State the decision and why it was made, including the relevant tradeoff or
rejected alternative when useful. Do not create comments for routine implementation
details that do not represent a meaningful decision.

## Track Future Work Discovered

When implementation reveals concrete work outside the active issue's scope, such
as a bug, optimization, or intentionally deferred work:

1. Search existing issues before creating a new one. Reuse an issue that already
   covers the work.
2. If no issue covers it, create a focused issue and add the exact `To Be Triaged`
   label. Add any applicable specialized labels as well.
3. In the new issue body, link the issue that was active when the work was
   discovered and explain the discovery context.
4. If the active issue is a child of an epic, update the epic body so it contains
   a `## Future Work Discovered` section with one `###` subsection per epic child
   that discovered future work. Add the new issue as a link under the active
   child's subsection.
5. Do not make the newly created issue a child of the active epic or add it to the
   epic's implementation checklist.

Do not file speculative ideas, duplicates, or work that belongs in the active
issue.

## Review Completed Issue Work

Give every completed issue a thorough independent review. A user request, skill,
or other instruction that also requests review does not increase the number of
reviews.

Fix review findings when practical. Minor nitpicks may be fixed without another
review. If the first review finds substantive problems and they are fixed, run at
most one full second review. Do not enter a review-and-fix loop. If significant
problems remain after the second review, bring them to the user's attention before
proceeding.

## Work Efficiently in This Codebase

These practices keep work fast and reviewable. They matter because the slowest
part of most changes is processing a large working context repeatedly, not running
the build or tests.

### Navigate with code intelligence, not whole-file reads

Prefer code-intelligence tools over reading entire files when locating or
understanding code. If a Go language server (gopls) is available, use it:

- Use document-symbol to learn a file's structure and pick exact edit locations
  instead of reading the whole file.
- Use go-to-definition, find-references, and incoming-calls to trace behavior
  across the parser, compiler, and lowering layers instead of grepping and reading
  several large files.
- Use hover for a symbol's type, signature, and doc instead of opening its file.
- Reserve full or ranged file reads for the specific span you are about to edit.

If no Go language server is configured, you can set one up (for example, gopls via
the `lsp-setup` skill). Until then, prefer ranged reads and symbol search over
whole-file reads.

### Keep Go files focused

Keep Go source and test files focused, generally under about 1,000 lines. The
repository has no Go file over 1,000 lines; do not reintroduce one. When a file
grows past that, split it into focused files within the same package by moving
whole declarations verbatim (no behavior change). When adding substantial new code,
start a new focused file rather than appending to an already-large one.

### Decompose large independent work into narrowly scoped sub-agents

For mechanical or naturally independent multi-file work, delegate one small,
self-contained scope per sub-agent (for example, one file or one cohesive concern)
running in isolated worktrees, rather than one agent carrying the whole task. Small
scope means small context per turn, which is dramatically faster. Give each
sub-agent a precise file-and-symbol inventory up front so it does not spend turns
rediscovering the surface.

### Validate proportionally; reserve the corpus check for Oracle changes

Match validation to the change. For most changes, `go build ./...`,
`go vet ./...`, the affected package's `go test`, and `mage lint` are enough; a
purely mechanical move also wants `gofmt`/`git diff --check`.

For fast iteration, prefer `go test -short ./...`: it skips the slow full-game
simulation tests in `cmd/council4` (and the `cardgen` source round-trips), which
otherwise re-run on every change to `mtg/game`/`mtg/rules`. Run the full
`go test ./...` (which includes those simulations) once as the final gate before
committing, alongside `go vet ./...` and `mage lint`.

For changes to the Oracle compiler or cardgen lowering, the strongest
behavior-preserving check is the corpus comparison: compile the full card corpus
with `cardgen/oracle/cmd/compilecards` on the merge base and on the branch, then
`diff -rq` the two generated trees and compare the JSON reports. A byte-identical
result proves zero behavioral change across every supported card, and the whole
check takes only a few seconds. Use it for refactors that must not change output;
for intentional behavior changes, confirm the diff is exactly the intended cards
and explain every difference.
