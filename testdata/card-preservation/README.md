# Retained card input and installation evidence

Corpus recompilation, registry/source fingerprint baselines and refresh tooling
are retired. Follow the
[repository validation policy](../../.github/copilot-instructions.md#verify-proportionally)
and [authored-Go guide](../../docs/card-authoring.md), not the old preservation
checklists in Git history.

`corpus.json` identifies the exact Scryfall input bytes, publication date and rules
edition used by existing authoring/reporting tools. Input-integrity hashes
identify data; they are not executable-source bindings or gameplay evidence.
Explicit reporting maintenance reads existing local input and compiler evidence;
it does not download or compile a corpus.

`installed-compiler.json` retains the completed import's **18,700** accepted
identities: 17,946 ordinary/multiface cards and 754 token definitions. Its recorded
source paths are installation evidence, not a regeneration task or behavior
certificate. Existing generated sources remain installed.
Authored replacements supersede matching compiler definitions without deleting
them; `cards.CompilerExclusions()` identifies known runtime blockers.

The retired baseline captured 2,779 name-only registry entries and 591 token
definitions, structurally loadable at that checkpoint. Candidate Oracle matches
were not exact identity proof. Those historical counts are not current support
claims, and no baseline/waiver refresh is required.

Behavioral regression coverage remains in named-card and shared-rule scenarios,
including `mtg/rules/lki_test.go`, `replacement_property_test.go`,
`continuous_property_test.go` and `mtg/game/clone_test.go`. Preserve these useful
tests and exact authored identity/program evidence.
