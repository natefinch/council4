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

`GameResult` is the structured output from a completed game. It records the winner, elimination order, loss reasons, turn count, and per-turn draw/loss/action/choice/resolve/combat-damage/creature-damage/permanent-death logs. The `report` package will consume `[]GameResult` to produce deck analytics.

## Current implementation status

Implemented now:

- `Engine` skeleton and deterministic RNG configuration.
- `Engine.NewGame` for deterministic game setup using the engine RNG.
- `PlayerAgent`, `PlayerObservation`, and result/log data types.
- `ChoiceAgent` support for non-action decisions with deterministic fallback and per-turn choice logging, including trigger targets/order, optional effects, payment choices, and scry/surveil decisions.
- Opening hand setup and card drawing.
- Phase helpers for beginning, main, combat, ending, cleanup, and advancing to the next turn.
- Extra turn handling in LIFO order, skipping eliminated players.
- Priority loop with multiplayer pass-around-table behavior and stack-aware all-pass handling.
- State-based actions for player elimination from 0 life, lethal poison, lethal commander damage, and failed draws.
- Permanent state-based actions for 0 toughness, lethal damage, deathtouch damage, 0 planeswalker loyalty, 0 battle defense, illegal Auras/attachments, token cleanup, legendary-rule duplicates, and +1/+1/-1/-1 counter cancellation.
- Legal action generation for passing, playing lands, casting supported spells with player or permanent targets, activating simple mana/equip/loyalty/general abilities, Cycling from hand, Kicker spell variants, and richer attacker declarations.
- Action application for passing, playing lands, casting supported spells, activating simple mana/equip/loyalty/general abilities, Cycling from hand, paying attack taxes, and declaring attackers.
- Mana cost payment helpers that use pool mana first, then auto-tap untapped basic lands and simple tap mana abilities from mana rocks or non-summoning-sick mana dorks, with generic spell cost modifiers and attack-tax payment exclusions.
- Stack resolution for creature spells entering the battlefield, instant/sorcery spells moving to graveyard, modal spell effects, triggered abilities, equip activated abilities, Cycling abilities, Kicker effects, loyalty abilities, delayed triggers, and general non-mana activated abilities.
- Effect primitive execution for drawing cards, gaining life, losing life, player damage, permanent damage, destroy, exile, bounce, sacrifice, tap/untap, mass selector effects, token creation, runtime P/T modifiers, scry/surveil/mill, fight, transform, phase out, emblems, prevention, regeneration, and delayed triggers.
- Hand-written spell implementation escape hatch through `CardDef.ImplementationID`, `Engine.RegisterCardImplementation`, and `CardContext` mutation helpers.
- Typed `game.GameEvent` emission for spell casts/resolutions, ETB, death, damage, draw, discard, zone changes, and attack/block declarations.
- Triggered ability detection from `game.TriggerPattern`, choice-mediated target selection, APNAP stack placement with choice-mediated same-controller ordering, optional-trigger resolution choices, and resolution through `StackTriggeredAbility`.
- Player- and permanent-targeted spell action generation using `TargetSpec` and runtime `game.Target` values.
- Resolution-time target re-checking with counter-by-rules behavior when all targets become illegal.
- Colorless and X-cost payment, with legal X choices capped for action generation.
- Simple sacrifice-as-cost for spells, with sacrificed permanents excluded from mana payment plans.
- Choose-one modal spell support using `Mode`, `ChosenModes`, and mode-specific target validation/resolution.
- Flash timing for non-instant cards with the Flash keyword.
- Combat step structure, summoning-sickness clearing, single/all attacker choices, multi-blocks, goad attack requirements, attack taxes, Flying/Reach/Menace block legality, planeswalker and battle attack targets, first strike/double strike damage passes, Trample/Deathtouch combat damage assignment including validated attacker-provided assignments, Lifelink and commander combat damage, Indestructible/regeneration survival from destroy/lethal damage, phasing checks, combat damage to players and permanents, and lethal permanent cleanup.
- Effective characteristic calculation through runtime continuous effects: copy/control/text/type/color/ability/P-T layers, timestamps/dependencies, face-down baseline values, dynamic star P/T, counters, temporary modifiers, and static `EffectModifyPT` effects from battlefield permanents.
- Battlefield zone-change helpers for moving card-backed permanents and tokens to destination zones, detaching attachments, and removing tokens from non-battlefield zones as an SBA.
- Aura and Equipment skeleton support with attach/unattach helpers, attach-on-resolution for targeted permanent spells, basic creature-only attachment legality, and illegal attachment/aura SBAs.
- Cleanup-step maximum hand-size discard to seven cards, runtime duration/prevention/regeneration expiry, and marked-damage removal.

