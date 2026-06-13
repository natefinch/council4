# Oracle parser

Package `parser` owns Oracle syntax and grammatical recognition.
`Parse(source, Context)` lexes source and returns a lossless `Document` plus
localized diagnostics. `Context` contains only card-face facts needed to
classify syntax: `InstantOrSorcery`, `Planeswalker`, and `Saga`.

The package owns syntax ability kinds, source-spanned phrases and sentences,
literal Oracle vocabulary, typed trigger clauses, activation restrictions,
static-rule syntax, and attached-subject selection syntax. Unrecognized or
ambiguous grammar preserves source metadata and fails closed rather than
inventing typed syntax.

`parser` imports `lexer` and `shared`, never `compiler`. `ParseSentences` remains
a narrow transitional API for legacy compiler paths that still recognize
sentence text.
