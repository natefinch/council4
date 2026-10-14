# Authored Go card programs

`game.Mode.Program` is a data-only `ProgramReference{ID, Version}`. Costs,
targets, triggers and mode selection stay in the ordinary Card Definition.
The Engine resolves the selected mode synchronously through an installed
immutable `cardprogram.Catalog`, using the same rules as declarative cards.
No parser, interpreter, coroutine or copied Go call stack is involved.

## An ordinary body

```go
var drawThree = game.ProgramReference{
    ID: "oracle-uuid/spell/0",
    Version: "1",
}

func resolveDrawThree(ctx cardprogram.Context) error {
    outcome, err := ctx.Execute(game.Draw{
        Amount: game.Fixed(3),
        Player: game.TargetPlayerReference(0),
    })
    if err != nil {
        return err
    }
    // Ordinary Go may branch on actual outcomes or keep local variables.
    // Draw's scalar receipt is not exposed as a universal mutation count.
    _ = outcome
    return nil
}
```

Declare the target in `game.Mode.Targets` and bind its `Program` instead of a
`Sequence`. A mode cannot contain both. Attach that content to a spell,
activation, trigger or chapter as usual; declaration, costs, announcement,
target revalidation and priority remain Engine responsibilities.

Register a `cardset.Entry` with its exact `OracleID`, constructor, and
`Programs: []cardprogram.Program{{Reference: drawThree, Resolve: resolveDrawThree}}`.
Authored sources belong in `mtg/cards/authored`, not generated letter files.
The default registry includes those entries and exposes its `Programs()`.
Duplicate code IDs and duplicate Oracle identities fail; replacing a unique
legacy name requires explicit `ReplaceLegacy`. Use `LookupOracleID` to resolve
an exact authored identity when printed names are shared.

The synthetic parity card in tests is **not** production support. A new real
card is added only after its entire behavior and relevant scenarios are complete.

## Execution and ownership

- `Source()` returns captured card/object/stack IDs, controller and X. It does
  not expose a live permanent or Game pointer.
- `Targets()` returns a copy of the declared runtime targets. Rules operations
  still revalidate references; do not use plain Go writes as targeting/mutation.
- `Execute(primitive)` validates the operation's target scope and dispatches
  the existing Effect Resolver. Its `Outcome` distinguishes unavailable from
  executed-but-ineffective, with copied public result snapshots. It does not
  promise that every legacy scalar is an actual mutation count.
- `Ask(request)` uses normal ChoiceAgent observations, validation, logging and
  recorded fallbacks. It does not choose a strategic answer inside card code.
- `Destroy(operation)` performs one simultaneous batch through the shared
  destruction, replacement and zone-change machinery. Each member preserves
  its pre-action incarnation, characteristics and controller, whether it moved,
  its actual destination, and its arrived card incarnations. Protected members
  remain correlated but have no arrival; an unavailable single reference is
  distinct from an available empty group. `Before.ZoneCards` uses the actual
  pre-move versions, not LKI's predicted destination versions.
- `CreateTokens(definition, recipients...)` creates one simultaneous event across
  the declared `TokenRecipient{Player, Count}` quantities, using token/entry replacements
  and returns copied snapshots of the tokens actually created. Invalid inputs
  fail; a valid effect for a departed player produces no tokens (CR 800.4).

Contexts expire when the synchronous handler returns. Never retain them,
live pointers or mutable per-game state in a closure/global. Named pure
resolution functions and immutable declarations are the default. Shared
primitive payloads must also be treated as immutable. Linked/deferred state
belongs in game-owned records, not the Go stack.

Target choices, payments and replacements are not ad-hoc card callbacks.
Simultaneous operations must use a batch-capable Engine operation, not a loop
of independently completed removals. Terastodon is the first production example
of correlated batch outcomes; a real graveyard arrival qualifies even if the
commander later moves to command as an SBA (CR 903.9a). Hand/library command-zone
handling remains an optional replacement, not that SBA (CR 903.9b).
Following epic slices deepen links, replacement programs and policy extensions.

## Consumers, errors and portability

`rules.NewEngineWithPrograms(rng, catalog)` installs code before loading a
game. `ValidatePrograms(configs)` reports missing/mismatched code before setup;
`NewGame` also enforces that preflight. Checked registry lookup and deck loading
report program errors as unresolved definitions, not playable partial cards.
Malformed operations fail explicitly. Invocation errors surface through the
existing unsupported-game failure path, never a successful-looking fallback.

The CLI passes the loaded registry's catalog. `sim.Config`, `GoldfishConfig`
and `Matchup` accept an optional immutable catalog; nil uses the production
catalog. Simulator/search Engines share code while retaining independent
contexts and mutable game state. Replay records carry exact program versions;
custom replays may provide a catalog as the third `Replay` argument.

Bindings and authored declaration data round-trip through JSON. Installed Go
functions do not: serializing a Catalog fails explicitly. A consumer needs the
matching compiled catalog to load those references. Version changes must be
intentional; do not reuse a version to pretend different code is the same
implementation. Existing legacy cards and APIs remain available.

Validate through actual stack/choice/clone/parallel/replay scenarios. Metadata,
Go compilation and preservation fingerprints are not rules certification.
