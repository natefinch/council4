# Implementation Loop

Implements a selected epic's earliest unblocked child, or an eligible
`Ready For Dev` issue within the user's assigned scope. New card work uses
authored Go over shared Engine operations, not compiler diagnostic rankings.

The default is implement, test, inspect and commit directly to main when
authorized, with one writer and no mandatory PR or nested epic. Checks and review
follow the [repository policy](../../../.github/copilot-instructions.md#verify-proportionally).
Push permission remains explicit.

Blocked, parked and superseded issues stay out of the queue. Existing runtime
prerequisites and unmerged work are preserved. See [SKILL.md](SKILL.md).
