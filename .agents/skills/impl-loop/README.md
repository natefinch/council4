# Implementation Loop

Implements a selected epic's earliest unblocked child, or an eligible
`Ready For Dev` issue within the user's assigned scope. New card work uses
authored Go over shared Engine operations, not compiler diagnostic rankings.

The default is implement, test, inspect and commit directly to main when
authorized, with one writer and no mandatory PR, nested epic or independent
review for routine changes. Run linters only when the user explicitly asks;
lint is not a local commit/completion gate. Push permission remains explicit.

Blocked, parked and superseded issues stay out of the queue. Existing runtime
prerequisites and unmerged work are preserved. See [SKILL.md](SKILL.md).
