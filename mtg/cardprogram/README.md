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
- `View().ActivePlayer()` returns the current turn's player identity without
  exposing turn state; like other queries it fails after the context expires.
- `View().Player(reference)` resolves a currently available player;
  `View().Players(group)` returns living members in APNAP order.
- `View().Permanents(group)` returns immutable current public snapshots,
  including ownership and token identity, before a later operation changes them.
  Native bodies can use `CapturedPlayerReference` for a resolved seat; it retains
  that seat rather than substituting its controller/decision maker, and becomes
  unavailable when that player leaves.
- `Execute(primitive)` validates the operation's target scope and dispatches
  the existing Effect Resolver. Its `Outcome` distinguishes unavailable from
  executed-but-ineffective, with copied public result snapshots. It does not
  promise that every legacy scalar is an actual mutation count.
  `Execute(Reveal)` also returns `Outcome.Cards`: minted, exact invocation-local
  handles for the actual revealed library cards, including explicit-card reveals.
  Departed tokens awaiting SBA do not consume a revealed-card count.
- `Ask(request)` uses normal ChoiceAgent observations, validation, logging and
  recorded fallbacks. It does not choose a strategic answer inside card code.
  `ChoiceNumber` uses `NumberRange`, empty `Options`, and one returned scalar
  rather than an option index. An unset maximum does not enumerate possible
  values; the enclosing operation still validates its exact costs and targets.
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
  Identity, quantity and cross-type additions modify one evolving creation
  event; applicability is recomputed after each chosen replacement and inherited
  applied-effect identities prevent self-recursion. Entry choices for every
  resulting token precede all arrivals, so newly entering replacement sources
  cannot affect other members of their own simultaneous entry.
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
  Reveal-only searches with `SearchMustFindIfAvailable` require a fixed positive,
  unrestricted quantity; the minimum is the full count, clamped to availability
  (CR 701.23d), rather than merely one card.
  Every real library-search path opens the same Engine-owned before-find
  opportunity for `StaticAbility{ZoneOfFunction: zone.Library,
  CastWhileSearchingOwnLibrary: true}`. Panglacial Wurm uses this inspectable
  static permission and trample, with no dummy resolving program. A look at the
  top, another player's search, ordinary priority, or an already closed search
  grants no permission. Cast cards leave the library before payment and cannot
  be found; candidates are rebuilt after mana abilities change the library.
  The searching actor pays ordinary announced costs, even when another player
  controls the decisions. Total costs are fixed before manual mana activations;
  only that payer receives mana-ability choices, not ordinary priority.
  Unpaid proposals restore the original library position and reverse reversible
  mana activations. CR 733.1's irreversible draws/library moves/reveals/shuffles
  are retained, including mana consumed by those activations. Cast spells and
  their triggers wait for the enclosing search's normal completion boundary.
  Scope metadata clones by value for observations and supported outer simulation
  boundaries; a mid-search/payment clone is not a saved Go call stack.
  Ordinary actions are rejected there, and attempts to resume priority/turn
  execution report unsupported continuation explicitly.
- `View().TargetCards()` mints exact invocation-local handles for the still-legal
  announced cards in public graveyards or face-up exile. Illegal targets and
  departed incarnations are omitted; hidden-zone cards are not disclosed.
- `MoveCards(placements...)` assigns search/target/revealed handles to hand/graveyard/exile in one
  simultaneous placement, applying zone replacements before committing the moves.
  After replacements, each owner orders their actual simultaneous graveyard
  arrivals under CR 404.3; redirected cards participate only in their real zone.
  Replacement shuffles wait until the full placement finishes. Handles cannot be
  forged, reused across invocations, or followed through a zone change or a later
  shuffle. Target handles also retain their actual source zones. Clone preserves
  the library's shuffle-correlation epoch. For simultaneous battlefield entry,
  programs use the existing `Execute(PutOnBattlefield{Sources: ...})` operation.

`MoveCards` also accepts a battlefield-only placement across owners, using the
same shared simultaneous entry pipeline. Aura admission/attachment choices are
APNAP and non-targeting, against the pre-entry board; host attachment links are
not published during another cohort member's entry replacement. Optional
entry-to-zone replacements make their payment decisions before the cohort
moves, and only actual battlefield arrivals are returned. A library-only
`CardPlacement{Destination: zone.Library, LibraryBottom: true}` instead reorders
still-current revealed library cards in their owners' chosen bottom order,
without inventing a zone change or a new incarnation. Battlefield entry and
bottom reordering each require their own placement, rather than approximating a
mixed-destination batch.

