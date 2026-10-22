# Implementation Loop

Implements a selected epic's earliest unblocked child, or an eligible
`Ready For Dev` issue within the user's assigned scope. New card work uses
authored Go over shared Engine operations, not compiler diagnostic rankings.
Broadly authorized continuous production also replenishes small reviewed
saved batches and works their canonical Engine prerequisites, then explicitly
reassesses dependent cards; exhausted saved batches alone do not end that scope.

The production path is implement, integrate latest main, test and inspect, then
open a PR and merge only after all exact-head CI succeeds when authorized.
Keep one integration writer and no duplicate epic. Checks and review
follow the [repository policy](../../../.github/copilot-instructions.md#verify-proportionally).
Push permission remains explicit.

Blocked, parked and superseded issues stay out of the queue. Existing runtime
prerequisites and unmerged work are preserved. A blocked card remains unsupported
while its in-scope prerequisite is worked; it is not silently promoted on issue
closure. The coordinator alone changes queue ownership. See [SKILL.md](SKILL.md).
