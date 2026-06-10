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
- Create one fresh branch from current `origin/main` per implementation issue.
- Create one pull request per implementation issue.
- Independently review every completed implementation and fix all substantive
  findings before opening or merging its pull request.
- Merge each pull request before starting the next issue.
- Never silently omit known card-support work. Reuse an existing issue or create
  a concise issue labeled `Card Support TODO`.
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
  --body $'## Goal\n<short outcome>\n\n## Issues\n- [ ] #123\n- [ ] #124\n\n## Coordination\nReserved for the active impl-loop. Each issue ships in its own branch and pull request.'
```

An epic is an issue carrying the exact `Epic` label; a title alone does not make
an issue an epic. The epic is the coordination lock. Include every selected issue
by number so other agents can detect the reservation. Do not add issues already
present in another epic.

### 4. Implement One Issue

For each epic child, in impact order:

1. Refresh from the remote default branch:

   ```bash
   git fetch origin
   git switch -c copilot/<short-issue-slug> origin/main
   ```

2. Read the issue, relevant package documentation, architecture decisions, and
   existing tests.
3. Establish a baseline before editing, including corpus support when the issue
   affects card generation.
4. Implement the complete issue with strict, fail-closed behavior.
5. Add focused tests for success, rejection, runtime semantics, and regressions.
6. Measure corpus impact where applicable.
7. Inspect every newly supported card. Do not accept generated output solely
   because counts increased.
8. Update package documentation, supported-card counts, and generated support
   lists affected by the change.
9. Run the repository's established full validation commands.

Do not mix unrelated cleanup into the branch.

### 5. Preserve Deferred Card Work

When implementation uncovers intentionally excluded wording, mechanics, runtime
behavior, or compiler support:

1. Search for an existing open issue with the same specific scope.
2. If none exists, create a concise issue labeled `Card Support TODO`:

   ```bash
   gh issue create \
     --title "<actionable missing support>" \
     --label "Card Support TODO" \
     --body $'## Deferred work\n<known gap and why it is deferred>\n\n## Done when\n<observable completion condition>'
   ```

3. Resume the active implementation immediately.

Do not file speculative ideas, duplicate issues, or vague umbrella tasks.

### 6. Review Until Clear

Run an independent code review focused on correctness, regressions, validation
holes, runtime semantics, and fail-closed behavior.

For every substantive finding:

1. Fix the root cause.
2. Add or strengthen regression coverage.
3. Re-run relevant tests and full validation.
4. Request another focused review.

Repeat until the reviewer reports no significant issues. Do not open or merge a
pull request with unresolved substantive findings.

### 7. Pull Request and Merge

1. Commit with the repository's required commit-message conventions.
2. Push the branch.
3. Open a pull request that summarizes behavior and measured impact and closes
   the child issue.
4. Confirm the pull request is mergeable and required checks pass.
5. Merge it using the repository's preferred merge strategy.
6. Confirm the issue closed.
7. Update the epic checklist with the merged pull request.
8. Return to `origin/main` and start the next child from the newly merged commit.

Never stack the next issue on an unmerged implementation branch.

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
- **Merge conflict:** Merge current `origin/main` into the issue branch, resolve
  carefully, revalidate, and repeat review if conflict resolution changed code.
- **Failed checks:** Diagnose and fix them. Never merge by bypassing checks.
- **Missing labels:** Create `Epic` or `Card Support TODO` only when needed.
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
- Combine multiple implementation issues into one branch or pull request.
- Silently leave discovered card blockers untracked.
