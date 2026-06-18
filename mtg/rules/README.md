# rules package

`mtg/rules` contains the Magic rules engine. It owns turn progression, priority, state-based actions, effect execution, and the game loop that asks agents to choose actions.

The package operates on the core data types from `mtg/game`. The `game` package stores state; `rules` changes that state according to Magic rules.

## Main types

### Engine

`Engine` is the entry point for rules execution:

```go
rng := rand.New(rand.NewPCG(1, 2))
engine := rules.NewEngine(rng)
gameState := engine.NewGame(configs)
result := engine.RunGame(gameState, agents)
```

The engine receives a `*rand.Rand` so simulations and tests can be deterministic. Passing `nil` uses a deterministic default seed.

Use `Engine.NewGame` when you want the engine's RNG to control both initial library shuffles and later in-game randomness.

Use `Engine.RegisterCardImplementation` to attach rules-side hand-written implementations for card definitions whose `CardDef.ImplementationID` is set. A registered implementation owns spell-effect resolution for that card and should mutate state only through `CardContext` helpers so draw logs, events, prevention/replacement hooks, and zone changes stay coherent with declarative effect primitives. Duplicate, empty, nil, or missing implementation registrations panic because those are card implementation bugs rather than legal in-game outcomes.

`RunGame` currently supports opening hands, turn progression, drawing, passing priority, playing lands, mana abilities, spell casting and resolution, common permanent interactions, state-based actions, combat, game termination, and typed game-event emission at key mutation boundaries.

### PlayerAgent

`PlayerAgent` is the interface the engine consumes when it needs a player decision:

```go
type PlayerAgent interface {
	ChooseAction(obs PlayerObservation, legal []action.Action) action.Action
}
```

The interface lives here because `rules.Engine` consumes it. Concrete agents live in `mtg/agent` later.

Agents may also implement `ChoiceAgent` to answer bounded non-action choices such as triggered-ability target selection, ordering simultaneous triggers, or optional "you may" decisions. If an agent does not implement it, the engine records and uses deterministic fallback choices.

### PlayerObservation

`PlayerObservation` is the fog-of-war-safe view passed to an agent. It starts minimal and should grow only as agents need more information.

Do not pass `*game.Game` directly to agents; agents should not see hidden information such as opponents' hands or library order.

### GameResult

`GameResult` is the structured output from a completed game. It records the winner, elimination order, loss reasons, turn count, and per-turn draw/loss/action/choice/resolve/combat-damage/creature-damage/permanent-death logs. `TurnLog.Entries` preserves those per-turn facts in chronological order while the category-specific slices remain available for analytics and tests. The `report` package will consume `[]GameResult` to produce deck analytics.

## Current implementation status

Implemented now:

