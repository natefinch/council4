# Card Rollout Commands

## Parse a list

```bash
go run ./cardgen/cmd/cardbatch parse \
  -in cards.txt \
  -out .cardwork/run/cards.json
```

## Fetch Scryfall oracle data

```bash
go run ./cardgen/cmd/cardbatch fetch \
  -manifest .cardwork/run/cards.json \
  -cache .cardwork/run/cache/scryfall
```

## Print a worklist

```bash
go run ./cardgen/cmd/cardbatch worklist \
  -manifest .cardwork/run/cards.json \
  -repo . \
  -limit 3
```

Use command output when you want copy-pasteable per-card implementation commands:

```bash
go run ./cardgen/cmd/cardbatch worklist \
  -manifest .cardwork/run/cards.json \
  -repo . \
  -limit 3 \
  -format commands
```

## Validate attempts

```bash
go generate ./mtg/cards/...

go run ./cardgen/cmd/cardbatch validate \
  -manifest .cardwork/run/cards.json \
  -repo .
```

## Report unsupported cards

```bash
go run ./cardgen/cmd/cardbatch report \
  -manifest .cardwork/run/cards.json \
  -repo . \
  -md .cardwork/run/unsupported.md \
  -json .cardwork/run/unsupported.json
```