Not implemented yet:

- Full attachment legality beyond basic creature-only Aura/Equipment support.
- Mulligans, choice-based discard/sacrifice/exile/reveal/tutor decisions, agent-selected replacement ordering, state triggers, copy triggers, and generic APNAP simultaneous choices beyond trigger ordering.
- Flashback, Madness, Escape, Foretell, Morph/Disguise, Suspend, Evoke, Convoke, Delve, Ward, Prowess, proliferate, goad, copy-on-stack, cast-without-paying, and other non-combat keyword actions beyond Flash, basic Equip, Cycling, Kicker, scry/surveil/mill, fight, and transform.
- Cast-from-zone permissions, play-vs-cast effects, face-up special actions, Saga chapters, DFC back-face characteristics, day/night transitions, exile-on-resolution replacements, and full "can't be countered" support once counter effects exist.

## Game events

Rules helpers append `game.GameEvent` values to `game.Game.Events` as mutations happen. Events are the rules-facing stream for triggered abilities, replacement/prevention effects, analytics, and derived logging. They intentionally differ from `TurnLog` and `GameResult`: logs summarize a completed turn or game for reports, while events describe exact facts at the state-change boundary.

Event data types live in `mtg/game` because card definitions and trigger patterns need to reference the same vocabulary. Emission and later consumers live here in `mtg/rules`.

After state-based actions are checked in the priority loop, the engine consumes unprocessed events from `Game.TriggerEventCursor`, detects matching triggered abilities on battlefield permanents, chooses legal targets through the choice protocol when an agent can answer and deterministic fallback otherwise, and puts those abilities on the stack in APNAP order. When one player controls multiple simultaneous triggers, that player can choose their relative order; fallback preserves detection order. Optional triggered abilities still go on the stack and ask the controller whether to apply their effects as they resolve.

## Legal actions

The current engine generates these actions:

- `action.PlayLand(cardID)` for lands in the active player's hand during a main phase when the stack is empty and the land drop is available.
- `action.CastSpell(cardID, targets, xValue, modes)` for supported creature, instant, and sorcery spells. Current cast support covers colored, colorless, generic, and X costs; simple player/permanent targets; choose-one modal spells; Flash timing; simple sacrifice-as-cost; and color-based Protection targeting restrictions.
- `action.ActivateAbility(sourceID, abilityIndex, targets, xValue)` for simple mana abilities, Equip abilities, Cycling abilities, and general non-mana activated abilities. Mana abilities resolve immediately without using the stack; Equip, Cycling, and general non-mana abilities use the stack and re-check targets, including Protection restrictions, on resolution.
- `action.DeclareAttackers(attackers)` during the declare attackers turn-based action. Current attack generation is intentionally compact: all eligible attackers attack one alive opponent, or no attackers; goad filters out illegal no-attack and goading-player choices when a goaded creature can attack.
- `action.Pass()` for every player with priority.

Legal actions are ordered as play land, cast spell, activate ability, then pass so simple agents develop mana before spending it and choose productive actions before passing.

The priority loop treats agent output as untrusted: if an agent returns an action not present in the legal action list, the engine substitutes `Pass`.

