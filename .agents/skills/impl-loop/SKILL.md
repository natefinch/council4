---
name: impl-loop
description: Run an autonomous GitHub issue implementation loop. Use when the user says "run impl-loop", "go on autopilot", "work through Ready For Dev issues", "keep implementing issues", or asks to batch high-impact issues into an epic and implement, review, merge, and repeat until no eligible issues remain.
---

# Implementation Loop

Autonomously select high-impact, development-ready work, reserve it in an epic,
and carry each issue through implementation, review, pull request, and merge.
Continue until no eligible issues remain or a genuine external blocker requires
user input.

## Non-Negotiable Rules

- Work only on open issues labeled exactly `Ready For Dev`.
- Never select an issue already referenced by an existing issue labeled `Epic`.
- Reserve a batch in a newly created issue labeled exactly `Epic` before changing
  code.
- Mark the epic `In Progress`, then obtain the user's pull request strategy as
  required by the repository work methodology.
- Follow the selected individual, stacked, or single pull request strategy. Start
  the first working branch from the current `origin/main`.
- Mark each child issue `In Progress` when its implementation starts.
- Independently review every completed issue once. After substantive fixes, run at
  most one full second review and stop for user input if significant problems
  remain.
- Never silently omit known card-support work. Reuse an existing issue or create
  a concise issue labeled `To Be Triaged` and `Card Support TODO`, then record it
  in the epic as required by the repository work methodology.
- Run `go run github.com/magefile/mage@v1.15.0 cardSupport` at the end of every
  implementation issue so card-support measurement and documentation stay current.
- Keep running without asking for routine confirmation.

## Process

### 1. Refresh and Inventory

1. Verify `gh` authentication and repository access.
2. Fetch `origin`.
3. List open `Ready For Dev` issues:

   ```bash
   gh issue list \
     --state open \
     --label "Ready For Dev" \
     --limit 200 \
     --json number,title,body,labels,url
   ```

4. List all open and closed epic issues:

   ```bash
   gh issue list \
     --state all \
     --label "Epic" \
     --limit 200 \
     --json number,title,body,state,url
   ```

5. Exclude every candidate referenced by an epic body. Treat checklist entries,
   `#123` references, and explicit child-issue lists as reservations regardless
   of whether the epic is open or closed. When uncertain, inspect the candidate's
   timeline and comments before selecting it.

Do not reinterpret an issue without `Ready For Dev` as ready merely because it
looks easy or valuable.

### 2. Choose a High-Impact Batch

Choose up to five unreserved candidates that:

- unlock the most cards or remove a broadly shared compiler/runtime blocker;
- form a coherent implementation theme;
- have limited overlap in files and architecture;
- can each ship independently;
- are unlikely to interfere with work reserved by existing epics.

Use concrete evidence when available: corpus diagnostic counts, issue descriptions,
dependencies, supported-card deltas, or affected wording families. Prefer measured
impact over intuition.

If no eligible candidates remain, stop successfully and report that the loop is
complete.

### 3. Reserve the Batch in an Epic

Before implementation, ensure the `Epic` label exists, then create one epic issue:

```bash
gh issue create \
  --title "Epic: <cohesive batch outcome>" \
  --label "Epic" \
  --body $'## Goal\n<short outcome>\n\n## Issues\n- [ ] #123\n- [ ] #124\n\n## Coordination\nReserved for the active impl-loop.\n\n## Future Work Discovered'
```

An epic is an issue carrying the exact `Epic` label; a title alone does not make
an issue an epic. The epic is the coordination lock. Include every selected issue
by number so other agents can detect the reservation. Do not add issues already
present in another epic.

Add the exact `In Progress` label to the epic. If the user did not already select
individual pull requests, stacked pull requests, or one pull request for the epic,
ask them to choose before implementation. Record the selected strategy in the
epic's Coordination section.

### 4. Implement One Issue

For each epic child, in impact order:

1. Add the exact `In Progress` label to the child issue.
2. Fetch `origin`, then follow the epic's selected branch strategy:

   ```bash
   git fetch origin
   ```

   - For individual pull requests, create the child branch from `origin/main`.
   - For stacked pull requests, create the first branch from `origin/main` and each
     later branch from its preceding stack branch.
   - For one pull request, create the epic branch from `origin/main` before the
     first child, then continue on it for every later child.
3. Read the issue, relevant package documentation, architecture decisions, and
   existing tests.
4. Establish a baseline before editing, including corpus support when the issue
   affects card generation.
