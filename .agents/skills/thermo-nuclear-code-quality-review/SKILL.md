---
name: thermo-nuclear-code-quality-review
description: Opt-in deep Go maintainability review for a user-selected diff or design. Use only when explicitly asked for a thermo-nuclear or unusually deep architectural review.
disable-model-invocation: true
---

# Thermo-Nuclear Code Quality Review

This is an **opt-in** review, not a default gate for issue completion.
Follow the [repository policy](../../../.github/copilot-instructions.md)
and stay within the requested scope.
A routine authored-card change does not need an architectural audit.

## Review

- Read enough surrounding code to understand ownership and behavior.
- Find concrete simplifications that delete unnecessary complexity rather than
  move it between wrappers. Prefer ordinary Go and existing shared operations.
- Check type safety, errors, mutable ownership, resource lifetimes, dependency
  direction and whether the interface earns its implementation complexity.
- In card work, preserve inspectable declarations, engine-owned mutations,
  batch/timing rules, hidden information, object identity and clone isolation.
  Do not redesign authored Go into another compiler/DSL for stylistic consistency.
- Distinguish substantive bugs or maintainability problems from preferences.
  File size alone is not a defect; do not split declarations merely to satisfy
  a 1,000-line limit or require a new package/subagent per file.
- Named-card behavior tests and shared-operation tests are complementary.
  Do not replace runtime scenarios with structural assertions or hashes.
- Existing tests and interfaces are evidence, not permission to broaden the
  assignment into unrelated cleanup.

## Output and follow-through

Report actionable findings with relevant files, the consequence and a simpler
remedy. State uncertainty rather than inventing a defect. Do not require changes
just to make a review look useful.

Review does not itself authorize edits, commits, pushes, PRs or merges.
If asked to fix findings, make focused changes and rerun affected checks.
Do not repeatedly restart this broad audit after every edit; use a narrow
follow-up for a substantive corrected finding.