Warp World composes these operations in ordinary Go: snapshot per-owner counts,
move all permanents simultaneously and shuffle, reveal exact cards, enter all
artifact/creature/land cards together, then all remaining enchantments together,
then bottom the remainder. Multi-type cards participate in only the first
applicable stage. Event-time ETB and zone-change source capture prevents
later-stage witnesses from observing earlier events while retaining original
LKI sources and trigger multipliers; triggers still wait for normal priority.

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
Following epic slices deepen policy extensions.

Generous Gift, Pongify, Rapid Hybridization and Resculpt are complete authored
spell registrations. Their single targets and printed costs are ordinary
declarations. Unlike Terastodon, the three destruction spells create a token
for a legal target even when it does not move: they use `Destroy`'s frozen
controller for a departed target without requiring a graveyard arrival. A
surviving target uses its current controller after replacements (CR 608.2h),
including umbra armor destroying a control-changing Aura. Pongify and Rapid
Hybridization set `PreventRegeneration`; shields and umbra armor still apply.
Resculpt captures the target's controller with
`View().Player(ObjectControllerReference(...))` before executing `MovePermanent`
to exile, then uses `CreateTokens` regardless of a replacement's destination.
All four retain normal all-illegal-target rules, token/entry replacements and
CR 111.4 token names. `mtg/rules/removal_token_spells_test.go` covers real casts,
target-type changes, retargeted copies/clones, protection, controller snapshots,
zone/token replacements and commander/token SBAs using the pinned Oracle corpus
and official CR effective 2026-09-25. Each entry preserves its explicit
`Verification` references when registered.

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

## Individual draw replacement programs

Abundance is a complete production registration using
`ReplacementEffect.Program`. This narrow seam currently accepts individual
`EventCardDrawn` actions, not arbitrary event mutation policies. Printed
`MatchEvent`, `ControllerFilter`, source-relative `Condition`, zone filters,
`Optional`, `SelfReplacement` and duration remain inspectable data. Persistent
source abilities are registered from active battlefield permanents and follow
their current controller; phased/departed sources cannot apply.

The Engine owns the pending draw and CR 616 selection/recomputation. Draw
doublers, dig replacements, empty-library wins and dredge compete there;
prohibitions on drawing apply before replacements. Self-replacements have
CR 616.1a precedence. Applied effect identities follow each derived event
(CR 614.5), but sibling events receive independent lineage. A declined optional
action is skipped for the unchanged pending event, not falsely marked applied.
Replacement-created draws still use the same Engine loop and inherited lineage,
never a card-owned loop.
First-draw exemptions are derived from actual draws since the current draw-step
boundary, including separate later spells after Abundance replaced the turn-based
draw. An extra draw step starts a fresh exemption.

The installed immutable program uses the ordinary controlled context.
`View().Source().EventPlayer` and `EventPlayerReference()` identify the affected
player; the source controller remains the effect's controller, not the chooser.
There is no real stack object for this invocation (`StackID` is zero).
Returning without error consumes the pending draw even if no card was found:
decline is decided by the Engine **before** invocation and is not the same result.
Contexts, choices and nested draw lineage expire synchronously and are not
retained by Game clones or immutable catalogs. Missing bindings/unsupported
event categories/mixed legacy and program actions fail explicitly.

Abundance chooses land/nonland only after accepting the replacement.
`Execute(RevealUntil{MatchToDestinationRestChosenBottom: true})` reveals through
the first match or exhaustion, moves only the match through ordinary zone
replacement handling, and offers the shared exact-order choice for the remainder.
Remainder cards stay in the library; reordering them is not a zone change.
Putting the match into hand is not drawing (CR 121.5), so it neither triggers
draw abilities nor causes a failed-draw loss on an empty/nonmatching library.
Every individual draw finishes these operations before the next draw begins.
The corpus pin remains `testdata/card-preservation/corpus.json`; these scenarios
use CR 121, 614.5, 614.11 and 616 from the 2026-09-25 rules.

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

## Source-scoped graveyard play permissions