5. Implement the complete issue with strict, fail-closed behavior.
6. Add focused tests for success, rejection, runtime semantics, and regressions.
7. At the end of the implementation, run the repository support workflow instead
   of invoking cardgen or corpus compilation manually:

   ```bash
   go run github.com/magefile/mage@v1.15.0 cardSupport
   ```

8. Measure corpus impact from the resulting `supported.md`, `unsupported.md`,
   `unsupported-reasons.md`, README summary, and support-documentation diff.
9. Inspect every newly supported card in `.cardwork/card-support-generated`. Do
   not accept generated output solely because counts increased.
10. Update any additional package documentation affected by the change.
11. Run the repository's established full validation commands.

Do not mix unrelated cleanup into the branch.

For every pull request strategy, complete and review one child's logical change
before starting the next child. For individual pull requests, open and merge the
current child's pull request after its review. For stacked pull requests, open the
current child's pull request after its review but defer merging until the stack is
complete. For one pull request, keep each child's reviewed change on the epic
branch and open the combined pull request only after every child is complete and
reviewed.

### 5. Preserve Deferred Card Work

When implementation uncovers intentionally excluded wording, mechanics, runtime
behavior, or compiler support:

1. Follow the repository work methodology's `Track Future Work Discovered`
   process, including duplicate search, active-issue provenance, and epic update.
2. Add both `To Be Triaged` and `Card Support TODO` to a newly created issue.
3. Resume the active implementation immediately.

Do not file speculative ideas, duplicate issues, or vague umbrella tasks.

### 6. Review Completed Work

Run an independent code review focused on correctness, regressions, validation
holes, runtime semantics, and fail-closed behavior.

For every substantive finding:

1. Fix the root cause.
2. Add or strengthen regression coverage.
3. Re-run relevant tests and full validation.

Minor nitpicks may be fixed without further review. After substantive fixes, run
at most one full second review. Do not enter a review-and-fix loop. If significant
problems remain after the second review, bring them to the user's attention before
proceeding.

### 7. Pull Request and Merge

1. Commit and push each reviewed child's change according to the selected strategy.
2. Open and merge pull requests according to that strategy:
   - For individual pull requests, open and merge the current child's pull request
     before starting the next child.
   - For stacked pull requests, open the current child's pull request against its
     parent branch, build the rest of the stack, then merge the completed stack
     from earliest to latest while retargeting later pull requests as needed.
   - For one pull request, keep the reviewed child commits on the epic branch,
     then open and merge one combined pull request after all children are complete.
3. Summarize behavior and measured impact in every pull request and ensure the
   appropriate child issues will close when their work merges to the default
   branch.
4. Confirm every pull request is mergeable and required checks pass before merge.
5. Confirm the completed child issues closed.
6. Update the epic checklist with the merged pull request links.

### 8. Complete the Epic and Repeat

After all children merge:

1. Confirm every checklist item is closed.
2. Add a final epic comment with pull request links and cumulative impact.
3. Close the epic.
4. Re-run inventory and reserve the next eligible high-impact batch.

Continue until there are no open, unreserved issues labeled `Ready For Dev`.

## Handling Changed or Blocked Work

- **Issue already implemented:** Verify the behavior and closing pull request,
  comment with evidence, close the duplicate if appropriate, and update the epic.
- **Issue loses `Ready For Dev`:** Stop work on it, document the reason in the
  epic, and do not implement it.
- **Issue becomes externally blocked:** Record the blocker in the issue and epic,
  move to another reserved child, and ask the user only if no eligible progress
  remains.
- **Merge conflict:** Resolve against the pull request's intended base branch,
  revalidate, and stay within the repository's two-review maximum.
- **Failed checks:** Diagnose and fix them. Never merge by bypassing checks.
- **Missing labels:** Create `Epic`, `In Progress`, `To Be Triaged`, or
  `Card Support TODO` only when needed.
- **Unrelated dirty worktree:** Preserve it. Do not overwrite or revert user work.

## Output

Keep progress updates concise and milestone-oriented:

- epic created and child issues reserved;
- issue implemented with measured impact;
- review findings and resolution;
- pull request merged;
- epic completed or external blocker encountered.

Do not pause for routine approval between these milestones.

## Boundaries

**Will:**

- Operate autonomously across multiple issues and pull requests.
- Coordinate ownership through explicit epic references.
- Prioritize measured card-support impact.
- File focused deferred card-support issues.
- Stop only when no eligible work remains or external input is truly required.

**Will not:**

- Work on issues without `Ready For Dev`.
- Work on issues already reserved by any epic.
- Merge unreviewed, failing, or conflicted pull requests.
- Combine or stack implementation issues unless the user selected that strategy
  for the active epic.
- Silently leave discovered card blockers untracked.
