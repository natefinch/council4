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

Each letter package exports constructor entries (`[]cardset.Entry`) that build
immutable `game.CardDef` values lazily. A `go generate` step maintains the package
lists and `registry_sets.go` for `NewDefaultRegistry`.

## Adding card support

Author Go behavior over shared Engine operations. Preserve costs, targets,
triggers and static/replacement declarations as inspectable data. Oracle
recognition is not required; do not edit parser/lowering to add a card.
See [ADR 0012](../../docs/adr/0012-authored-go-card-implementations.md) and
[epic #3269](https://github.com/natefinch/council4/issues/3269).

The [versioned program context/catalog](../cardprogram/README.md) handles
synchronous Go resolving bodies for spells, activations, triggers and chapters.
Static/replacement applicability remains declarative; later migration slices
deepen those seams. The spell-only legacy escape hatch is kept for compatibility,
not a bypass for missing operations or policies.

Keep existing generated definitions and their runtime tests. Place future
authored sources in `mtg/cards/authored`; never point corpus regeneration at
them. Test shared operations and
actual named-card interactions, then validate production registration.

Skullbriar's haste and combat-damage trigger are ordinary declarations using
shared Engine operations. Its static `ZoneTransitionPolicy` is a typed
counter-retention exception, evaluated from the departing object's applicable
abilities and the **actual** destination after replacements (CR 113.6c, 122.2).
Off-battlefield counters, including existing exile-marker mechanics, share
`Game.CardCounters`; battlefield counters live only on the permanent. Retention
neither places counters again nor preserves object identity, temporary copy
effects, targets or links. Casting evaluates retention before costs can change
ability sources, and failed proposals restore the prior counter container.
The acceptance scenarios use the pinned Oracle corpus and the official
2026-09-25 CR with Skullbriar's 2011-09-22 Oracle rulings.

## Legacy regeneration

For explicitly assigned compiler correctness or corpus refresh work, compile
to a temporary tree and inspect it before intentionally updating generated
definitions. Keep the same corpus input for baseline/candidate comparisons.
See [`compilecards`](../../cardgen/oracle/cmd/compilecards/README.md).
Regeneration is not an every-card gate or permission to overwrite authored files.

After an intentional change to existing constructor entries, maintain the lists:

```bash
go generate ./mtg/cards/...
```

## Using the registry

To resolve names against the full committed corpus (every letter sub-package,
excluding tokens), use `NewDefaultRegistry` — this is what the decklist loader
uses so callers don't import every letter package by hand:

```go
import "github.com/natefinch/council4/mtg/cards"

reg := cards.NewDefaultRegistry()
bolt := reg.Lookup("Lightning Bolt")
```

To build a lazy registry from a specific subset, pass constructor entries:

```go
import (
    "github.com/natefinch/council4/mtg/cards"
    "github.com/natefinch/council4/mtg/cards/l"
    "github.com/natefinch/council4/mtg/cards/s"
)

reg := cards.NewLazyRegistry(l.Cards, s.Cards)
bolt := reg.Lookup("Lightning Bolt")
```

`Lookup` returns the first registered definition for a name. Use `LookupAll`
when distinct Oracle cards share the same printed name.

Playable tokens are not cards that can be looked up for deck construction, so
they are not added to `Registry`. Import `mtg/cards/tokens` and use
`tokens.Cards`, or import a token letter package directly. Every generated token
filename and Go variable includes its complete normalized Oracle UUID; this
keeps same-name token definitions distinct and stable across Scryfall printings.
