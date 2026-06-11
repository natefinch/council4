# Example: Autonomous Ready-for-Development Loop

## User Request

> Run impl-loop and keep expanding card support.

## Result

The agent inventories open `Ready For Dev` issues, removes issues already
referenced by an epic, and chooses a cohesive high-impact batch of up to five.
It creates an `Epic` issue listing those children before changing code.

For each child, the agent creates a fresh branch from `origin/main`, implements
and validates the issue, runs the `cardSupport` Mage target, audits the resulting
generated cards and support-documentation changes, files any newly discovered
deferred card work, runs independent reviews until clear, and merges a dedicated
pull request. It updates and closes the epic, then repeats with another eligible
batch until none remain.
