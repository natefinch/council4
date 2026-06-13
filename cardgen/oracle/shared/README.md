# Oracle shared infrastructure

Package `shared` owns generic source and token infrastructure used across Oracle
pipeline stages:

- byte/rune-aware `Position` and half-open `Span`;
- lexical `Kind` and lossless `Token`;
- source-spanned `Diagnostic` and `Severity`;
- generic token-list and source-slicing helpers.

The package contains no Oracle words, grammatical recognition, semantic
recognition, stage dispatch, parser context, or compiler context. It is
transitional: when a helper no longer has multiple stage consumers, move it to
the stage that owns it rather than growing `shared`. Source-spanned typed Oracle
atoms and their vocabulary belong to `parser`, not here.

Dependency direction starts here:

```text
shared <- lexer <- parser <- compiler
```
