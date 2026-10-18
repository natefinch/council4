# Authored Go card programs

`game.Mode.Program` is a data-only `ProgramReference{ID, Version}`. Costs,
targets, triggers and mode selection stay in the ordinary Card Definition.
The Engine resolves the selected mode synchronously through an installed
immutable `cardprogram.Catalog`, using the same rules as declarative cards.
No parser, interpreter, coroutine or copied Go call stack is involved.

## Capability lookup

Start with the rules concept, then inspect the declaration and its behavioral
test. This index includes **declarative** capabilities as well as native Go
operations; needing no Go program is a valid implementation. Follow the
[authoring workflow](../../docs/card-authoring.md) and read the relevant official
rules before designing the card. A row is navigation, not a promise that every
Oracle variant in that family is supported.

| Rules concept / wording | Existing hook | Working example / behavioral evidence | Boundary to check |
| --- | --- | --- | --- |
| Printed costs, targets, modes, timing | [`AbilityContent`, `Mode`, `TargetSpec`](../game/ability.go), ordinary activated/static declarations | [copy spells](../cards/authored/copy_spells.go), [paid spell scenarios](../rules/authored_copy_spells_test.go) | Announcement and target revalidation belong to the Engine, not a resolving callback. |
| "Whenever you commit a crime" | [`EventCrimeCommitted`, `TriggerPlayerYou`](../game/event.go), ordinary triggered abilities | [crime event scenarios](../rules/commit_crime_test.go), [Marchesa scenarios](../rules/marchesa_dealer_of_death_test.go) | A crime occurs when a spell or ability is put on the stack, once per announcement. Include spells, activated and triggered abilities; copies and later target changes don't create or undo the crime. |
| "That player controls" on a trigger | [`Selection.ControlledByEventPlayer`](../game/selection.go) | [Garland scenarios](../rules/garland_royal_kidnapper_test.go), [damaged-player targeting](../rules/damaged_player_target_test.go) | Keep the ability controller as chooser; match and recheck the captured event player, not any opponent. |
| "Defending player controls" | [`Selection.ControlledByDefendingPlayer`](../game/selection.go) | [Spear of Bashenga](../rules/the_spear_of_bashenga_test.go) | Attack events distinguish the attacker from the defending player. |
| Intervening "if", distinct card types, event history | [`TriggerCondition.InterveningCondition`](../game/ability.go), [`Condition`](../game/condition.go) | [intervening-if tests](../rules/triggers_interveningif_test.go) | Conditions that govern triggering must also recheck on resolution. Inspect the complete aggregate/selection declarations. |
| Resolve an event player or current target controller | [`EventPlayerReference`, `ObjectControllerReference`](../game/reference.go), `View().Player` | [Nim Deathmantle](../cards/authored/nim_deathmantle.go), [lifetime tests](../rules/authored_lifetime_test.go) | Historical identity, current controller, owner, payer and decision maker are different concepts. |
| Optional resolving payment, including mana and life | [`Pay`, `ResolutionPayment`](../game/primitive.go) through `Execute` | [Nim Deathmantle](../cards/authored/nim_deathmantle.go), [resolving-payment tests](../rules/resolution_payment_energy_test.go) | Use one Engine payment for a combined cost; ask a choice only after legality/affordability is established. |
| "Look at the top N; put one in hand and the rest in your graveyard" | [`Dig`](../game/primitive.go), gated by `Pay` when payment is optional | [dig scenarios](../rules/effects_dig_test.go), [Marchesa scenarios](../rules/marchesa_dealer_of_death_test.go) | `Dig` keeps the look private unless `Reveal` is set, chooses among available cards, and routes any unchosen cards to the graveyard; a short library is handled by the shared operation. |
| Destroy, then act on what actually moved | `Context.Destroy`, [`Departure.Cards`](context.go) | [Terastodon](../cards/authored/terastodon.go), [arrival composition](../rules/card_program_arrivals_test.go) | Use actual destination and minted public-arrival handles. Protected creatures, tokens and unrelated graveyard cards are not returns. |
| "Choose a card put into a graveyard this way" | `Destroy` receipts, `Ask`, `MoveCards` | [paid destruction/choice/reanimation composition and clones](../rules/card_program_arrivals_test.go) | Choices use normal Engine requests; do not turn arbitrary snapshots into movement authority. |
| Return actual cards with a controller/type/color/counter modification | [`CardPlacement`](context.go) with `Recipient`, `ContinuousEffects`, `EntryCounters`, `EntryTapped` | [modified simultaneous entry](../rules/card_program_arrivals_test.go) | One modified battlefield placement contains the entire simultaneous batch; all handles must belong to this invocation. Entry changes use the shared pre-entry path. |
| Public graveyard/exile targets | `View().TargetCards()`, `MoveCards` | [numeric target spells](../cards/authored/numeric_targets.go), [real target scenarios](../rules/numeric_target_cards_test.go) | Handles are exact incarnations. A card that leaves and returns is not the old target. |
| Revealed or private library search | `Context.Search` for revealed handles; declarative [`Search`](../game/primitive.go) for private placement | [qualified/private tutors](../cards/authored/production_tutors.go), [private-search scenarios](../rules/production_private_tutors_test.go) | Do not reveal a private find to obtain a handle. Unrestricted mandatory finds differ from qualified fail-to-find. |
| Copy a spell/ability or choose new targets | `Context.Copy`, [`CopyStackObject`](../game/primitive.go) | [copy spells](../cards/authored/copy_spells.go), [copy scenarios](../rules/authored_copy_spells_test.go) | Copies are not casts; preserve modes, X, illegal targets that may be retained, and recorded choices. |
| Create tokens together, including for multiple players | `CreateTokens` and [`CreateToken`](../game/primitive.go) | [removal/token spells](../cards/authored/removal_token_spells.go), [replacement scenarios](../rules/removal_token_spells_test.go) | Token/entry replacements and simultaneous cohorts are Engine-owned; do not loop independently completed batches. |
| Exchange control | `Context.Exchange` | [Gilded Drake](../cards/authored/gilded_drake.go), [choice-program scenarios](../rules/authored_choices_program_test.go) | Exchange is all-or-nothing and uses current effective controllers. |
| "Exiled with this" across different activations | [`MoveCard.PublishLinkedProducer`, `AbilityContent.LinkedProducers`](../game/primitive.go), original-source linked battlefield sources | [Worldgorger Dragon](../cards/authored/worldgorger_dragon.go), [linked-producer scenarios](../rules/linked_producer_test.go) | State is owned by the original source occurrence, not a mutable program closure or physical-card ID. |
| Return only matching linked exiles with entry changes | [`PutOnBattlefield.Selection`](../game/primitive.go), original-source sources, existing `ContinuousEffects`/`Recipient`/`EntryCounters` | [Ghost Vacuum operation composition](../rules/linked_producer_ghost_vacuum_test.go) | Filter each current linked card; do not use a whole-instruction gate as a per-card filter. Only noncontextual card predicates are supported here; unreturned cards retain their exact link. |
| "When it leaves" versus "until it leaves" | Existing [`LinkedKey`, `LinkedProducerKey`](../game/primitive.go), delayed declarations | [Worldgorger and Nim](../rules/authored_lifetime_test.go), [delayed programs](../rules/authored_delayed_program_test.go) | CR 607 linked triggers and CR 610.3 durations have different order/lifetime behavior. |
| "Creatures lose abilities", base P/T, colors or types | [`StaticAbility`](../game/ability_body.go), [`ContinuousEffect`](../game/continuous.go), `ApplyContinuous` | [Humility/Opalescence](../cards/authored/humility_opalescence.go), [layer scenarios](../rules/humility_opalescence_test.go) | Static groups stay live; resolving characteristic effects snapshot their recipients. Check timestamps, dependencies and additive versus replacing fields. |
| "Until your next turn", unblockability, goad | [`ApplyRule`, `Goad`](../game/primitive.go), [`EffectDuration`](../game/duration.go) | [goad attack tests](../rules/goaded_subject_attack_test.go) | Goad and can't-be-blocked are not ordinary granted abilities. Duration is tied to the creating effect's controller. |
| Replace an individual draw | [`ReplacementEffect.Program`](../game/replacement.go) with inspectable applicability | [Abundance](../cards/authored/abundance.go), [draw replacement tests](../rules/draw_replacement_program_test.go) | This program seam is for individual draws, not arbitrary replacement categories; preserve lineage and first-actual-draw rules. |
| Spell self-exile / return to hand / shuffle away | [`MoveResolvingSpell`](../game/primitive.go) | [stack resolution](../rules/stack.go), [self-shuffle scenarios](../rules/self_shuffle_into_library_test.go) | Use the resolving-spell operation, not `MoveCard` pretending a popped resolving spell is in a player zone. Copies have no physical card to move. |
| Cast while searching / graveyard play permission | Inspectable static permissions | [Panglacial Wurm](../cards/authored/panglacial_wurm.go), [Muldrotha](../cards/authored/muldrotha.go), [search scenarios](../rules/panglacial_wurm_test.go) | Authorization, mana-payment windows and usage records are not ordinary priority or card-owned state. |
| Commander choices, replacements, identity and simultaneous moves | [shared zone transition machinery](../rules/README.md#shared-zone-transitions) via Engine operations | [cross-operation movement contracts](../rules/zone_movement_contract_test.go), [lifecycle scenarios](../rules/zone_movement_lifecycle_test.go) | No card-specific commander logic or immediate SBA inside resolution. Library reorders are not zone changes. |
| Explicit identity/program/evidence admission | [`cardset.Entry`](../cards/cardset/cardset.go), checked registry | [verification admission tests](../cards/verification_registry_test.go) | New authored identities need explicit card-specific `Verification`; a string alone does not prove rules correctness. |

If a concept is not obvious here, use symbol lookup on the linked type, read its
**complete** declaration, then follow the Engine implementation and tests.
Do not infer absence from one truncated read or a guessed identifier. Before
declaring a blocker, check for an existing canonical issue and establish the
actual missing behavior as described in the authoring guide.

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

**Mana abilities require inspectable declarative instructions.** A
`ManaAbility.Content` containing a Go program is unsupported, even when the
program is installed and would execute `AddMana`. Card-definition validation,
checked registry lookup and Engine preflight reject that representation with an
explicit diagnostic. This includes declared granted, copied, token and delayed
ability templates; runtime token/grant operations validate introduced
program-bearing declarations before mutation. Use the existing tap-for-mana
helpers or supported `ManaAbility` instruction sequences instead. The Engine
does not run opaque Go speculatively to classify mana abilities or put them on
the stack as a workaround.

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
  Each member's `Cards` mints placement authority for its actual still-current
  graveyard/face-up-exile arrivals. Filter by the actual `Destination`, then
  compose recorded choices and `MoveCards`; never rebuild handles from raw
  snapshot IDs. Hidden-zone redirects and departed tokens mint no such authority.
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
  Actual public graveyard/exile arrival receipts are minted for another
  placement in the same invocation, including replacement redirects. Old
  handles cannot follow those cards through a later zone change.

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

A single battlefield `CardPlacement` may set `Recipient`, `ContinuousEffects`,
`EntryCounters` and `EntryTapped` for its entire simultaneous batch. Without a
recipient, the owners control their respective arrivals. These are ordinary
shared-entry changes, not a callback after entry; modified entries cannot be
split into multiple independent placement batches. Invalid non-battlefield
options or malformed entry declarations fail before mutation. Returned
battlefield snapshots describe arrivals, not reusable public-zone handles.

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
Card programs do not implement these commander checks or invoke SBAs themselves;
the shared movement/entry operations record actual arrivals and the Engine
processes them at the proper checkpoint. See the
[shared transition contract](../rules/README.md#shared-zone-transitions).
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
contexts and mutable game state. Replay records carry program IDs/versions,
including transitive delayed-handler dependencies; missing IDs or wrong versions
fail explicitly. Custom replays may provide a catalog as the third `Replay`
argument. Exact replay requires the same code revision, inputs and configuration:
these references are compatibility checks, not hashes of executable behavior.

Program references and authored declaration data round-trip through JSON. Installed Go
functions do not: serializing a Catalog fails explicitly. A consumer needs the
matching compiled catalog to load those references. Version changes must be
intentional; do not reuse a version to pretend different code is the same
implementation. Existing legacy cards and APIs remain available.

Validate through actual stack/choice/clone/parallel/replay scenarios under the
[repository policy](../../.github/copilot-instructions.md#verify-proportionally).
Metadata and Go compilation are not rules certification.

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
Entry preserves announced X and copied additional/alternative cost decisions
(including kicker) independently of actual casting. `HasSpellAnnouncement` and
`EnterHasSpellAnnouncement` include spell copies; `EnteredFromCast`/`EnterWasCast`,
cast origin and actual mana spending do not. Ordinary copies of battlefield
permanents do not inherit historical X or paid-cost decisions.

## Event-time trigger capture

Rules-emitted events capture matching ordinary trigger bodies, source incarnation,
controller, entry choices and applicable multipliers before later instructions can
change the battlefield. Choices of modes/targets, APNAP ordering and stack
placement still wait until priority; intervening-if conditions are checked at
trigger time and again on resolution.
ETB suppression is decided at that same capture boundary, using the post-event
board (including a simultaneous entering cohort). A later suppressor cannot erase
an existing trigger, and removing a suppressor cannot resurrect a prevented one.
Suppression precedes multiplication; only uncaptured synthetic events use live
compatibility filtering at placement.
Both the explicit permanent-entered event and the legacy battlefield zone-change
event use this same suppression and ETB-multiplier policy.

The capture policy covers every emitted event kind, not an ETB/damage-only
allowlist. Previously late-detected families include cast/resolved/copied spells,
life changes, counters, tap/untap, discard/cycling/reveal/land and exile plays,
blockers/blocked/unblocked attackers, scry/surveil, activation/mana production,
prevention/destruction replacement, face-up/mutate/phasing/Class/monstrous,
token creation/library search, and monarch/blessing/initiative/dungeon events.
Color-dependent legacy tapped-for-mana declarations observe the completed mana
production event, when its colors are known.

Cycled-card and cast-spell self triggers, graveyard step sources and explicit
leave/dies look-back exceptions use the shared source matcher. Damage captures
active ability sources before damage consequences; source-characteristic LKI does
not restore abilities that left before the damage. Library-to-graveyard
self triggers and inherent monarch abilities retain their dedicated capture.
Saga chapters preserve their classification. State, madness, delayed, reflexive,
room and mana-spend-rider queues retain their separate priority-boundary handling.
Raw synthetic events appended outside rules emission retain a compatibility scan;
they are not a substitute for emitting real gameplay events.
