# Example: Autonomous Ready-for-Development Loop

## User Request

> Run impl-loop and keep expanding card support.

## Result

The agent inventories open `Ready For Dev` issues, removes issues already
referenced by an epic, and chooses a cohesive high-impact batch of up to five.
It creates an `Epic` issue listing those children, marks it `In Progress`, and
asks the user to select individual, stacked, or combined pull requests before
changing code.

For each child, the agent marks it `In Progress`, follows the selected branch and
pull request strategy, implements and validates the issue, runs the `cardSupport`
Mage target, audits the resulting generated cards and support-documentation
changes, files and records any newly discovered deferred card work, and runs one
independent review plus at most one full follow-up review. It updates and closes
the epic, then repeats with another eligible batch until none remain.