- `Engine` skeleton and deterministic RNG configuration.
- `Engine.NewGame` for deterministic game setup using the engine RNG.
- `PlayerAgent`, `PlayerObservation`, and result/log data types.
- `ChoiceAgent` support for non-action decisions with deterministic fallback and per-turn choice logging, including trigger targets/order, optional effects, payment choices, resolution/proliferate choices, commander-color mana choices, and scry/surveil decisions.
- Opening hand setup and card drawing.
- Phase helpers for beginning, main, combat, ending, cleanup, and advancing to the next turn, including Saga lore advancement after the draw step and Read ahead entry-chapter choices.
- Extra turn handling in LIFO order, skipping eliminated players.
- Priority loop with multiplayer pass-around-table behavior and stack-aware all-pass handling.
- State-based actions for player elimination from 0 life, lethal poison, lethal commander damage, and failed draws.
- Conservative Commander deck legality validation for 99-card decks plus commander, singleton nonbasic names, simple legendary-creature commanders, and trusted color identity data.
- Permanent state-based actions for 0 toughness, lethal damage, deathtouch damage, 0 planeswalker loyalty, 0 battle defense, completed Sagas, illegal Auras/attachments, token cleanup, legendary-rule duplicates, and +1/+1/-1/-1 counter cancellation.
- Legal action generation for passing, playing lands, casting supported spells from hand, graveyard when permitted, prepared battlefield permanents, or a commander from the command zone with player or permanent targets, casting Morph/Disguise cards face-down, casting Mutate spells targeting owned non-Human creatures, turning face-down permanents face up, activating simple, modal, and choice-bearing mana/equip/loyalty/general abilities from the battlefield or graveyard, Cycling and Ninjutsu from hand, Kicker spell variants, and richer attacker declarations.
- Action application for passing, playing lands, casting supported spells and prepared spell copies, command-zone commander casts with tax/count tracking, Morph/Disguise face-down casts and turn-face-up special actions, activating simple and choice-bearing mana/equip/loyalty/general abilities from the battlefield or graveyard, Cycling and Ninjutsu from hand, paying attack taxes, and declaring attackers.
- Mana cost payment helpers that use pool mana first, then auto-tap untapped basic lands and simple tap mana abilities from mana rocks or non-summoning-sick mana dorks, with generic spell cost modifiers, source-only sacrifice costs, face-down turn-up costs, and attack-tax payment exclusions.
- Stack resolution for creature spells entering the battlefield, face-down permanent spells entering as hidden 2/2 creatures, Mutate spells merging over or under their target while preserving the target object's battlefield state, instant/sorcery spells moving to graveyard, modal spell effects, triggered abilities, equip activated abilities, Cycling abilities, Ninjutsu creatures entering tapped and attacking, Kicker effects, loyalty abilities, delayed triggers, and general non-mana activated abilities. A Mutate spell whose target is illegal on resolution enters as an ordinary creature.
- Effect primitive execution for drawing cards, gaining life, losing life, player damage, permanent damage, destroy, counter, exile, bounce, sacrifice, discard, tap/untap, adding/removing/moving permanent counters, adding poison/energy/experience counters to referenced players, exact-zone referenced-card movement, card- and face-specific bounded cast permissions, mass permanent/player selector effects, token creation/investigate with referenced recipients, supported library search with card-type/supertype filters that lets the searching player choose which matching cards to find or legally fail to find, reveal, discover, linked reveal-to-battlefield flows, shuffle-permanent-into-library, runtime P/T modifiers, declarative runtime continuous and rule effects from battlefield permanents and stack-zone spell abilities, condition-gated effects, referenced damage sources for creature-sourced damage, event-derived/opponent-count/excess-damage dynamic amounts, scry/surveil/mill, fight, transform, phase out, emblems, proliferate, goad, prevention, regeneration, and delayed triggers. Player counter placement emits typed counter-addition events, and poison loss remains a state-based action. Source-scoped uncounterable rule effects prevent only their own spell from being countered. Unsupported effect primitives and unsupported variants are logged instead of silently no-oping.
- Hand-written spell implementation escape hatch through `CardDef.ImplementationID`, `Engine.RegisterCardImplementation`, and `CardContext` mutation helpers.
- Typed `game.Event` emission for spell casts/resolutions, ETB, death, damage, draw, discard, sacrifice, scry, surveil, activated abilities, counter addition, reveal, face-up turns, zone changes, attack/block declarations, and supported beginning-of-step events. Supported draw, life-change, scry, and surveil events record the affected player's event ordinal during the current turn. Permanent zone changes preserve controller, owner, origin/destination, face-down state, and simultaneous-move identity. Sacrifices emit a distinct authoritative event for both resolving effects and paid costs, and preserve simultaneous batches. Scry and surveil each emit one distinct action event, independently from any card movements they cause. Activated-ability events preserve the activating player, source permanent/card, ability index, stack object, and mana-ability classification, but payment-time mana activations do not yet emit them; only non-mana-only activation trigger patterns are accepted and matched. Turned-face-up events use the changed permanent as their trigger subject. Attack declarations preserve their exact player/planeswalker/battle recipient; block and fight events preserve both combatants; and attacker, blocker, fight, and combat-damage events preserve declaration/resolution/pass batch IDs.
- Triggered ability detection from `game.TriggerPattern`, including permanent-zone-change origin/destination and excluded-destination filters, face-down filters, event-subject Selection, attached/departed sources, LKI event-permanent references, and explicit simultaneous one-or-more coalescing. Combat matching uses typed Selections for the attacker/blocker, related combatant, damage source/recipient, and attacked permanent; exact player and recipient relations; combat/noncombat filters; per-attack-target batches; and source/attachment-bound relations. Queued events without a nonzero batch ID are never inferred to be simultaneous. Detection also covers Saga chapter crossings, choice-mediated target selection, APNAP stack placement with choice-mediated same-controller ordering, optional-trigger resolution choices, source-exclusion filters for "another" trigger wording, and structured intervening-if conditions including kicked or cast permanent entry, referenced source live state, event-permanent current/LKI Selection matching and counters, controller-permanent checks, counter-kind absence from event-permanent last-known information, and event-history conditions. Event-history conditions (`Condition.EventHistory`) use `triggerMatchesEvent` over `EventsThisTurn()` or `EventsPreviousTurn()` with the source permanent as the resolution context; a nil source fails closed. Intervening conditions are enforced both when creating and resolving the trigger. Supported beginning-of-step triggers, latched state triggers, implicit Prowess triggers, and resolution through `StackTriggeredAbility` use the same event pipeline.
- Player- and permanent-targeted spell action generation using `TargetSpec` and runtime `game.Target` values, including min/max target counts, mixed target slots, structured target predicates with legacy string fallback, Shroud, Hexproof, Protection, and exact-one target slots chosen by an opponent during spell or ability announcement. Non-controller chooser support currently covers cast and activated ability announcement; triggered-ability stack placement remains on the existing controller-choice path.
- Resolution-time target re-checking with counter-by-rules behavior when all targets become illegal.
- Colorless and X-cost payment, with legal X choices capped for action generation and dynamic X-based effect amounts on resolution.
- Simple sacrifice-as-cost for spells, with sacrificed permanents excluded from mana payment plans.
- Modal spell and non-mana activated-ability support using `Mode`, `ChosenModes`, min/max mode counts, duplicate-mode flags, and mode-specific target validation/resolution.
- Prepare support for permanents that enter prepared: their controller may pay the alternate spell face's cost at its normal timing to cast a copy from the battlefield. Casting the copy unprepares the source without moving it, and the copy resolves through normal spell handling.
- Flash timing for non-instant cards with the Flash keyword, and Flashback-style graveyard casting for cards with a `Flashback` alternative cost.
- Combat step structure, summoning-sickness clearing, single/all attacker choices, multi-blocks, goad and source-scoped must-attack requirements, attack taxes, Flying/Reach/Menace block legality, static attack/block prohibitions, planeswalker and battle attack targets, first strike/double strike damage passes, Trample/Deathtouch combat damage assignment including validated attacker-provided assignments, Lifelink, Toxic, Wither and Infect creature-damage counters, and commander combat damage, Indestructible/regeneration survival from destroy/lethal damage, phasing checks, combat damage to players and permanents, and lethal permanent cleanup.
- Effective characteristic calculation through runtime continuous effects: copy/control/text/type/color/ability/P-T layers, timestamps/dependencies, face-down baseline values, dynamic star P/T, counters, temporary modifiers, static `EffectModifyPT` effects from battlefield permanents, and Mutate's top-component characteristics plus all-component abilities.
- Battlefield zone-change helpers for moving card-backed permanents, tokens, and every component of a merged Mutate permanent to destination zones, detaching attachments, applying per-card commander replacement, applying ETB tapped/counter/payment replacements, clearing Adventure and Suspend state when cards leave exile, and removing tokens from non-battlefield zones as an SBA.
- Deterministic Commander command-zone replacement for supported battlefield, stack, discard, mill, and surveil zone changes.
- Commander mulligan scaffolding with first-free multiplayer mulligan and deterministic bottoming; default simulations currently keep opening seven.
- Aura and Equipment skeleton support with attach/unattach helpers, attach-on-resolution for targeted permanent spells, basic creature-only attachment legality, and illegal attachment/aura SBAs.
- Cleanup-step maximum hand-size discard to seven cards, runtime duration/prevention/regeneration expiry, and marked-damage removal.

