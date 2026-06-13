# Oracle parser

Package `parser` owns Oracle syntax and grammatical recognition.
`Parse(source, Context)` lexes source and returns a lossless `Document` plus
localized diagnostics. `Context` contains only card-face facts needed to
classify syntax: `InstantOrSorcery`, `Planeswalker`, `Saga`, and the card's own
`CardName` (used to recognize self references). `CardName` is threaded onto the
returned `Document`.

The package owns syntax ability kinds, source-spanned phrases and sentences,
literal Oracle vocabulary, typed trigger clauses, activation restrictions,
static-rule syntax, and attached-subject selection syntax. Unrecognized or
ambiguous grammar preserves source metadata and fails closed rather than
inventing typed syntax.

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

`parser` imports `lexer` and `shared`, never `compiler`. `ParseSentences` remains
a narrow transitional API for legacy compiler paths that still recognize
sentence text.
