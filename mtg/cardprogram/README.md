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

- `View().Source()` returns captured card/object/stack IDs, controller, X and the
  triggering event's historical player identity (even after that player leaves).
  It does not expose a live permanent or Game pointer.
- `View().Targets()` returns a copy of the declared runtime targets. Rules operations
  still revalidate references; do not use plain Go writes as targeting/mutation.
- `View().Player(reference)` resolves a currently available player;
  `View().Players(group)` returns living members in APNAP order.
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
- `Copy(operation, copier)` reuses stack copying and optional retargeting, returning
  the IDs actually created. Each copier has its own recorded choice, inherits modes,
  X and additional-cost/payment subjects, and owns its spell copies independently
  of later control changes. Copies produce copy events, never cast events.
  Retargeting preserves target counts and may keep illegal targets unchanged
  while selecting legal replacements for other slots (CR 707.10c).
- `Exchange(first, second)` atomically exchanges both live permanents' current
  layer-2 controllers, with copied before/after snapshots and a typed `Exchanged`
  outcome. A missing/phased object, identical object, same controller or departed
  recipient produces no partial change. Successful changes reset sickness and
  remove both permanents from combat.
- `Search(operation)` performs a reveal-only search and returns local `FoundCard`
  incarnation handles. Search options disclose the library only to the authorized
  chooser; controlled searches use the searcher's information permissions.
  Distinct-name searches use the existing staged choices and permit fail-to-find.
- `MoveCards(placements...)` assigns found handles to hand/graveyard in one
  simultaneous placement, applying zone replacements before committing the moves.
  Replacement shuffles wait until the full placement finishes. Handles cannot be
  forged, reused across invocations, or followed through a zone change or a later
  shuffle. Clone preserves the library's shuffle-correlation epoch.

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
Following epic slices deepen replacement programs and policy extensions.

Hive Mind, Gifts Ungiven and Gilded Drake are the next production compositions.
Gilded Drake's `ResolveWithIllegalTargets` declaration is inspectable rule data:
illegal slots are still rescreened, but the sacrifice consequence can resolve
when all targets became illegal. Other triggered abilities retain normal
all-illegal resolution rules.

## Linked and delayed lifetimes

Worldgorger Dragon and Nim Deathmantle are complete production registrations,
not compiler admissions. Their costs, Flying/Trample, Equipment statics, Equip
and trigger filters are ordinary declarations. Oracle input is the pinned corpus
in `testdata/card-preservation/corpus.json`; lifetime scenarios use the official
Comprehensive Rules effective 2026-09-25 (400.7, 607 and 610.3).

- Worldgorger's entry body owns an implementation-qualified
  `AbilityContent.LinkedProducers` channel. `MovePermanent.PublishLinkedProducer`
  accumulates actual exile incarnations under the original permanent ObjectID;
  `OriginalSourceLinkedBattlefieldSource` returns only those exact incarnations
  still in exile, simultaneously and under their owners' control. Copies of the
  same trigger accumulate; a new permanent occurrence owns a different namespace.
  Explicit source-reference exclusion still works after the source leaves and
  does not exclude a later reentry of the same physical card.
- These are CR 607 linked abilities, **not** CR 610.3 "until" effects. If the
  leave trigger resolves before the entry trigger, the latter still exiles.
  Tokens cease to exist at SBA and cannot return as cards. Replaced exiles do
  not publish intended arrivals; replaced returns and moved exiled cards cannot
  be reclaimed through their physical CardIDs.
- Nim's `Execute(Pay)` uses the Engine's optional resolution payment planner;
  only actual successful payment enables return. The death event freezes the
  public graveyard incarnation (400.7e). A normal `PutOnBattlefield` publishes the
  actual arrived permanent into an Engine-owned link; attachment follows that
  ObjectID, never a later physical-card reentry. A missing/blinked Equipment
  cannot attach its new occurrence through the old trigger. The black Zombie
  effect replaces creature subtypes without removing other subtype families.
  Engine attachment legality enforces protection against the Equipment's
  effective characteristics; failure to attach does not undo payment or return.
- `Program.Dependencies` declares handlers created by opaque Go resolution.
  Catalog construction checks exact versions and copies dependency slices.
  `Execute` preflights nested handlers before scheduling; loaded games and
  Simulator mutation seams also validate pending delayed, reflexive and inline
  stack bodies. Missing entries fail explicitly.
- Existing `CreateDelayedTrigger` stores stable program bindings and typed
  captured CardIDs/zone versions, object IDs and historical values in Game state.
  The registered delayed linked-return fixture uses named, noncapturing functions,
  survives source departure and multiple same-turn activations, and proves clone
  independence both before scheduling and with scheduled/on-stack handlers.
  No closure, coroutine, VM or live stack pointer is retained.

The production Worldgorger plus declarative reanimation-fixture test stops at
stable priority seams after several recurrences; it is not an Animate Dead
certification or a runtime iteration cap. No Worldgorger trigger is dropped,
no card-specific cap is installed and no timeout is reported as a rules draw.

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
