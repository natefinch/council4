# Example: Prepare the next migration step

## User request

> Prepare the next unblocked issue in #3269.

## Result

The agent checks native dependencies and current completion evidence. After the
workflow reset, it refines #3270 and marks it ready if its prerequisite is met.
It leaves later children blocked instead of promoting the whole epic.

Real runtime prerequisites such as #3232 and #3263 remain discoverable.
Compiler-only expansion retains its history but is parked. Card handoffs target
authored Go and actual runtime scenarios, not disappearing compiler diagnostics.
