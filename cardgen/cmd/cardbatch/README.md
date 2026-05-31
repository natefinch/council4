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
```

- `parse` reads a plain-text card list and writes a unique-card manifest.
- `fetch` downloads Scryfall oracle data into the manifest, using
  `.cardwork/cache/scryfall` by default.
- `missing` marks manifest rows as existing or missing based on the generated
  source path under `mtg/cards/<letter>/`.

The manifest is workflow state, not the source of truth. Generated card files,
the card registry, and validation results decide whether a card is actually
supported.
