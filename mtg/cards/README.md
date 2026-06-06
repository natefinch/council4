# mtg/cards

Package `cards` provides a registry mapping canonical card names to `game.CardDef` values. Card definitions are organized in letter-based sub-packages (`g/`, `l/`, `s/`, etc.) and aggregated by the `Registry` type.

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
└── s/
    ├── doc.go
    ├── serra_angel.go
    ├── sol_ring.go
    ├── soul_warden.go
    ├── swords_to_plowshares.go
    └── cards.go          # generated
```

Each card is an exported `*game.CardDef` variable in its letter sub-package (e.g., `var LightningBolt = &game.CardDef{...}`). A `go generate` step produces `cards.go` per sub-package listing all cards.

## Adding a card

1. Generate the mechanical scaffold:
   ```bash
   go run .agents/skills/card-impl/main.go "Card Name"
   ```

2. Fill in the categorized ability fields on `game.CardFace` (use the `card-impl` Copilot skill or do it manually): `SpellAbility`, `ActivatedAbilities`, `ManaAbilities`, `LoyaltyAbilities`, `TriggeredAbilities`, `ReplacementAbilities`, and `StaticAbilities` as appropriate. Do not populate the legacy `Abilities` field. For plain non-parameterized keywords, append reusable `StaticAbilityBody` templates such as `game.FlyingStaticBody` or `game.DeathtouchStaticBody`. Front-face data lives in the embedded `game.CardFace` on `CardDef`; for double-faced cards, `Back` holds back-face data. Follow the expanded/raw-string source layout shown in `mtg/cards/k/karplusan_forest.go`.

3. Regenerate the card list:
   ```bash
   go generate ./mtg/cards/...
   ```

4. If this is a new letter directory, add a `doc.go` with the `go:generate` directive:
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
