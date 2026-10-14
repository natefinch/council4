# Example: Work through the migration

## User request

> Run impl-loop on #3269. Make local commits directly on main; do not push.

## Result

The agent reads the epic's live prerequisites and takes the first unblocked
child, preserving existing assignments. It does not create another epic or ask
for a PR strategy. After #3271, the next migration child is #3270, not the
largest unsupported compiler diagnostic.

For code, it runs focused tests during development and the test suite/linter
once before committing. Docs-only changes use documentation checks instead.
It stages intended files, records the local commit and advances only when the
child's acceptance criteria are met. It does not claim that local commits are
published or that blocked work is complete.
