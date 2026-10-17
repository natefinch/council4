# Retired compiled-corpus installer

The bulk import workflow is retired. New support follows the
[authored-Go guide](../../docs/card-authoring.md) and
[repository validation policy](../../.github/copilot-instructions.md#verify-proportionally),
not another compiler run or import.

The completed additive import preserved existing implementations and authored
sources. `installed-compiler.json` retains its identity/path
evidence; it is a historical installation record, not a refresh task or behavior
certificate. Existing generated cards and runtime tests remain.

The default registry prefers explicit authored replacements.
`cards.CompilerExclusions()` keeps proven runtime blockers out of deck lookup
without deleting their generated sources.
