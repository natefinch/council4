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

Every supported trigger family reaches `TriggerPattern` through a mechanical
typed adapter. Phase/step, player-event, zone-change, spell/ability, combat,
damage, permanent-state, counter, sacrifice, mutate, and targeting meaning is
already present in parser syntax. `trigger_pattern.go` maps only closed enums,
typed selections, relations, zones, recipients, causes, and qualifiers; it
contains no Oracle trigger wording recognizers or subject-text parsing. Invalid
or partial constructed syntax fails closed, and retained event text and tokens
cannot change compilation. Event-history conditions likewise arrive as typed
parser event syntax and a typed turn window; condition compilation reuses the
same mechanical trigger adapters.

`condition.go` compiles the remaining conditions from typed parser
`ConditionClause` nodes matched to each condition by source span. It maps the
parser's closed predicate, control scope, comparison, threshold, counter, object
binding, subject span, and `ConditionSelection` onto the engine condition
vocabulary, and contains no Oracle condition wording recognition: no
prefix/suffix/contains text matching and no token-spelling interpretation. A
numeric "at most N" comparison becomes a negated "at least N+1" minimum, and the
introducer kind supplies the base negation. Any clause whose typed selection,
counter, scope, or predicate falls outside the closed semantic vocabulary leaves
the predicate unsupported. `compiler.go` still locates condition boundaries and
classifies each introducer kind, intervening-if, and duration-skip structurally;
that boundary framing is deliberately retained there, separate from the
text-blind predicate mapping in `condition.go`.

`static_declaration.go` compiles static declarations from the typed
`StaticDeclarationSyntax` nodes the parser emits, matched to the ability by
declaration family and consumed mechanically. It dispatches on the parser's
ordered declaration kinds—power/toughness change, keyword grant, rule,
cost modifier, and card-ability grant—and reads the affected group, deltas,
granted keywords, rule meaning, cost shape, and card filter from already-compiled
content and typed parser payloads. It contains no Oracle static-declaration
wording recognition: no `matches*`/token-sequence recognizers, no
prefix/suffix/contains text matching, and no token-spelling interpretation.
Source/group asymmetries (a source keyword grant requires a condition; a group
grant forbids one), the dynamic-amount agreement check, and the supported-rule
table are enforced over typed nodes and compiled effects alone. Any ability whose
typed declarations or compiled content fall outside the closed vocabulary records
a structural blocker instead of a declaration.

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

Later-family grammar outside resolving effects and trigger events may still
inspect retained text to identify a whole phrase production, but reusable atom
meanings inside those productions are consumed from parser-emitted,
source-spanned atoms.

`compiler` imports `parser` and `shared`. Cardgen lowering consumes compiler IR;
retained source metadata remains available for diagnostics and strict source
consumption checks.
