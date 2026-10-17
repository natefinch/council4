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
authored sources in `mtg/cards/authored`. Test shared operations and actual
named-card interactions, then validate production registration following the
[repository policy](../../.github/copilot-instructions.md#verify-proportionally).

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

## Retained legacy definitions

The complete pinned compiler-accepted corpus is checked in: **17,946 card
definitions and 754 token definitions** (18,700 total). Existing definitions
are preserved rather than overwritten by the additive import; the exact
installed identities and paths are recorded in
[`installed-compiler.json`](../../testdata/card-preservation/installed-compiler.json).
No generation or Oracle download is needed to use a fresh checkout.
The bulk compiler/import workflow is retired; do not repeat it for new support.

Authored implementations take precedence over their 24 overlapping compiler
definitions. The generated copies remain available in the letter packages.
The generated Demonic Tutor and Diabolic Tutor definitions are also present,
but the default playable registry excludes them until the proven private-search
placement defect [#3289](https://github.com/natefinch/council4/issues/3289) is fixed.
Compiler acceptance and source presence are not runtime certification.

Focused legacy correctness fixes use the existing code and behavioral tests,
not corpus recompilation or fingerprint baselines.

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
