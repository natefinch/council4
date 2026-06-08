# checkparser

`checkparser` runs the `cardgen/oracle` lexer and syntax parser over every
non-empty root and card-face Oracle text in a Scryfall card bulk-data array.

The command stream-decodes the input and uses a bounded parser worker pool.
Reports are deterministic and contain the Scryfall identity, card/face name,
Oracle text, diagnostic severity, summary, detail, and exact source span.

```bash
go run ./cardgen/oracle/cmd/checkparser \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/oracle-parser-report.json
```

Flags:

- `-in`: required Scryfall card bulk-data JSON array.
- `-out`: report path, or `-` for standard output. Default `-`.
- `-format`: `json` or `text`. Default `json`.
- `-workers`: parser worker count. Default `runtime.NumCPU()`.
