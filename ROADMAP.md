# Council4 Roadmap

Council4's product goal is to accept four Commander decklists, run repeated AI-controlled test games with those decks, and produce a report about the individual deck being tested: win rate, finishing position, game length, card performance, mana development, tempo, and common failure modes.

Use this file as the project-level feature checklist. Check items off as they land, and keep package READMEs plus `CONTEXT.md` vocabulary in sync when implementation names or domain terms change.

## Current foundation

- [x] Four-player Commander game state with players, zones, battlefield, stack, turn order, and combat state.
- [x] Deterministic setup and seeded engine RNG.
- [x] Opening hands and per-turn draw.
- [x] Turn progression through beginning, main, combat, ending, and cleanup.
- [x] Multiplayer priority loop with pass handling.
- [x] Basic land play and land-per-turn tracking.
- [x] Simple mana payment using basic lands.
- [x] Simple creature, instant, and sorcery casting through the stack.
- [x] Runtime player targets for simple targeted spells.
- [x] Effect primitives for draw, gain life, lose life, and player damage.
- [x] State-based player elimination for 0 life, poison, commander damage, and failed draws.
- [x] Combat step structure, attacker declarations, blocker declarations, combat damage to players/creatures, and lethal creature damage cleanup.
- [x] CLI hardcoded `land`, `spells`, and `combat` modes for rules-engine smoke testing.

## Phase 6 — Complete core combat

- [x] Multi-blocking with deterministic blocker order.
- [x] Evasion and blocking restrictions: Flying, Reach, and Menace.
- [x] First Strike and Double Strike combat damage steps.
- [x] Trample damage assignment to defending players.
- [x] Deathtouch lethal-damage assignment rules.
- [x] Lifelink life gain from combat damage.
- [x] Indestructible survival from destroy effects and lethal-damage SBAs.
- [x] (completed in Phase 9) Initial combat damage prevention and replacement effects. Full replacement/prevention ordering is deferred to Phase 9C.
- [x] Commander combat damage tracking from commander permanents.
- [x] (completed in Phase 7) Attacking planeswalkers and battles.
- [x] Basic attack requirements and goad target preference.
- [x] (completed in Phase 9C) Attacker-chosen damage assignment order/division, regeneration, attack taxes, and broader prevention/replacement behavior. Protection and the initial prevention slice were completed in Phase 9.

## Phase 7 — Permanent interaction and richer state-based actions

- [x] Runtime permanent targets and target legality checks.
- [x] Destroy, exile, bounce, sacrifice, tap/untap, and damage-to-permanent effect primitives.
- [x] Board wipes and mass-effect execution.
- [x] Token creation and token cleanup semantics.
- [x] +1/+1 and -1/-1 counters, including counter cancellation.
- [x] P/T modifications from +1/+1 and -1/-1 counters for combat and lethal-damage SBAs.
- [x] Aura and Equipment skeleton: attach/unattach helpers, attach-on-resolution for targeted permanent spells, basic creature-only legality, and illegal attachment/aura SBAs.
- (deferred to Phase 9C) Richer attachment legality beyond the implemented Aura/Equipment skeleton. Basic Equip actions were completed in Phase 8.
- [x] Maximum hand size and cleanup discard.
- [x] More state-based actions: 0 toughness, lethal damage, legendary rule, +1/+1/-1/-1 counter cancellation, illegal attachments, aura legality, planeswalker loyalty, battle defense.
- [x] (deferred from Phase 6) Attacking planeswalkers and battles.
- [x] (completed in Phase 9C) Regeneration and other destruction replacement behavior (deferred from Phase 6).

## Phase 8 — Mana, casting, and costs