When all active players pass in succession, the loop ends the current phase or step only if the stack is empty. If the stack has an object, the engine resolves the top object, resets the pass count, returns priority to the active player, and continues.

## Mana payment

The mana-payment layer supports colored, true colorless, generic, X, hybrid, mono-hybrid, phyrexian, and snow costs. `canPayCost` and `payCost` use current mana pools first, then tap untapped basic lands or simple tap mana abilities controlled by the player while preserving scarce constrained resources such as colored and snow mana. Basic land mana is inferred from the land's name or subtype: Plains for white, Island for blue, Swamp for black, Mountain for red, and Forest for green.

Simple mana abilities are activated abilities marked `IsManaAbility` with no targets, no timing restriction, no loyalty cost, and only add-mana effects. They may be exposed as legal actions for floating mana, or auto-used during cost payment. Creature mana abilities with tap costs respect summoning sickness.

General activated abilities support mana costs, X costs, tap costs, typed sacrifice/discard additional costs, timing restrictions (`NoTimingRestriction`, sorcery-only, combat, upkeep, and once-per-turn variants), target generation from `TargetSpec`, and stack resolution through the same effect primitives used by spells and triggers. If a source leaves the battlefield while paying a sacrifice cost, its card or token definition is preserved on the stack so the ability can still resolve.

Cycling is the initial keyword-action carry-forward slice. Cards in hand with an activated ability carrying the `Cycling` keyword, a mana cost, and a typed discard-this-card additional cost can be activated at instant speed. The engine pays the mana cost, discards the source card with normal discard/zone-change events, puts a `StackActivatedAbility` on the stack with `SourceCardID` preserved, and resolves its effects, usually drawing a card.

Spell cost payment supports typed additional costs, payment choices through the existing `ChoiceAgent` path, phyrexian mana-vs-life choices, minimal spell alternative costs that replace the normal mana cost while preserving required additional costs, Kicker as a combined cost option, and generic cost increases/reductions/set/minimum effects. Deterministic fallback chooses the first valid payment option for agents that do not answer payment choices. Sacrificed permanents and declared attackers are excluded from relevant mana payment plans, so they cannot be used as both mana sources and costs/attackers.

Cost modifiers run after normal/alternative/kicker cost selection and before additional/mana payment planning. Current modifiers cover generic increases, reductions, set values, and minimum generic rules for spells, plus Ghostly Prison-style attack taxes. Ability-specific modifiers and X-value enumeration after reductions remain carry-forward work.

Mana pools empty at phase and step boundaries before later priority windows can use stale mana.

## Combat

Combat follows the real step sequence: beginning of combat, declare attackers, declare blockers, combat damage, and end of combat. The engine initializes `game.CombatState` for the duration of the combat phase, asks the active player to declare attackers, gives players priority in each combat step, applies state-based actions after combat damage, and clears combat state when combat ends.

The current combat implementation supports single-attacker, all-attacker, and no-attacker declaration actions. Goaded eligible attackers must attack and prefer non-goading players when such a target is available. Blockers can gang-block a single attacker, and blocker order is recorded for deterministic or attacker-provided damage assignment. Flying attackers can be blocked only by creatures with Flying or Reach, and Menace attackers require at least two blockers.

Unblocked attackers deal effective numeric power as combat damage to their attack target: the defending player, a planeswalker, or a battle. Blocked attackers assign lethal damage through blocker order, with non-trample excess assigned to the last blocker. Trample assigns only lethal damage to blockers and sends the remainder to the attack target; Deathtouch makes 1 damage lethal for assignment, and Deathtouch plus Trample combines accordingly. First Strike and Double Strike use a first-strike combat damage step only when at least one attacker or blocker has First Strike or Double Strike. Lifelink gains life as combat damage is dealt, commander combat damage is tracked for actual commander card instances, and prevented damage does not grant lifelink, mark deathtouch damage, or count as commander damage.

