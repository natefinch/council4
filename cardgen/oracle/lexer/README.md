# Oracle lexer

Package `lexer` is the lexical stage of the Oracle pipeline. `NewLexer` returns
a synchronous pull scanner whose `Next` method emits lossless `shared.Token`
values through `shared.EOF`. `InvalidReason` explains invalid tokens.

The lexer owns character scanning, whitespace handling, punctuation, braced
symbols, UTF-8 validation, and source positions. It deliberately knows no
Oracle vocabulary: all English words are `shared.Word` tokens.

`lexer` imports only `shared`. Its output is consumed by `parser`; it never
imports parser or compiler concepts.
