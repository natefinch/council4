# Oracle parser

Package `parser` owns Oracle syntax and grammatical recognition.
`Parse(source, Context)` lexes source and returns a lossless `Document` plus
localized diagnostics. `Context` contains only card-face facts needed to
classify syntax: `InstantOrSorcery`, `Planeswalker`, `Saga`, and the card's own
`CardName` (used to recognize self references). `CardName` is threaded onto the
returned `Document`.

The package owns syntax ability kinds, source-spanned phrases and sentences,
literal Oracle vocabulary, typed trigger clauses, activation restrictions,
static-rule syntax, resolving-effect syntax, and selection syntax. Unrecognized
or ambiguous grammar preserves source metadata and fails closed rather than
inventing typed syntax.

`effect_syntax.go` composes resolving instructions from parser-owned productions.
Each `Sentence` carries ordered, source-spanned `EffectSyntax` and `TargetSyntax`
nodes. Effects carry their typed verb and contextual variant, fixed or dynamic
amount, power/toughness deltas, duration and delayed timing, local Selection,
origin and destination zones, counter kind, exact add-mana output, replacement
modifier, static subject, references, and embedded resolution payment. Each
effect also owns its exact clause, targets, references, and grammatical-subject
targets/references; coordinated follow-ons carry an explicit prior-subject
context instead of inferring it from verb spelling. Targets carry typed cardinality
and a Selection containing object kind, controller relation, flags, types,
supertypes, subtypes, colors, keyword, zone, and numeric filters. Retained text
and tokens are lossless metadata, not the source of downstream meaning.
Target selections require every token in the noun phrase to belong to a typed
atom or a narrow composition production; unknown qualifiers and unknown
cardinalities invalidate the target rather than weakening it.

Effect grammar excludes activation costs, trigger introductions, reminder text,
quoted text, and typed trailing activation restrictions. Coordinated instructions
remain ordered clauses, while malformed dynamic formulas, payments, contextual
verbs, and target forms fail closed at the narrow production that could not be
recognized. Specialized replacement modifiers are attached only to the replacing
effect and reject selection modifiers that the runtime replacement cannot
represent.

It also owns the reusable, composable semantic atoms that downstream stages
consume without re-inspecting source spelling. `atoms.go` recognizes colors,
card types, supertypes, subtypes, object nouns, zones, counter kinds, cardinal
and ordinal number words, selection modifiers, and plural→singular noun
normalization, returning typed values. `keyword.go` owns the complete supported
keyword vocabulary and emits source-spanned `Keyword` syntax with composable
typed parameter shapes: mana costs, integers, Enchant targets, and Protection
predicates over colors, card types, and creature/land subtypes. It also emits
typed `with`/`without` keyword-selector syntax. Mana-symbol parsing, canonical
keyword names, Protection list grammar, and Enchant target normalization live
only in the parser; malformed or ambiguous parameter grammar leaves the keyword
unparameterized and therefore fails closed downstream. `references.go`
recognizes explicit self/source references (the card's own name, `this`/`that`
objects, and exact pronouns) as typed `Reference` values. `Parse` emits these atoms
as source-spanned typed values attached to each `Ability` and modal `Mode` node
(the `Atoms` field), so the compiler consumes them by span rather than calling
recognizers on raw tokens. Recognizers fail closed on unknown or ambiguous
spelling.

`parser` imports `lexer` and `shared`, never `compiler`. `ParseSentences` is the
lossless sentence splitter used internally and remains available to syntax
clients; semantic compilation consumes the typed nodes emitted by `Parse`.
