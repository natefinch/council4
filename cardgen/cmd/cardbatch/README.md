# cardbatch

`cardgen/cmd/cardbatch` is the deterministic command-line workflow for
card-generation batches. It owns disk input/output and resumable manifest state;
the `card-impl` agent skill remains responsible for per-card oracle-text
implementation.

## Commands

```bash
go run ./cardgen/cmd/cardbatch parse -in cards.txt -out .cardwork/cards.json
go run ./cardgen/cmd/cardbatch fetch -manifest .cardwork/cards.json
go run ./cardgen/cmd/cardbatch missing -manifest .cardwork/cards.json
go run ./cardgen/cmd/cardbatch worklist -manifest .cardwork/cards.json -repo . -limit 10
go run ./cardgen/cmd/cardbatch validate -manifest .cardwork/cards.json -generate
go run ./cardgen/cmd/cardbatch report -manifest .cardwork/cards.json -repo .
```

- `parse` reads a plain-text card list and writes a unique-card manifest.
- `fetch` downloads Scryfall oracle data into the manifest, using
  `.cardwork/cache/scryfall` by default.
- `missing` marks manifest rows as existing or missing based on the generated
  source path under `mtg/cards/<letter>/`.
- `worklist` prints missing or invalid card names, optionally as `card-impl`
  commands, so an agent can attempt a small batch without the Go tool invoking
  the skill directly.
- `validate` optionally regenerates card package lists and validates existing
  generated card definitions. Failed validation leaves the manifest row invalid
  with issues instead of marking the card implemented.
- `report` writes Markdown and JSON reports for fetch errors, missing generated
  files, and invalid generated cards.

The manifest is workflow state, not the source of truth. Generated card files,
the card registry, and validation results decide whether a card is actually
supported.