- [x] Mana abilities as actions, including mana dorks and mana rocks.
- [x] Multicolor, colorless, generic, and variable mana costs.
- [x] (completed in Phase 9B) Hybrid, phyrexian, and snow mana costs.
- [x] X spells and X-cost choice handling.
- [x] Simple sacrifice-as-cost for spells.
- [x] (completed in Phase 9B) Alternative costs, richer additional costs, and cost-choice UI. Cost reductions/increases were completed in Phase 9C.
- [x] (completed in Phase 9C; from Phase 6) Attack taxes and attack cost payments.
- [x] Modal spells and mode selection for choose-one modal spell abilities.
- [x] Equip actions using activated ability actions and stack resolution.
- [x] Kicker, Flashback, Madness, Morph/Disguise, Cycling, Suspend, Convoke, Delve, Ward, Storm, Cascade, and a broad first slice of non-combat keyword actions. Carry-forward: Escape, Foretell, Evoke, copy-on-stack, cast-without-paying, and richer choice-heavy variants.
- (deferred to Phase 9C) Richer attachment legality beyond the basic Aura/Equipment skeleton (deferred from Phase 7).
- [x] Flash and instant-speed timing support for non-instant cards.
- [x] (completed in Phase 7) Legal target re-checking on resolution and counter-by-rules for all-targets-illegal spells.

## Phase 9 — Abilities, events, and effects architecture

- [x] Event system for game events: cast, resolve, ETB, death, damage, attack, block, draw, discard, reveal, face-up turns, and zone changes.
- [x] Triggered ability detection, trigger ordering, and stack placement.
- [x] General activated ability action generation and resolution beyond Phase 8 mana abilities and basic Equip.
- [x] Initial static abilities and continuous P/T effect support.
- [x] Initial replacement and prevention effects: shield-counter damage prevention and destroy replacement.
- [x] (deferred from Phase 6) Initial combat damage prevention/replacement through shield counters and color-based Protection.
- [x] (deferred from Phase 6) Initial Protection restrictions and prevention behavior for protection from colors.
- [x] (completed in Phase 9B; originally from Phase 8) Alternative costs, richer additional costs, and payment-choice framework. Cost reductions/increases were completed in Phase 9C.
- [x] Initial keyword/action carry-forward slice: Cycling as a hand-zone activated ability with discard-as-cost and draw-on-resolution.
- [x] (completed across Phase 9C and follow-up work; originally from Phase 8) Kicker, Flashback, Madness, Morph/Disguise, Suspend, Convoke, Delve, Ward, Storm, Cascade, and many non-combat keyword/action primitives. Carry-forward: Escape, Foretell, Evoke, copy-on-stack, cast-without-paying, and richer choice-heavy variants.
- [x] (completed in Phase 9C) Continuous effect layer system, including characteristic-defining abilities and dynamic star P/T.
- [x] (completed in Phase 9C) Turn-duration effects and cleanup expiry.
- [x] Initial choice framework for trigger targets, same-controller trigger ordering, and optional triggered effects.
- [x] (partially completed in Phase 9C) Richer choice framework for scry/surveil plus deterministic first slices for discard, supported search, and reveal effects. Carry-forward: agent-driven discard/sacrifice/exile/reveal/tutor choices, richer search destinations, and other non-action choices.
- [x] Hand-written card implementation escape hatch behind the same card implementation interface.

## Phase 9B — Costs and payment architecture

This phase makes cost payment a first-class rules subsystem before Commander rules and broad card implementation work depend on it. The focus is reusable infrastructure plus representative vertical slices, not every cost-related mechanic.

- [x] Concrete payment vocabulary: payment plans, mana units, symbol payments, additional-cost selections, life payments, and payment choices.
- [x] Provenance-aware mana pool that preserves existing simple color-count APIs while tracking at least snow mana.
- [x] Incremental symbol-level mana planner for colored, generic, colorless, X, hybrid, mono-hybrid, snow, and phyrexian symbols.
- [x] Typed additional-cost data for common costs such as sacrifice, discard, pay life, exile, reveal, and tap costs; migrate current string-based sacrifice/Cycling costs.
- [x] Payment choice plumbing through the existing choice framework, with deterministic fallback for agents that do not answer payment choices.
- [x] Minimal alternative-cost vertical slice where an alternative cost replaces the normal mana cost and can include additional costs.
- [x] Cost-modifier seam for reductions, increases, and taxes without a speculative full modifier/layer pipeline; real producers landed in Phase 9C.
- [x] Attack-cost/tax design seam; implement a real Ghostly Prison-style slice only after static cost modifiers have a real producer.
- [x] Documentation updates for `CONTEXT.md`, package READMEs, and roadmap carry-forward notes.