Not implemented yet:

- Full attachment legality beyond basic creature-only Aura/Equipment support.
- Agent-driven mulligan decisions, choice-based discard/sacrifice/exile/reveal/tutor decisions, agent-selected replacement ordering, state triggers, copy triggers, and generic APNAP simultaneous choices beyond trigger ordering.
- Escape, Foretell, Evoke, copy-on-stack, cast-without-paying, and choice-rich keyword-action variants beyond the currently supported deterministic primitives.
- Play-vs-cast effects, other nonstandard Saga timing, DFC back-face characteristics, day/night transitions, exile-on-resolution replacements beyond Flashback, unsupported search destinations, and stack-copy effects.

## Game events

Rules helpers append `game.Event` values to `game.Game.Events` as mutations happen. Events are the rules-facing stream for triggered abilities, replacement/prevention effects, analytics, and derived logging. They intentionally differ from `TurnLog` and `GameResult`: logs summarize a completed turn or game for reports, while events describe exact facts at the state-change boundary.

Event data types live in `mtg/game` because card definitions and trigger patterns need to reference the same vocabulary. Emission and later consumers live here in `mtg/rules`.

After state-based actions are checked in the priority loop, the engine consumes unprocessed events from `Game.TriggerEventCursor`, detects matching triggered abilities on battlefield permanents, chooses legal targets through the choice protocol when an agent can answer and deterministic fallback otherwise, and puts those abilities on the stack in APNAP order. Subject-controller and cause-controller filters are matched independently, including for targeted-object events. When one player controls multiple simultaneous triggers, that player can choose their relative order; fallback preserves detection order. Optional triggered abilities still go on the stack and ask the controller whether to apply their effects as they resolve.

## Legal actions

The current engine generates these actions:

- `action.PlayLand(cardID)` for lands in the active player's hand during a main phase when the stack is empty and the land drop is available.
- `action.CastSpell(cardID, targets, xValue, modes)` for supported creature, instant, and sorcery spells. Current cast support covers colored, colorless, generic, and X costs; simple player/permanent targets; choose-one modal spells; Flash timing; simple sacrifice-as-cost; and color-based Protection targeting restrictions.
- `action.CastMutateSpell(cardID, targetID)` for a creature card in hand with a
  typed Mutate cost and an owned non-Human creature target. The over-or-under
  placement is a resolution-time `ChoiceAgent` decision.
- `action.CastFaceDown(cardID, face, kind)` for Morph/Disguise cards in hand during sorcery-speed windows, using the face-down `{3}` alternative cost.
- `action.TurnFaceUp(permanentID)` for controlled face-down permanents whose Morph/Disguise turn-up cost is payable. This is a special action and does not use the stack.
- `action.ActivateAbility(sourceID, abilityIndex, targets, xValue)` and
  `action.ActivateAbilityWithModes(sourceID, abilityIndex, targets, xValue,
  modes)` (or the target-partition-preserving variant generated for ambiguous
  optional modal targets) for simple mana abilities, Equip abilities, Cycling
  and Ninjutsu abilities, and general non-mana activated abilities from the
  battlefield or graveyard. Mana abilities resolve immediately without using
  the stack; Equip, Cycling, and general non-mana abilities use the stack and
  re-check chosen mode-specific targets, including Protection restrictions, on
  resolution.
- `action.DeclareAttackers(attackers)` during the declare attackers turn-based action. Current attack generation is intentionally compact: all eligible attackers attack one alive opponent, or no attackers; attack requirements filter out illegal no-attack choices, and goad also filters out goading-player choices when a legal alternative exists.
- `action.Pass()` for every player with priority.

Legal actions are ordered as play land, cast spell, face-down cast, activate ability, Cycling/Suspend/special actions, then pass so simple agents develop mana before spending it and choose productive actions before passing.

