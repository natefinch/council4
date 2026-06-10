# Example: Refine the Entire Backlog

## User Request

> Run the ready-for-dev loop.

## Result

The agent lists all open, non-epic issues without `Ready For Dev`, chooses a
batch of up to five, and investigates each issue against the repository.

For each actionable issue, it preserves the original body and adds an
`Implementation guidance` section containing the current limitation, suggested
approach, relevant files, acceptance criteria, validation commands, and
non-goals. It applies `Ready For Dev` only after that handoff is complete.

Duplicates, obsolete issues, and already implemented work are closed with
evidence. Broad coordination issues are decomposed into implementation-sized
children. The agent refreshes the issue inventory after every batch and continues
until no eligible issue remains unrefined.