State-based actions destroy creatures with lethal marked damage or 0 effective toughness. Indestructible prevents destroy effects and lethal/deathtouch-damage destruction, but not 0-toughness death; marked damage remains until cleanup. Shield counters prevent damage and replace destruction by removing one shield counter before the permanent moves zones. Effective power and toughness currently include base numeric P/T, +1/+1 and -1/-1 counters, simple until-end-of-turn P/T modifiers, and initial static P/T continuous effects from battlefield permanents. Card-backed permanents move to their owner's graveyard; tokens move through the destination zone and then cease to exist as an SBA.

This slice intentionally omits combat tricks beyond the existing priority windows and richer attack-tax producers beyond explicit `Game.AttackTaxes`; those carry forward to broader card implementation work.

## Static and continuous effects

The continuous-effect layer derives effective permanent values on demand rather than mutating printed card definitions. Runtime `ContinuousEffect` values cover copy, control, text, type, color, ability, and P/T layers with timestamp/dependency ordering. Battlefield static abilities with `EffectModifyPT` contribute to matching permanents through source-aware selectors such as `EffectSelectorCreaturesYouControl` and `EffectSelectorOtherCreaturesYouControl`; if the source leaves the battlefield, the next effective-value calculation naturally stops applying the effect.

The layer system still has carry-forward work for richer CDA forms, exact copy/back-face interactions, and performance memoization as card coverage grows.

## Replacement and prevention

Damage helpers apply prevention before mutating life totals, counters, marked damage, combat logs, or damage events. Prevention shields track remaining amount and expire with turn duration. Shield counters prevent the next damage event to that permanent and emit `EventDamagePrevented` instead of `EventDamageDealt`. Color-based Protection, represented by `AbilityDef.ProtectionFromColors`, prevents damage from matching colored sources and makes those permanents illegal targets for matching spells and abilities both when chosen and when resolved.

Destroy effects use a pre-zone-change replacement hook. If multiple supported replacement effects apply, the engine records deterministic fallback ordering in `Game.ReplacementDecisions`. Shield counters and regeneration shields can replace destruction; regeneration taps the permanent, removes it from combat, and clears marked damage. Destruction replacement only applies to destroy events; 0 toughness, 0 loyalty, 0 defense, illegal Auras, sacrifice, exile, and bounce still move through their normal zone-change paths.

## Hand-written card implementations

Most cards should be represented declaratively with `game.AbilityDef` and `game.Effect` primitives. Cards that need behavior outside the current primitive set may set `CardDef.ImplementationID` and register a matching `CardImplementation` on the `Engine`.

This hook currently covers instant and sorcery spell-effect resolution after normal target re-checking. Permanents, triggered abilities, and activated abilities still resolve through the existing declarative paths. Hand-written implementations receive a `CardContext` instead of `*game.Game`; context methods wrap the same rules helpers used by declarative effects, so custom code participates in draw logging, events, damage prevention, and other mutation-boundary behavior.

## State-based actions

`applyStateBasedActions` loops until stable and panics if state-based actions do not converge. Current checks eliminate players for:

- Life total 0 or less.
- 10 or more poison counters.
- 21 or more commander damage from one commander.
- A failed draw from an empty library (`game.Game.FailedDraws`).

Permanent SBAs handle lethal and deathtouch-marked creature damage, 0 toughness, 0 loyalty planeswalkers, 0 defense battles, illegal Auras, illegal non-Aura attachments, legendary-rule duplicates, +1/+1 and -1/-1 counter cancellation, and tokens ceasing to exist outside the battlefield. Permanent death logs record the permanent object ID, source card ID when present, token name when needed, owner/controller, and death reason.

## Package boundaries

`rules` may import `mtg/game` and `mtg/game/action`. It should keep engine internals unexported unless another package genuinely needs them.

The `game` package must remain pure data and should not import `rules`.
