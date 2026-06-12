# Implementation Loop

Runs the repository's issue-to-merge workflow on autopilot.

Use this skill to select high-impact issues labeled `Ready For Dev`, reserve
cohesive batches in a new issue carrying the exact `Epic` label, and implement
each child using the user's selected individual, stacked, or combined pull request
strategy. It skips issues already referenced by an epic and files focused,
traceable `To Be Triaged` and `Card Support TODO` issues for newly discovered
deferred card work.

The loop continues until no open, unreserved `Ready For Dev` issues remain or an
external blocker requires user input.