## Phase 9C — Non-Commander gameplay rules completion

This phase closes major gameplay-rule gaps that are not Commander-specific before card data generation and broad simulations depend on them. It is a dependency-ordered umbrella of implementation slices, not a single small feature. It is based on the Magic Comprehensive Rules effective 2026-04-17, especially CR 510, 514, 603, 606-616, 701-702, 704, 707-714, 723-724, and 731, plus the card text parsing guide.

- [x] Effective characteristics foundation: runtime continuous effects, copyable values, copy effects, control/type/subtype/supertype/color/text/ability/keyword changes, layer ordering with timestamps/dependencies, face-down baseline values, and dynamic star P/T.
- [x] Runtime durations and cleanup expiry: until end of turn, this turn, until your next turn, cleanup damage/removal expiry, delayed triggers, and next-end-step scheduling. Carry-forward: richer next-time replacement consumption.
- [x] Last-known information and linked ability infrastructure for battlefield exits, dies/LTB trigger type matching, delayed source identity, and paired exile/return effects. Carry-forward: pruning stale LKI/linked records and exact exile object identity across repeated zone changes.
- [x] Replacement/prevention engine slice: deterministic replacement-order records, prevention shields, regeneration shields, ETB tapped/counter replacements, draw-step skip effects, and replacement-aware damage/destroy events. Carry-forward: agent-selected CR 616 ordering and broader as-enters choices.
- [x] Combat choices and cleanup hardening: single-attacker choices, attacker-provided blocker damage division with order/trample/deathtouch validation, attack taxes through payment planning, regeneration removal from combat, phasing checks, and eliminated-player combat/stack/permanent cleanup.
- [x] Real cost modifiers and taxes through the Phase 9B seam: generic reductions/increases/set/minimum rules, Ghostly Prison-style attack taxes, and split second action restriction. Carry-forward: ability cost modifiers, X enumeration after reductions, and "can't be countered" for counter effects.
- [x] Expanded choice framework slice: scry/surveil choices through `ChoiceAgent` with deterministic fallback and logging, plus mill. Follow-up keyword-action work added deterministic discard, library-to-hand search with optional reveal/shuffle, and reveal events. Carry-forward: agent-driven tutor/search, discard/sacrifice/exile/reveal choices, modal variants beyond choose-one, full top/bottom ordering payloads, and generic APNAP simultaneous choices.
- [x] Special action/card-form slice: planeswalker loyalty abilities, emblems with ability data, transform and phase-out primitives, phase-in during untap, Suspend setup, and Morph/Disguise face-down/turn-face-up actions. Carry-forward: Foretell setup, cast-from-zone/play-vs-cast permissions, exile-on-resolution replacement, Sagas, DFC back faces, day/night, and richer attachment legality.
- [x] Keyword-action infrastructure slices: Kicker/if-kicked hooks, fight, scry, surveil, mill, transform, proliferate, goad, counter, discard, supported search, reveal, and investigate primitives. Carry-forward: Escape, Foretell, Evoke, copy-on-stack, cast-without-paying, "can't be countered", richer search variants, and richer choice-backed keyword actions.
- [x] Trigger hardening slice: delayed triggers, next-end-step scheduling, intervening-if checks at trigger and resolution, dies/LTB LKI matching, cast triggers, and APNAP/same-controller ordering choices. Carry-forward: general state triggers, copy triggers, delayed-trigger intervening-if data, and string-condition parsing.
- [x] Scenario/unit fixtures for representative 9C slices plus `CONTEXT.md`, package README, and roadmap updates.