The priority loop treats agent output as untrusted: if an agent returns an action not present in the legal action list, the engine substitutes `Pass`.

When all active players pass in succession, the loop ends the current phase or step only if the stack is empty. If the stack has an object, the engine resolves the top object, resets the pass count, returns priority to the active player, and continues.

## Mana payment

The mana-payment layer supports colored, true colorless, generic, X, hybrid, mono-hybrid, phyrexian, and snow costs. `canPayCost` and `payCost` use current mana pools first, then tap untapped basic lands or simple tap mana abilities controlled by the player while preserving scarce constrained resources such as colored and snow mana. Basic land mana is inferred from the land's name or subtype: Plains for white, Island for blue, Swamp for black, Mountain for red, and Forest for green.

Simple mana abilities are activated abilities marked `IsManaAbility` with no targets, no timing restriction, no loyalty cost, and only add-mana effects. They may be exposed as legal actions for floating mana, or auto-used during cost payment. Creature mana abilities with tap costs respect summoning sickness, and conditional mana abilities use the same activation-condition checks in explicit activation and automatic payment planning.

General activated abilities support mana costs, X costs, tap costs, typed sacrifice/discard/exile additional costs, timing restrictions (`NoTimingRestriction`, sorcery-only, combat, upkeep, and once-per-turn variants), structured activation conditions, modal content with mode-specific target generation, target generation from `TargetSpec`, and stack resolution through the same effect primitives used by spells and triggers. Graveyard activated abilities can exile their own source card as a cost, and if a source leaves its zone while costs are paid, the source card or token definition is preserved on the stack so the ability can still resolve.

Cycling is the initial keyword-action carry-forward slice. Cards in hand with an activated ability carrying the `Cycling` keyword, a mana cost, and a typed discard-this-card additional cost can be activated at instant speed. The engine pays the mana cost, discards the source card with normal discard/zone-change events, puts a `StackActivatedAbility` on the stack with `SourceCardID` preserved, and resolves its effects, usually drawing a card.

Spell cost payment supports typed additional costs, payment choices through the existing `ChoiceAgent` path, phyrexian mana-vs-life choices, minimal spell alternative costs that replace the normal mana cost while preserving required additional costs, Kicker as a combined cost option, and generic cost increases/reductions/set/minimum effects. Deterministic fallback chooses the first valid payment option for agents that do not answer payment choices. Sacrificed permanents and declared attackers are excluded from relevant mana payment plans, so they cannot be used as both mana sources and costs/attackers.

Cost modifiers run after normal/alternative/kicker cost selection and before additional/mana payment planning. Current modifiers cover generic increases, reductions, set values, and minimum generic rules for spells, plus Ghostly Prison-style attack taxes. Ability-specific modifiers and X-value enumeration after reductions remain carry-forward work.

Mana pools empty at phase and step boundaries before later priority windows can use stale mana.

## Payment planner

Cost payment is implemented by the `mtg/rules/payment` package. `rules`
adapts `*game.Game` and rules-only helpers to the payment package through
`rulesPaymentState`; production rules code goes through the package-local
`paymentOrch` seam instead of calling the payment package directly.

Payment request types:

- `payment.SpellRequest` — bundles player, card, source zone, X value, kicker flag, and optional preferences for checking or paying a spell's mana and additional costs.
- `payment.AbilityRequest` — bundles player, ability source permanent, ability definition, X value, and optional preferences for checking or paying an activated-ability cost.
- `payment.GenericRequest` — covers generic mana payments such as attack taxes, Cycling, Ward, Madness, Suspend, and resolution-payment effects.

Payment orchestrator entry points:

- `paymentOrch.canPaySpellCosts(g, req)` — returns true if spell costs are currently payable.
- `paymentOrch.paySpellCosts(g, req)` — pays spell costs and returns additional-cost labels; fails if the plan is stale.
- `paymentOrch.buildAbilityCostPlan(g, req)` — checks activated-ability cost planning without applying it.
- `paymentOrch.payAbilityCosts(g, req)` — pays activated-ability costs, including tap and additional costs.
- `paymentOrch.canPayGenericCost(g, req)` / `paymentOrch.payGenericCost(g, req)` — check or pay generic mana and additional costs.

The payment package keeps payment-plan internals private: mana source discovery,
Convoke/Delve selection, additional-cost matching, and plan application are
implementation details behind `payment.Planner`.

## Combat

Combat follows the real step sequence: beginning of combat, declare attackers, declare blockers, combat damage, and end of combat. The engine initializes `game.CombatState` for the duration of the combat phase, asks the active player to declare attackers, gives players priority in each combat step, applies state-based actions after combat damage, and clears combat state when combat ends.

The current combat implementation supports single-attacker, all-attacker, and no-attacker declaration actions. Goaded and `RuleEffectMustAttack`-affected eligible attackers must attack; goaded attackers prefer non-goading players when such a target is available. Blockers can gang-block a single attacker, and blocker order is recorded for deterministic or attacker-provided damage assignment. Flying attackers can be blocked only by creatures with Flying or Reach, Menace attackers require at least two blockers, and object-bound `RuleEffectCantBeBlocked` effects prevent blockers from being legally declared for the affected attacker.

