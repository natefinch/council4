# Oracle compiler

Package `compiler` owns semantic compilation and the semantic intermediate
representation consumed by card generation. `Compile(parser.Document, Context)`
lowers an already parsed document; it never accepts raw source or invokes the
top-level `parser.Parse` or `parser.ParseSentences` entry points. Compiler
`Context` is currently empty; card-name, source-reference, and resolving-effect
recognition belong to parser syntax.

The package owns compiled abilities and content, semantic trigger patterns,
conditions, references, selectors, static declarations, costs, effects, and
fail-closed semantic diagnostics. Typed parser paths—resolving effects, targets and selections, amounts, durations,
zones, counters, add-mana output, replacement modifiers, references, embedded effect payments, phase/step and
player-event triggers, activation restrictions, and static-rule syntax—compile
from typed nodes without consulting retained literal text.

Reusable semantic atoms—colors including excluded/non-color forms, card types
including excluded/non-type forms, supertypes, subtypes, object nouns, zones,
counters, cardinal and ordinal numbers, and explicit self/source references—are
recognized by the parser and mapped here to engine types (`color.Color`,
`types.Card`, `types.Sub`, `zone.Type`, `counter.Kind`, and reference kinds).
Keywords and keyword selectors are likewise parser-owned typed syntax.
`CompiledKeyword` carries the parser-recognized keyword kind plus typed mana,
integer, Enchant-target, or Protection parameter data; canonical name and
parameter text remain metadata only. The compiler maps typed Protection atoms
to engine predicates but performs no keyword-name or parameter recognition.
`effect_syntax.go` is a mechanical adapter from parser resolving syntax to
compiler IR. It maps enums and typed values; it contains no Oracle vocabulary or
effect recognizers. The core effect, keyword, target, reference, amount, zone,
counter, trigger, and condition compilation consumes parser syntax or atoms by
span rather than deriving these meanings from token spelling. Genuine identity
values, such as subtypes, remain typed engine values. Compiled effects preserve
parser-owned clause, target, reference, and grammatical-subject ownership so
ordered-effect lowering does not rediscover clause boundaries from tokens.

Later-family grammar outside resolving effects may still inspect retained text
to identify a whole phrase production, but reusable atom meanings inside those
productions are consumed from parser-emitted, source-spanned atoms.

`compiler` imports `parser` and `shared`. Cardgen lowering consumes compiler IR;
retained source metadata remains available for diagnostics and strict source
consumption checks.