## Phase 10 — Commander format rules

- [x] Conservative deck legality checks: 99-card deck plus commander, singleton nonbasic names, simple legendary-creature commander legality, and trusted `CardDef.ColorIdentity` subset checks. Carry-forward: partners/backgrounds, "any number of" singleton exceptions, and computed color identity from card data.
- [x] Commander zone replacement for battlefield, stack, hand-discard, mill, and surveil zone changes. Carry-forward: owner choice instead of deterministic command-zone replacement.
- [x] Casting commanders from the command zone using explicit cast source zones.
- [x] Commander tax and commander cast-count tracking through the cost/payment seam.
- [x] Commander damage from each original commander card instance to each player, including stolen commanders and excluding token/copy object IDs.
- [x] Commander mulligan scaffolding, including multiplayer first-free mulligan and deterministic bottoming. Carry-forward: real agent mulligan decisions and non-draw mulligan event semantics.
- [ ] Multiplayer draw-rule and seating-order hardening beyond the current clockwise `TurnOrder` and Phase 9C eliminated-player cleanup.
- [x] Optional bracket/power-level metadata pass-through for simulations and reports.

## Phase 11 — Card data foundation

- [x] `mtg/cards` registry package mapping canonical card names to card definitions.
- [x] Scryfall Oracle Cards bulk data as the source of truth for Card Definition metadata and Oracle text.
- [x] Deterministic Oracle recognition, typed lowering, validation, and Go source rendering in `cardgen`.
- [x] Isolated Card Generation tooling under `cardgen/`, including the Oracle compiler and `gencardlist`.

See [`CARD_FEATURES_ROADMAP.md`](./CARD_FEATURES_ROADMAP.md) for the detailed card-text feature coverage roadmap that feeds generated card implementation work.

## Phase 11B — Decklists and broad card implementation rollout

- [x] Compile the full Scryfall corpus with source-spanned unsupported diagnostics and collision-safe deterministic output.
- [x] Validate assembled Card Definitions with `game.ValidateCardDef` before rendering and compile generated packages as a corpus gate.
- [x] Remove the superseded LLM scaffold generator, agent rollout skills, and manifest/worklist/report workflow so Card Generation has one path.
- [x] `mtg/deck` runtime package for Moxfield/MTGO-style decklist parsing used by simulations.
- [x] Explicit four-deck input model and Commander deck loading path that turns card names into validated `CardDef` references.
- [ ] Continue conservative Oracle parser and typed lowering coverage using the numbered compiler-expansion checklist.
- [ ] Generated card implementations should target the current keyword/action infrastructure for Kicker, Flashback, Madness, Morph/Disguise, Suspend, Convoke, Delve, Ward, Storm, Cascade, counter, discard, supported search/reveal, proliferate, goad, investigate, and related primitives. Carry-forward: Escape, Foretell, Evoke, copy-on-stack, cast-without-paying, and richer choice-backed variants.
- [ ] Broaden the supported-card corpus through frequency-driven compiler expansion and inspect every newly generated card at each breakpoint.
- [ ] Add runtime smoke validation for compiled Card Implementations, beyond structural Card Definition validation, so sample casts and activations catch unsupported or illegal behavior.

## Phase 12 — Agent and observation system

- [x] Rich `PlayerObservation` with own hand, public zones, battlefield, stack, life totals, commander state, known information, and legal actions.
- [x] Hidden-information boundaries: agents never see opponents' hands or library order.
- [x] Stateful agent hooks for observing actions and maintaining known information.
- [x] Strategy interface for scoring legal actions.
- [x] Generic rule-based Commander strategy using board presence, card advantage, mana efficiency, threat removal, and survival.
- [x] Deck pre-analysis: tags, mana curve, commander profile, archetype classification, and power-level estimate.
- [x] Threat assessment and target selection for multiplayer.
- [x] Combat attack/block heuristics beyond `FirstLegal`.
- [x] Combat and non-action choice heuristics for Phase 9C gameplay choices, after the rules engine exposes them in observations.
- [x] Mana planning and sequencing heuristics.
- [x] Stack interaction heuristics for removal and counterspells.
- [x] Personality/skill knobs: aggression, risk tolerance, politics weight, noise, and archetype bias.
- [x] Random/baseline agent for comparison.
- [ ] Future: IS-MCTS with determinization, game cloning, and configurable simulation budgets.
- [ ] Future: optional LLM-driven agent for qualitative experiments, not default simulations.

