# corpusdelta

`corpusdelta` runs the repetitive validation and review preparation for an
Oracle compiler expansion. It:

1. runs `compilecards` into a clean `.cardwork` output directory;
2. compares the previous and current reports by stable Scryfall card ID;
3. verifies input, eligible, excluded, generated, and unsupported report counts
   and every newly generated source path;
4. regenerates `docs/supported.md`;
5. writes a deterministic JSON inspection manifest containing Oracle text,
   generated source paths, regressions, and diagnostic-count changes; and
6. copies the generated tree into a temporary package under `cardgen`, then runs
   `go test` and `go vet` on every generated package.

Run it from the repository root:

```bash
go run ./cardgen/oracle/cmd/corpusdelta \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -baseline .cardwork/step-7-report.json \
  -out .cardwork/step-8-generated \
  -report .cardwork/step-8-report.json \
  -manifest .cardwork/step-8-delta.json
```

The generated output root is deliberately restricted to `.cardwork` so this
workflow cannot overwrite the hand-maintained `mtg/cards` tree.

Flags:

- `-in`: required Scryfall Oracle Cards JSON array.
- `-baseline`: required previous `compilecards` JSON report.
- `-out`: generated-card root. Default `.cardwork/current-generated`.
- `-report`: current report path. Default `.cardwork/current-report.json`.
- `-manifest`: inspection manifest path. Default
  `.cardwork/current-delta.json`.
- `-supported`: supported-card Markdown path. Default `docs/supported.md`.
- `-compile`: run `compilecards` before comparison. Default `true`.
- `-validate`: test and vet generated packages. Default `true`.

Use `-compile=false` only when inspecting an existing report and generated tree.
Reports made before explicit corpus exclusions remain readable: their
`card_count` is both the input and eligible count, and their excluded set is
empty.
