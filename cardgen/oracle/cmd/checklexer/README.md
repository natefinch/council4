# checklexer

`checklexer` scans every rules-text field in a Scryfall card bulk-data JSON
array and reports cards whose Oracle text the `cardgen/oracle` lexer cannot
tokenize.

The command streams the bulk file instead of loading all card objects into
memory. A bounded worker pool checks root `oracle_text` fields and every
non-empty `card_faces[].oracle_text` field in parallel. Results are sorted back
into input and face order, so reports are deterministic regardless of worker
count.

## Usage

From the repository root:

```bash
go run ./cardgen/oracle/cmd/checklexer \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/oracle-lexer-report.json
```

Flags:

- `-in`: required path to a Scryfall card bulk-data JSON array.
- `-out`: report path; `-` writes to standard output. Default `-`.
- `-format`: `json` or `text`. Default `json`.
- `-workers`: concurrent lexer workers. Default `runtime.NumCPU()`.

JSON output contains total card and Oracle-text counts plus an `unsupported`
array. Each unsupported entry includes the Scryfall card and Oracle IDs, card
and face names, set and collector number, exact Oracle text, and all lexical
issues with reasons and source spans.

The text format is intended for terminal use:

```bash
go run ./cardgen/oracle/cmd/checklexer \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -format text
```

An empty `unsupported` array means every non-empty Oracle text in the bulk file
reached EOF without a `shared.Invalid` token.
