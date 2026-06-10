# mtg/cards

Package `cards` provides a registry indexing `game.CardDef` values by canonical card name. Sanctioned card definitions are organized in letter-based sub-packages (`g/`, `l/`, `s/`, etc.) and aggregated by the `Registry` type. Playable token definitions live under `tokens/` and are intentionally separate from the name registry.

## Architecture

```
mtg/cards/
├── registry.go          # Registry type combining all sub-packages
├── g/
│   ├── doc.go           # go:generate directive
│   ├── glorious_anthem.go
│   └── cards.go         # generated: lists all Cards in package g
├── l/
│   ├── doc.go
│   ├── lightning_bolt.go
│   └── cards.go         # generated
├── s/
    ├── doc.go
    ├── serra_angel.go
    ├── sol_ring.go
    ├── soul_warden.go
    ├── swords_to_plowshares.go
│   └── cards.go          # generated
└── tokens/
    ├── cards.go          # generated: all playable token definitions
    └── b/
        ├── bear_<oracle-id>.go
        └── cards.go      # generated token letter list
```

Each card is an exported `*game.CardDef` variable in its letter sub-package (e.g., `var LightningBolt = &game.CardDef{...}`). A `go generate` step produces `cards.go` per sub-package listing all cards.

## Generating cards

1. Run `cardgen/oracle/cmd/compilecards` against a Scryfall Oracle Cards bulk
   file and a temporary output directory. The compiler emits only complete,
   validated Card Definitions and reports unsupported cards.
2. Inspect and validate the generated packages before copying selected files or
   intentionally targeting `mtg/cards` directly. See
   [`cardgen/oracle/cmd/compilecards`](../../cardgen/oracle/cmd/compilecards/README.md).
3. For an exceptional mechanic outside compiler coverage, write a Card
   Implementation manually and use `ImplementationID` only when declarative
   Effect Primitives cannot represent the behavior.
4. Regenerate the Card Registry lists after manual changes:
   ```bash
   go generate ./mtg/cards/...
   ```

5. If this is a new letter directory, add a `README.md` and a `doc.go` with the
   `go:generate` directive:
   ```go
   package x
   //go:generate go run github.com/natefinch/council4/cardgen/cmd/gencardlist
   ```

## Using the registry

```go
import (
    "github.com/natefinch/council4/mtg/cards"
    "github.com/natefinch/council4/mtg/cards/l"
    "github.com/natefinch/council4/mtg/cards/s"
)

reg := cards.NewRegistry(l.Cards, s.Cards)
bolt := reg.Lookup("Lightning Bolt")
```

`Lookup` returns the first registered definition for a name. Use `LookupAll`
when distinct Oracle cards share the same printed name.

Playable tokens are not cards that can be looked up for deck construction, so
they are not added to `Registry`. Import `mtg/cards/tokens` and use
`tokens.Cards`, or import a token letter package directly. Every generated token
filename and Go variable includes its complete normalized Oracle UUID; this
keeps same-name token definitions distinct and stable across Scryfall printings.