Unblocked attackers deal effective numeric power as combat damage to their attack target: the defending player, a planeswalker, or a battle. Blocked attackers assign lethal damage through blocker order, with non-trample excess assigned to the last blocker. Trample assigns only lethal damage to blockers and sends the remainder to the attack target; Deathtouch makes 1 damage lethal for assignment, and Deathtouch plus Trample combines accordingly. First Strike and Double Strike use a first-strike combat damage step only when at least one attacker or blocker has First Strike or Double Strike. Lifelink gains life as combat damage is dealt, Wither and Infect sources put -1/-1 counters on creatures instead of marking damage, commander combat damage is tracked for actual commander card instances, and prevented damage does not grant lifelink, mark deathtouch damage, or count as commander damage.

State-based actions destroy creatures with lethal marked damage or 0 effective toughness. Indestructible prevents destroy effects and lethal/deathtouch-damage destruction, but not 0-toughness death; marked damage remains until cleanup. Shield counters prevent damage and replace destruction by removing one shield counter before the permanent moves zones. Effective power and toughness currently include base numeric P/T, +1/+1 and -1/-1 counters, simple until-end-of-turn P/T modifiers, and initial static P/T continuous effects from battlefield permanents. Runtime permanent state also tracks class level and whether a permanent is monstrous; player state tracks Speed and the once-per-turn speed increase guard for Start Your Engines. Card-backed permanents move to their owner's graveyard; tokens move through the destination zone and then cease to exist as an SBA.

This slice intentionally omits combat tricks beyond the existing priority windows and richer attack-tax producers beyond explicit `Game.AttackTaxes`; those carry forward to broader card implementation work.

## Combat in-place module

Combat behavior is concentrated in `combatEngine` (`combat_engine.go`), an in-place module following the same seam pattern as `effectResolver` and `paymentOrchestratorType`. `Engine.runCombatPhase` is a one-liner that constructs `combatEngine{e}` and calls `runPhase`; all combat logic lives on `combatEngine`.

`combatEngine` methods and their responsibilities:

- `runPhase` — full combat-phase sequence: step setup, priority windows, attacker/blocker declaration, first-strike and normal damage passes, mana-pool draining.
- `runPriorityStep` — set step, emit beginning-of-step event, run priority loop, drain mana pools.
- `runPriority` — grant priority to the active player and run the priority loop.
- `declareAttackers` — enumerate legal attacker choices, ask the active player's agent, log and apply the chosen declaration.
- `declareBlockers` — same for each defending player in priority order.
- `legalAttackers` — enumerate legal `DeclareAttackers` actions, including attack requirements, goad target constraints, and attack-tax affordability checks.
- `legalBlockers` — enumerate legal `DeclareBlockers` actions, including Flying/Reach/Menace restrictions.
- `applyAttackers` — validate and apply a `DeclareAttackersAction`: attack-requirement satisfaction, attack-tax payment via `paymentOrch`, tapping non-Vigilance attackers, and recipient-aware attacker-declared event emission with one declaration batch ID.
- `applyBlockers` — validate and apply a `DeclareBlockersAction`: eligibility re-check, Menace count enforcement, blocker-order tracking, and related-combatant event emission with one declaration batch ID across defending players.
- `resolveDamagePass` — assign and mark combat damage for all attackers in one first-strike or normal pass, dispatch to the package-level `resolveBlockedCombatDamage` / `resolveUnblockedCombatDamage` helpers, and batch the pass's damage events.
- `canPayAttackTax` / `payAttackTax` / `attackTaxCost` / `attackingPermanentExclusions` — attack-tax integration through `paymentOrch`.

`Engine` methods `applyDeclareAttackers` and `applyDeclareBlockers` and the package-level functions `legalDeclareAttackersActions` and `legalDeclareBlockersActions` are thin wrappers that preserve the existing call surface used by `actions.go` and tests.

Pure game-state helpers (eligibility predicates, damage computation, goad bookkeeping, blocker ordering) remain as package-level free functions in `combat.go` because they carry no `combatEngine` state and are independently useful.

### Extraction decision criteria

Promote `combatEngine` to a `mtg/rules/combat` subpackage when **all** of the following hold:

1. The subpackage boundary removes a meaningful coupling (the pure helpers in `combat.go` would move with it, or their callers are already isolated).
2. At least one non-combat caller (e.g. a card implementation) needs to import a combat type that is currently unexported. Moving avoids awkward re-exports.
3. The subpackage would have its own tests that are faster to run in isolation than the full `mtg/rules` suite.
4. The interface surface is stable enough that further churn won't force repeated cross-package changes.

Do **not** extract solely to reduce file size or to match the `payment` precedent; `payment` moved because it has a well-defined algorithmic boundary and no direct game-mutation needs. Combat orchestration calls `Engine` methods (`runPriorityLoop`, `applyStateBasedActionsWithLog`) and mutates `*game.Game` directly, so extraction would require passing either an `Engine` interface or callbacks — that interface should be designed only when the payoff is clear.

## Static and continuous effects

The continuous-effect layer derives effective permanent values on demand rather than mutating printed card definitions. Runtime `ContinuousEffect` values cover copy, control, text, type, color, ability, and P/T layers with timestamp/dependency ordering. Battlefield static abilities contribute through source-aware groups and selections or through `AffectedSource` templates bound to the specific source permanent. `StaticAbility.Condition` gates both forms dynamically, including controller checks for permanent types, subtypes, colors, and colorless permanents. If the source leaves the battlefield or a condition becomes false, the next effective-value calculation naturally stops applying the effect.

