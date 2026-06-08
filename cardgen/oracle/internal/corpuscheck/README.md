# corpuscheck

Package `corpuscheck` is shared command infrastructure for checking Scryfall
card bulk-data arrays.

`Check` stream-decodes the top-level card array, expands root and card-face
Oracle text fields, and sends them through a bounded worker pool. The caller
supplies the lexer, parser, or compiler-specific check function.

Only unsupported results are retained. They are sorted back into card and face
input order before being returned, making reports deterministic regardless of
worker count. `WriteText` provides the common terminal report format; commands
may also JSON-encode `Report` directly.
