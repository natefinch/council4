# compilecards

`compilecards` stream-decodes a Scryfall Oracle Cards bulk-data array, compiles
cards in parallel, and writes deterministic Go definitions for only the cards
whose complete rules text is supported by the executable backend.

The initial strict backend supports vanilla faces and plain, non-parameterized
keyword abilities that have reusable `mtg/game` templates. It never emits TODOs
or partial ability implementations. Unsupported cards, unsupported layouts,
source-generation failures, non-ASCII package names, and filename collisions
are written to the report.

Writes are serialized after compilation. Existing files at matching generated
paths are overwritten. Each affected letter package's `cards.go` registry is
then regenerated from all CardDef declarations in that directory.

For a safe full-corpus trial, target a temporary cards root:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/generated-cards \
  -report .cardwork/oracle-compile-report.json
```

To overwrite matching repository card files:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out mtg/cards \
  -report .cardwork/oracle-compile-report.json
```

Flags:

- `-in`: required Scryfall Oracle Cards JSON array.
- `-out`: cards package root. Default `mtg/cards`.
- `-report`: unsupported report path, or `-` for standard output. Default `-`.
- `-format`: `json` or `text`. Default `json`.
- `-workers`: compiler/source-generator worker count. Default
  `runtime.NumCPU()`.