The layer system still has carry-forward work for richer CDA forms, exact copy/back-face interactions, and performance memoization as card coverage grows.

## Selection matching

`Selection` is defined as pure data in `mtg/game`; the single rules-side interpreter lives in `selection.go` here. `matchSelection(*selectionSubject, *game.Selection)` implements every Selection field semantic exactly once, and the legacy target, controller-controls, trigger, and mass-effect paths route through it via thin adapters rather than re-implementing characteristic checks.

`selectionSubject` is a tagged struct (not an interface) that captures the genuine per-context differences while the field logic stays shared:

- **Kind** (`subjectPermanent`, `subjectEventPermanent`, `subjectCastSpell`) selects the characteristic source: a live permanent's effective/base value set, a triggering event's permanent (including last-known information, the cast card, or a `TokenDef`), or a cast spell's card types, supertypes, subtypes, colors, and mana value. Event-permanent Selection reads live or last-known type, supertype, subtype, color, tapped state, combat state, keyword, power, and toughness, plus printed mana value; this preserves exact departed-subject matching. The `subjectCastSpell` kind reads `event.CardTypes` for type-based predicates, `event.CardSupertypes`/`event.CardSubtypes` for supertype/subtype predicates, `event.Colors` for color predicates, and `event.ManaValue` for mana-value predicates, so cast triggers are matched against characteristics recorded on the `EventSpellCast` at cast time.
- **`clampPower`** distinguishes the target read (power clamped to ≥ 0 and always applicable) from the strict controller-controls read (requires printed power). **`useBase`** forfeits power and toughness, preserving the base-characteristic condition behavior.
- **`controller`/`viewer`** carry controller relativity so `ControllerYou`/`ControllerOpponent` resolve against the correct player (chooser for opponent-chosen targets), and **`sourceObjectID`** drives `ExcludeSource`.

The adapters are: `targetSelection` (targets, `clampPower`), `controllerControlsMatchingSelection` (conditions, base/effective and counting/total-power kept outside the matcher), referenced-condition object matching after `resolveObjectReference`, `triggerSubjectSelection`/`triggerCardSelection` (trigger event subject and cast-spell filters), and `selectorSelection` (mass effects). `selectorSelection` returns fixed package-level `Selection` values so the hot continuous-matching path stays allocation-free, and it returns `ok=false` for the domain selectors (`EquippedCreature`, `AllCreaturesExceptTarget`, `OtherCreaturesDefendingPlayerControls`) whose candidate-domain semantics are expressed by `game.GroupReference` and resolved by the reference resolver's `groupMembers`. The effect-selector path also short-circuits to no match when an `Other...` selector has no source permanent, a divergence from target "another" wording that `ExcludeSource` alone cannot express; the reference resolver preserves that divergence for `ExcludeSource` groups. `selection_parity_test.go` characterizes every legacy `TargetPredicate`, `PermanentFilter`, trigger filter, and `EffectSelector` constant against reference oracles to prove the shared matcher is behavior-preserving, and `reference_resolver_test.go` proves `groupMembers(selector.GroupReference())` matches the legacy mass-effect enumeration for every selector.

## Reference resolution

`referenceResolver` (`references.go`) is the internal module that binds a `*game.Game` and the resolving `*game.StackObject` and owns every runtime reference lookup. It is constructed per resolution by `newReferenceResolver(g, obj)` and exposes:

- `object(game.ObjectReference)` — resolves a target-slot, source, attached, linked, or event object to a live `*game.Permanent` or its last-known-information snapshot (`resolvedObjectReference`).
- `player(game.PlayerReference)` — resolves the controller, a target player, or an object's controller/owner, rejecting eliminated players.
- `permanentAt`/`playerAt` — target-slot and sentinel (`TargetIndexController`, `TargetIndexSourcePermanent`) lookups.
- `groupMembers(game.GroupReference)` — enumerates a group's object IDs in battlefield order, owning candidate-domain enumeration (battlefield, attached object, object-controlled), `Selection` matching, and object-reference exclusions that Selection deliberately keeps outside itself.

The free functions `resolveObjectReference`, `resolvePlayerReference`, `resolvePermanentOrLastKnown`, and `targetPermanentObjectID` are thin adapters that delegate to the module, and `effectResolver.permanentAt`/`playerAt` and the mass-effect enumeration in `selectedPermanentIDsForSelector` route through it. The continuous-effect hot path keeps using `selectorSelection` directly and is intentionally not routed through `GroupReference`, so its allocation behavior is unchanged.

## Targeting result semantics

Target enumeration uses an explicit `targetChoiceResult` struct so callers never infer outcome from nil-slice shape. The `kind` field carries one of four states:

- `targetNoTargetsRequired` — the spell or ability has no target specs; the single legal choice is nil (cast with no targets).
- `targetLegalChoicesFound` — at least one legal target combination exists; `choices` contains one entry per combination. Optional specs with no board candidates produce a single nil choice (no targets selected).
- `targetNoLegalChoices` — specs are present and valid but no legal candidates exist on the current board state.
- `targetInvalidSpec` — a spec has an invalid min/max range (e.g. min > max); `err` describes the problem. This represents a card-definition bug rather than a board-state outcome.

The entry points for action enumeration and trigger target selection consume `targetChoiceResult` directly:

