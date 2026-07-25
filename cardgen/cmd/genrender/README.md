# genrender

`cardgen/cmd/genrender` generates the parts of the cardgen renderer that are
fully determined by the shape of the `mtg/game` value types.

## Why it exists

The renderer emits Go source for generated card definitions, so it needs a
mapping from every constant a game value can hold to the Go expression naming
that constant. Those mappings used to be hand-written `switch` statements in
`cardgen/render_enum.go`, which meant every constant added to `mtg/game` needed a
matching renderer edit, and a missing one failed silently at card-compile time
rather than at build time.

The hand-written tables had drifted. At the time this command was introduced they
covered 252 constants; the reachable game types declare over 1,500. Concretely,
`renderStep` mapped 7 of `game.Step`'s 14 values — untap, declare blockers,
first-strike damage, combat damage, and cleanup had no rendering at all — and
`zone.Stack`, `game.DurationNextTime`, and `game.ResolutionChoiceCardName` were
also unmapped. A lowering that produced any of them would have failed to render,
and the card would have been dropped.

## What it generates

`cardgen/render_literals_generated.go`, containing one table per enum type that
is reachable from `game.CardDef` or `game.CardFace`:

- a `map[T]string` of constant to Go expression for ordinary enums;
- an ordered `[]enumFlag[T]` for the bitmask types listed in `bitmaskTypes`,
  whose values are combinations of flags rather than single constants.

Reachability is computed by walking the type graph from the root types through
struct fields, slice and map elements, pointers, generic type arguments, and
interface implementations. Restricting generation to reachable types keeps the
file to constants the renderer can actually emit.

## Usage

```bash
go generate ./cardgen/
```

The directive lives in `cardgen/doc.go`. The command must run with its working
directory inside this module: it type-checks `mtg/game` from source using the
standard library's `go/importer` source compiler, which resolves imports through
`go/build`. It deliberately uses only the standard library, because the module
has no third-party dependencies.

To inspect the type graph without writing a file:

```bash
go run ./cardgen/cmd/genrender -list
```

## Guardrails

- `TestGeneratedFileIsCurrent` regenerates the file and fails if it differs from
  the committed copy, so the tables cannot go stale.
- `TestBitmaskTypesAreFlagShaped` fails if a type named in `bitmaskTypes` is no
  longer reachable, or if any of its non-zero constants is not a distinct power
  of two. Flag shape is not detectable from values alone — an enum with constants
  0, 1, and 2 looks exactly like a bitmask — so the list is explicit and this
  test keeps it honest.
- `TestRootTypesAreReachable` pins shapes the walk must find, including
  `game.ManaSpendRider`, which is reachable only through `opt.V[ManaSpendRider]`
  and so guards the generic type-argument traversal.