Muldrotha, the Gravetide is a complete authored production entry using
`RuleEffectPlayFromGraveyardPerPermanentType`, not a resolving program or
card-name exception. Its inspectable `PermissionKey` names an allowance on the
source's current battlefield object. Permission discovery uses effective
abilities, including copied/granted statics and ability removal.

Legal land and cast actions carry a `game.PlayPermission` with source object,
key and the **one** permanent type the player chooses. A multi-type card offers
separate branches; Bestow and face-down casting use the announced spell types,
not the printed types. Agents/Simulator can inspect and select these payloads.
Direct Engine callers use the ordinary choice callback when a selection is
needed. Independent permissions are offered separately without spending this
allowance. Copies of spells are not casts.

`Game.PermanentTypePlayUsesThisTurn` owns successful-use records per source,
player and type; Clone isolates them and every new turn (including extra turns)
resets them. Blink creates a fresh source object, but losing/regaining control
or removing/restoring an ability does not erase the same player's prior usage.
Announcement, costs, timing, cast prohibitions, land-play limits and payment
rollback remain Engine responsibilities. Usage is captured before costs (the
source can leave while paying), committed only after success, and retained if
the spell is countered. Morph/disguise graveyard casts use the existing spell
payment planner, with normal taxes and the announced creature characteristics.

Scenarios use `.cardwork/corpus/oracle-cards.json` and the official rules
effective 2026-09-25 (CR 601, 305, 400.7, 702.37 and 708); registry admission is
not inferred from Oracle compiler recognition.

## Next-turn player decision control

Mindslaver is a complete production entry with its ordinary six-mana casting
cost, four-mana/tap/sacrifice activation, and one player target. The Go body
calls `ControlNextTurn(PlayerReference)`. The Engine captures the resolving
ability's controller and source card/object IDs into a data-only
`game.PlayerControlEffect`, independent of source departure. No player agents,
observations, Game pointers, seat swaps or object-control changes are captured.

The next **actual** turn consumes pending control, including extra turns.
Skipped turns leave it pending; competing pending effects overwrite by creation,
not source-entry timestamp. Control lasts through every cleanup/cleanup-priority
step until the next turn begins. A departing controller or target removes it
without reviving overwritten effects. Clone copies both pending/active records
and scoped search permissions; the Simulator uses the same common dispatch.

`Game.DecisionMaker` is the shared policy for priority, attackers/blockers and
all engine `ChoiceAgent` callbacks (modes, targets, payments, replacement order
and resolving choices). The action actor, payer and object controllers are
unchanged. `PlayerObservation.Player` / `ChoiceRequest.Player` identify the actor;
`PlayerObservation.DecisionMaker`, `ActionLog.DecisionMaker` and
`ChoiceDecision.DecisionMaker` identify the deciding seat. `Hand()` is the actor's
hand, while permission-filtered `HandOf` also exposes the deciding player's own
hand and any players they control. Private top, face-down permanent/spell and
foretell permissions are shared directionally under CR 723.4, never with
uninvolved seats.

Opposition Agent remains a distinct **search-only** override. It can replace
Mindslaver's decision maker during the search, retains its own exile/play
beneficiary, and cannot leak control or library visibility afterward.
`LibraryLookable` exposes the entire searched library, not just matching cards,
only during a real search. Turn control alone grants no library visibility.
SearchContext keeps actor and decision maker separate; determinizations preserve
known hands/top cards and sample unrelated hidden hands/face-down identities.
Publicly designated commander identities are never reassigned.

Concession remains the affected seat's personal action (`Engine.Concede`),
never a controlling agent's legal action. Personal/tournament decisions use
`DecisionPersonal` and are not redirected or used to expose outside-game cards.
The Engine has no outside-game card selection operation.

Acceptance scenarios use `.cardwork/corpus/oracle-cards.json`, current Oracle
rulings retrieved from Scryfall, and official Comprehensive Rules effective
2026-09-25 (CR 723, 800.4a, 614.10 and 513–514). Scripted production activation,
resolution, actual turns/combat/payments/replacements/observers and search/clone
tests, not winning-game smoke results, certify this admission.

## Permanent-spell copies

Resolving a permanent-spell copy uses normal battlefield-entry replacements and
ETB events, but is not token creation (CR 111.13): token quantity, identity and
addition replacements do not apply, and no token-created event is emitted.