- `targetChoicesForSpell` — resolves target specs for a spell given its card def and chosen modes.
- `targetChoicesForAbilityFromSourceObject` — resolves target specs for an ability with an explicit source object ID (used when the source permanent's identity must be excluded via the "another" predicate).
- `targetChoicesForSpecs` — low-level entry point for tests and special cases.

Callers iterate `result.choices` to produce one legal action per target combination. Action enumeration callers (`legalCastActions`, `legalCommanderCastActions`, `legalActivateAbilityActions`, `firstLegalSpellCastChoice`) check `result.kind` before ranging over `result.choices`: `targetInvalidSpec` is an explicit branch with diagnostic `err` context, although current runtime behavior still skips those actions the same way it skips no-legal-target board states. `triggerTargets` returns `(nil, false)` when the result kind is `targetNoLegalChoices` or `targetInvalidSpec`, keeping those triggers off the stack until a board state with legal targets exists.

## Effect resolver

Instruction resolution is structured around `effectResolver` in `effects.go`. The resolver bundles the per-resolution context (`*Engine`, `*game.Game`, `*game.StackObject`, agents array, `*TurnLog`) so primitive handlers do not repeat those parameters.

`effectResolver` exposes convenience methods used by primitive handlers:

- `quantity(q)` — resolves a fixed or dynamic `game.Quantity`.
- `permanentAt(index)` — looks up a target permanent from the stack object's target list, delegating to the reference resolver module.
- `playerAt(index)` — looks up a target player or the instruction controller, delegating to the reference resolver module.

Resolution follows a two-step call chain:

1. `resolveInstruction` checks instruction conditions and result gates, handles optional instructions, and finds the handler registered for the primitive's sealed `PrimitiveKind`.
2. The typed handler in `primitive_handlers.go` performs the action and returns an `effectResolved` outcome.

`effectResolved` captures whether the instruction was accepted, whether it applied, and any computed amount or excess damage. Named results are written to the stack object so later "if you do" and "that much" instructions observe the actual outcome (CR 608.2c).

Damage helpers apply prevention before mutating life totals, counters, marked damage, combat logs, or damage events. Prevention shields track remaining amount and expire with turn duration. Shield counters prevent the next damage event to that permanent and emit `EventDamagePrevented` instead of `EventDamageDealt`. Color-based Protection, represented by `ProtectionKeyword`, prevents damage from matching colored sources and makes those permanents illegal targets for matching spells and abilities both when chosen and when resolved.

Destroy effects use a pre-zone-change replacement hook. If multiple supported replacement effects apply, the engine records deterministic fallback ordering in `Game.ReplacementDecisions`. Shield counters and regeneration shields can replace destruction; regeneration taps the permanent, removes it from combat, and clears marked damage. Destruction replacement only applies to destroy events; 0 toughness, 0 loyalty, 0 defense, illegal Auras, sacrifice, exile, and bounce still move through their normal zone-change paths.

## Hand-written card implementations

Most cards should be represented declaratively with categorized ability bodies on `CardFace` and typed `game.Primitive` instructions. Cards that need behavior outside the current primitive set may set `CardDef.ImplementationID` and register a matching `CardImplementation` on the `Engine`.

This hook currently covers instant and sorcery spell-effect resolution after normal target re-checking. Permanents, triggered abilities, and activated abilities still resolve through the existing declarative paths. Hand-written implementations receive a `CardContext` instead of `*game.Game`; context methods wrap the same rules helpers used by declarative effects, so custom code participates in draw logging, events, damage prevention, and other mutation-boundary behavior.

## State-based actions

`applyStateBasedActions` loops until stable and panics if state-based actions do not converge. Current checks eliminate players for:

- Life total 0 or less.
- 10 or more poison counters.
- 21 or more commander damage from one commander.
- A failed draw from an empty library (`game.Game.FailedDraws`).

Permanent SBAs handle lethal and deathtouch-marked creature damage, 0 toughness, 0 loyalty planeswalkers, 0 defense battles, illegal Auras, illegal non-Aura attachments, legendary-rule duplicates, +1/+1 and -1/-1 counter cancellation, and tokens ceasing to exist outside the battlefield. Permanent death logs record the permanent object ID, source card ID when present, token name when needed, owner/controller, and death reason.

## Action builder

All `action.Action` values produced inside the `rules` package must be constructed through the package-local `actionBuild` singleton (`actionBuilderType`). `actionBuild` is the only place in the package that calls `mtg/game/action` constructors directly.

Every builder method calls `Action.Validate()` on the newly-constructed action and panics if validation fails — this catches programming errors (e.g. zero card IDs) at the construction site rather than silently emitting invalid actions.

The action package constructors already copy all slice arguments; the builder provides an additional validation layer without duplicating those copies.

## Payment orchestration

All spell and ability payment operations inside production `rules` code must go through the package-local `paymentOrch` singleton (`paymentOrchestratorType`). The orchestrator currently delegates to the package-level payment functions without changing their behaviour; its purpose is to be the production seam for future transactional payment concerns such as rollback, logging, and plan instrumentation. Payment planner unit tests may still exercise lower-level package functions directly, while characterization tests should cover `paymentOrch` so future wrapper behavior is not invisible.

`Engine.applyActionWithChoices` validates incoming actions before applying them. Invalid or hand-built actions that do not match the `action.Action` constructor invariants are rejected instead of being applied with zero-valued payloads.

Methods:

- `paymentOrch.canPaySpellCosts(g, req)` — wraps `canPaySpellCosts`.
- `paymentOrch.paySpellCosts(g, req)` — wraps `paySpellCosts`.
- `paymentOrch.buildSpellCostPlan(g, req)` — wraps `buildSpellCostPlan`.
- `paymentOrch.buildAbilityCostPlan(g, req)` — wraps `buildAbilityCostPlan`.
- `paymentOrch.payAbilityCosts(g, req)` — wraps `payAbilityCosts`.

The request structs (`spellPaymentRequest`, `abilityPaymentRequest`) and the orchestrator remain package-local.

## Effect resolver

All effect-primitive execution routes through the package-local `effectResolver` struct (`effects.go`). `Engine.resolveEffect` and `Engine.resolveEffectWithChoices` construct an `effectResolver` via `newEffectResolver(e, g, obj, agents, log)` and then call `resolver.resolve(effect)`.

`effectResolver` bundles the five parameters that every effect case previously threaded individually:

```go
type effectResolver struct {
    engine *Engine
    game   *game.Game
    obj    *game.StackObject
    agents [game.NumPlayers]PlayerAgent
    log    *TurnLog
}
```

The `resolve` method contains the full resolution body: condition guards, optional/choice/payment handling, the amount-and-result-remembering defer (CR 608.2c), the unsupported-effect log path, selector-driven mass effects, and the per-type switch covering all current primitives.

This struct is the intended seam for a future effect pipeline: middleware, logging wrappers, cached target lookup, or per-effect handler methods can be introduced here without changing `Engine`'s public resolution entry points.

## Mutation boundaries

`mutations.go` defines the package-local helpers that form the intended mutation boundaries for high-churn `game.Game` state changes. New rules code must route through these helpers rather than calling stack, zone, or event primitives directly.

### Stack push helpers

All code that puts an object onto the stack must use one of:

- **`pushSpellToStack(g, obj, castEvent)`** — pushes a spell stack object, emits `EventObjectBecameTarget` for each target, emits the `EventZoneChanged` event pre-built by the caller, then emits `EventSpellCast` from the same event. Used for all spell casts (hand, command zone, graveyard, exile, suspend, cascade, madness). Callers must populate `CardTypes`, `CardSupertypes`, `CardSubtypes`, `Colors`, and `ManaValue` in the `castEvent` using the face-specific `cardTypes(spellDef)`, `cardSupertypes(spellDef)`, `cardSubtypes(spellDef)`, `spellColors(spellDef)`, and stack mana-value helpers so that type-, supertype-, subtype-, color-, and mana-value-filtered cast triggers match correctly; face-down casts explicitly record creature type, no colors, no supertypes/subtypes, and mana value 0. Callers that produce storm copies must call `stormCopyCount` *before* this helper because `EventSpellCast` is emitted inside.

- **`pushAbilityToStack(g, obj)`** — pushes an activated or triggered ability stack object and emits target events. Used for non-mana activated abilities, cycling, and triggered abilities.

Delayed triggered abilities with no targets may call `g.Stack.Push` directly (no target events to emit). Storm copies call `g.Stack.Push` directly because copies are silent (no zone-change or cast events).

### Card lookup helper

**`cardInstanceFaceDef(g, cardID, face)`** retrieves a `*CardInstance` and its `*CardDef` face in one call. Returns `(nil, nil, false)` when the card is absent or has no such face. Use this instead of the two-step `GetCardInstance` + `cardFaceDef` pattern in resolution paths where both the card and its face def are needed.

### Existing mutation helpers

The following helpers in `zones.go`, `events.go`, and `payment_apply.go` predate Phase 4 and follow the same boundary convention:

- `createCardPermanentFace` — moves a card to the battlefield and emits `EventZoneChanged` + `EventPermanentEnteredBattlefield`.
- `removePermanentFromBattlefield` — removes a permanent from `g.Battlefield` without events (callers emit events via `emitPermanentLeaveEvents`).
- `movePermanentToZone` — handles replacement effects, commander zone replacement, detachment, and zone-change events for battlefield exits.
- `discardCardFromHand` — moves a card hand→destination with replacement and commander zone replacement, emits `EventZoneChanged` + `EventCardDiscarded`.
- `moveStackCardToGraveyard` — moves a spell from the stack to its destination zone (respecting Flashback exile replacement) with `EventZoneChanged`.
- `moveExiledCardToGraveyard` — moves a card exile→graveyard with `EventZoneChanged`.
- `setPermanentTapped` — sets tapped state and emits `EventPermanentTapped` or `EventPermanentUntapped`.
- `emitEvent` / `emitZoneChangeEvent` — the only two paths that may append to `g.Events`.

### Conventions

- Emit `EventZoneChanged` before the domain-specific event for the same transition (e.g. emit zone change before `EventSpellCast`).
- Do not append directly to `g.Events` outside `emitEvent`.
- Do not call `g.Stack.Push` outside `pushSpellToStack`, `pushAbilityToStack`, or the two explicit exceptions above (delayed triggers and storm copies).

## Package boundaries

`rules` may import `mtg/game` and `mtg/game/action`. It should keep engine internals unexported unless another package genuinely needs them.

The `game` package must remain pure data and should not import `rules`.
