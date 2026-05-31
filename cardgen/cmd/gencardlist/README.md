# gencardlist

`cardgen/cmd/gencardlist` is the generation command used by `go generate` in
`mtg/cards/<letter>` packages. It scans the current card sub-package for exported
`*game.CardDef` variables and writes that package's generated `cards.go` list.

## Usage

Run it through the existing package directives:

```bash
go generate ./mtg/cards/...
```

Each letter package invokes:

```go
//go:generate go run github.com/natefinch/council4/cardgen/cmd/gencardlist
```

The command intentionally lives under `cardgen/` with the other card-generation
tooling so runtime packages under `mtg/` only contain game, rules, card data, and
simulation code.
