# genregistry

`cmd/genregistry` is the registration-list command used by `go generate` in the
`mtg/cards` package. It scans `mtg/cards` for single-letter card sub-package
directories (`a/`, `b/`, …) and writes `registry_sets.go`, which aggregates each
package's exported `Cards` slice into the `defaultCardSets` function backing
`cards.NewDefaultRegistry`.

Token definitions (the `tokens/` sub-package) are intentionally excluded: they
are not real cards and must not resolve from a decklist.

## Usage

Run it through the existing package directive:

```bash
go generate ./mtg/cards/...
```

`mtg/cards` invokes:

```go
//go:generate go run github.com/natefinch/council4/cmd/genregistry
```

Re-run it after adding a new letter sub-package so the new package is included in
the default registry automatically.

This only maintains registry lists for existing Go definitions. It does not
interpret Oracle text or generate card implementations.
