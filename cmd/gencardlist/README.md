# gencardlist

`cmd/gencardlist` is the registration-list command used by `go generate` in
`mtg/cards/<letter>` packages. It scans the current card sub-package for exported
card constructors (or eager token definitions) and writes that package's
generated `cards.go` registration list.

## Usage

Run it through the existing package directives:

```bash
go generate ./mtg/cards/...
```

Each letter package invokes:

```go
//go:generate go run github.com/natefinch/council4/cmd/gencardlist
```

This only maintains registry lists for existing Go definitions. It does not
interpret Oracle text or generate card implementations.
