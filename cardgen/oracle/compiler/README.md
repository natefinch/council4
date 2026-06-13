# Oracle compiler

Package `compiler` owns semantic compilation and the semantic intermediate
representation consumed by card generation. `Compile(parser.Document, Context)`
lowers an already parsed document; it never accepts raw source or invokes the
top-level `parser.Parse` entry point. Legacy compiler paths still call
`parser.ParseSentences` on retained source as a transitional dependency while
their literal-text interpretation moves into typed parser syntax. Compiler
`Context` contains only `CardName`.

The package owns compiled abilities and content, semantic trigger patterns,
conditions, references, selectors, static declarations, costs, effects, and
fail-closed semantic diagnostics. Typed parser paths—phase/step and player-event
triggers, activation restrictions, and static-rule syntax—compile from typed
nodes without consulting retained literal text.

`compiler` imports `parser` and `shared`. Cardgen lowering consumes compiler IR
and imports parser syntax only where exact source metadata is still required for
strict consumption checks.
