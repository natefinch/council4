# Oracle compiler

Package `compiler` owns semantic compilation and the semantic intermediate
representation consumed by card generation. `Compile(parser.Document, Context)`
lowers an already parsed document; it never accepts raw source or invokes the
top-level `parser.Parse` entry point. Legacy compiler paths still call
`parser.ParseSentences` on retained source as a transitional dependency while
their literal-text interpretation moves into typed parser syntax. Compiler
`Context` is currently empty; card-name and source-reference recognition belongs
to parser syntax.

The package owns compiled abilities and content, semantic trigger patterns,
conditions, references, selectors, static declarations, costs, effects, and
fail-closed semantic diagnostics. Typed parser paths—phase/step and player-event
triggers, activation restrictions, and static-rule syntax—compile from typed
nodes without consulting retained literal text.

Reusable semantic atoms—colors including excluded/non-color forms, card types
including excluded/non-type forms, supertypes, subtypes, object nouns, zones,
counters, cardinal and ordinal numbers, and explicit self/source references—are
recognized by the parser and mapped here to engine types (`color.Color`,
`types.Card`, `types.Sub`, `zone.Type`, `counter.Kind`, and reference kinds).
The core effect, keyword, target, reference, amount, zone, counter, trigger, and
condition compilation consumes parser atoms by span rather than deriving these
meanings from token spelling. Genuine identity values, such as subtypes, remain
typed engine values.

Later-family grammar may still inspect retained text to identify a whole phrase
production, but reusable atom meanings inside those productions are consumed from
parser-emitted, source-spanned atoms.

`compiler` imports `parser` and `shared`. Cardgen lowering consumes compiler IR
and imports parser syntax only where exact source metadata is still required for
strict consumption checks.