## Phase 13 — Simulation harness

- [ ] `mtg/sim` package for running repeated games with the same four decklists.
- [ ] CLI accepts four decklist paths and identifies the deck being tested.
- [ ] Configurable game count, seed, worker count, agent profile, and output paths.
- [ ] Per-game deterministic seed derivation from a master seed.
- [ ] Parallel execution across CPU cores.
- [ ] Structured `SimulationResult` aggregate.
- [ ] Replay/debug support: store seed plus action history for any game.
- [ ] Failure capture for panics, unsupported cards, and illegal action regressions.
- [ ] Smoke fixtures for known small decklists.

## Phase 14 — Reporting and analytics

- [ ] `report` package consuming `[]rules.GameResult` / `SimulationResult`.
- [ ] Text summary to stdout.
- [ ] Detailed JSON report file.
- [ ] Win rate and average finishing position.
- [ ] Game length distribution and turns-to-win/turns-to-lose.
- [ ] Per-card draw, cast, resolve, and zone-change frequency.
- [ ] Per-card performance: cards seen in wins vs. losses, cards stranded in hand, cards frequently discarded/removed.
- [ ] Mana curve analysis: lands played per turn, mana available, mana spent, missed land drops.
- [ ] Land flood and land screw indicators.
- [ ] Expensive-card rotting-in-hand indicators.
- [ ] Tempo analysis: turn the deck comes online, board presence over time, damage clock.
- [ ] Commander cast-count distribution and commander dependency indicators.
- [ ] Opponent interaction analysis: removal aimed at tested deck, countered spells, board wipes.
- [ ] Report fixtures/golden tests for stable output.

## Phase 15 — Rules conformance and quality hardening

- [ ] Game cloning for tests, replay, and future MCTS.
- [ ] Scenario fixture format for concise rules regression tests.
- [ ] Golden tests for representative Commander staples.
- [ ] Property/fuzz tests for zone moves, target legality, priority convergence, and SBA convergence.
- [ ] Comprehensive smart-priority tests so skipped priority never hides legal responses.
- [ ] Hardening and property tests for Phase 9C cost modifiers, continuous effects, durations, replacement/prevention ordering, and combat choices.
- [ ] Performance benchmarks for per-game runtime and simulation throughput.
- [ ] Determinism tests for fixed seeds and parallel simulation.
- [ ] Error model for unsupported cards and unsupported mechanics.
- [ ] Documentation for current rule coverage and known limitations.

## Source notes

This roadmap is based on:

- `CONTEXT.md` for project vocabulary and relationships.
- `docs/adr/0001-hybrid-declarative-card-implementations.md` for card implementation architecture.
- `docs/adr/0002-smart-priority-skips-empty-passes.md` for priority performance strategy.
- `docs/adr/0003-design-decisions-session-2026-05-22.md` for product goal, v1 scope, deck input, metrics, and simulation volume.
- `docs/adr/0004-engine-architecture.md` for package layout and dependency direction.
- `docs/research/card-game-ai-research.md` for agent, simulation harness, parallelism, replay, cloning, and MCTS roadmap.
- `docs/research/COMMANDER-AGENT-PLAYBOOK.md` and `docs/research/COMMANDER-STRATEGY.md` for Commander agent heuristics, archetypes, threat assessment, mulligans, politics, and combat decisions.
- `docs/research/CARD-TEXT-PARSING.md` for card implementation generation, ability kinds, costs, targets, replacement/prevention, layers, zones, and keyword coverage.
