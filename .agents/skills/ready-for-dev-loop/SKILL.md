---
name: ready-for-dev-loop
description: Autonomously refine every open implementation issue that lacks the "Ready For Dev" label. Use when the user says "run the ready for dev loop", "triage all open issues", "prepare issues for implementation", "investigate unready issues", or asks to add implementation guidance and mark the backlog Ready For Dev.
---

# Ready For Dev Loop

Investigate open issues, turn each one into a self-contained implementation
handoff, and apply the exact `Ready For Dev` label. Continue until no eligible
open issue remains unrefined.

This skill prepares work; it does not implement the issues.

## Completion Condition

The loop is complete when every open implementation issue has the exact
`Ready For Dev` label.

Issues labeled `Epic` are coordination containers, not implementation issues.
Do not add `Ready For Dev` to an epic, and exclude epics from the completion
query. If an epic contains implementation work directly rather than child
issues, create or identify the appropriate child issues instead.

## Non-Negotiable Rules

- Review every open, non-epic issue lacking `Ready For Dev`.
- Investigate the repository before proposing an implementation.
- Preserve the issue's original requirements and context.
- Add concrete, repository-specific implementation guidance to the issue body.
- Apply `Ready For Dev` only after the issue is actionable by another agent.
- Never use generic advice such as "update the code and add tests."
- Do not implement code while running this skill.
- Keep running without asking for routine confirmation.

## Process

### 1. Refresh and Inventory

Verify GitHub CLI access, fetch the remote, and list open issues:

```bash
git fetch origin
gh issue list \
  --state open \
  --limit 500 \
  --json number,title,body,labels,url
```

An issue is eligible when:

- it is open;
- it does not have the exact `Ready For Dev` label; and
- it does not have the exact `Epic` label.

Process a bounded batch, then refresh the inventory. This avoids acting on stale
state while other agents edit or close issues.

### 2. Prioritize the Batch

Refine issues in this order:

1. Issues blocking active epics or other Ready For Dev work.
2. High-impact card-support and runtime issues.
3. Issues with clear scope but missing implementation details.
4. Ambiguous or broad issues that need decomposition.

Use batches of up to five. Parallelize independent repository investigation when
possible, but update each issue separately.

### 3. Investigate One Issue

Read the full issue body, comments, linked pull requests, and related issues.
Then inspect:

- the packages and files that own the behavior;
- package `README.md` files and architecture decision records;
- existing types, helpers, validation, and runtime paths;
- tests covering adjacent behavior;
- recent commits or pull requests that changed the same area;
- corpus reports or diagnostic counts when card support is involved.

Determine:

- the root limitation, not merely the observed symptom;
- the canonical layer where the change belongs;
- reusable helpers or abstractions;
- expected runtime and fail-closed behavior;
- specific tests and validation commands;
- likely edge cases and explicit non-goals;
- dependencies on other open issues.

Do enough investigation that another agent can begin implementation without
repeating basic discovery. Do not design speculative architecture unsupported by
the repository.

### 4. Resolve Bad Issue Shapes

Before labeling:

- **Duplicate:** Comment with the canonical issue and close the duplicate.
- **Already implemented:** Verify the behavior and merged change, comment with
  evidence, and close the issue.
- **Too broad:** Create focused child issues with clear completion conditions,
  update the parent to reference them, and use the parent as an epic if it is
  purely coordination work.
- **Mixed independent changes:** Split them into separate issues.
- **Invalid or obsolete:** Explain why and close it.

Closed issues do not need `Ready For Dev`. Do not label a coordination-only epic
as implementation-ready.

### 5. Add an Implementation Handoff

Preserve the original body and append or replace one managed section:

```markdown
## Implementation guidance

### Current limitation
<repository-specific explanation of what currently prevents completion>

### Suggested approach
1. <concrete change in the owning package/file>
2. <wiring or runtime behavior>
3. <validation and fail-closed behavior>

### Relevant code
- `path/file.go`: <relevant type, function, or responsibility>
- `path/file_test.go`: <adjacent coverage to extend>

### Acceptance criteria
- [ ] <observable behavior>
- [ ] <important edge case or rejection behavior>
- [ ] <documentation or generated output is updated>

### Validation
- `<existing repository test or validation command>`

### Non-goals
- <explicitly deferred adjacent work, when useful>
```

Use the exact heading `## Implementation guidance` so future runs can locate and
refresh the section without duplicating it. Keep existing user-written sections
intact unless they are demonstrably obsolete; explain material corrections in an
issue comment.

Write the updated body through a temporary file outside the repository:

```bash
gh issue edit <number> --body-file <temporary-body-file>
```

Remove temporary files after use.

### 6. Quality Check and Label

Before applying the label, confirm the issue:

- states an observable completion condition;
- identifies the owning code and likely change points;
- explains relevant runtime or compiler semantics;
- names tests that should be added or changed;
- identifies meaningful edge cases and fail-closed behavior;
- records known dependencies and non-goals;
- is small enough for one branch and one pull request.

Then apply the exact label:

```bash
gh issue edit <number> --add-label "Ready For Dev"
```

If the label does not exist, create it once:

```bash
gh label create "Ready For Dev" \
  --description "Reviewed and refined issues ready for implementation" \
  --color "1BDD5D"
```

Never apply the label before the body update succeeds.

### 7. Refresh and Continue

After each batch:

1. Query open issues again.
2. Confirm edited issues now carry `Ready For Dev`.
3. Detect newly opened or concurrently changed issues.
4. Process the next eligible batch.

Continue until the completion query returns no open, non-epic issue without
`Ready For Dev`.

## Card-Support Issues

For card compiler or runtime support, the handoff should also include:

- representative Oracle wording or card examples already present in the issue;
- current diagnostic family and corpus count when readily measurable;
- which compiler stage owns recognition, lowering, rendering, or runtime support;
- a requirement to measure corpus delta and inspect every newly supported card;
- strict rejection expectations for unsupported variants;
- documentation files and supported-card counts that must be regenerated.

Do not create additional speculative card-support issues merely while refining an
existing issue. Create a related issue only when investigation proves distinct
work that cannot fit the current issue.

## Error Handling

- **GitHub authentication failure:** Stop and report the exact failure.
- **Issue changed concurrently:** Re-fetch it, merge the new context into the
  handoff, and retry. Never overwrite newer requirements.
- **Repository evidence conflicts with the issue:** Document the discrepancy and
  correct the issue body before labeling.
- **External product decision required:** Add the known technical investigation
  and a focused decision question. Do not mark it ready until the decision is
  resolved; continue refining other issues.
- **No implementation-sized scope exists:** Split or close the issue rather than
  falsely labeling it ready.

## Progress Updates

Report only meaningful milestones:

- batch selected;
- issue refined and labeled;
- issue split, deduplicated, or closed;
- blocker preventing the final completion condition;
- all eligible issues are Ready For Dev.

Do not pause for routine approval between issues or batches.

## Boundaries

**Will:**

- Investigate code, tests, documentation, history, and related issues.
- Produce implementation-ready GitHub issue bodies.
- Split, deduplicate, or close malformed issues when evidence supports it.
- Apply the exact `Ready For Dev` label after quality checks.
- Continue until the eligible backlog is fully refined.

**Will not:**

- Implement issue code.
- Label issues without repository-specific investigation.
- Mark epics as Ready For Dev.
- Overwrite original requirements or concurrent edits.
- Rubber-stamp ambiguous, obsolete, or oversized issues.

