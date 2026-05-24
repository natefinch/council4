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
- (deferred) Combat damage prevention and replacement effects.
- [x] Commander combat damage tracking from commander permanents.
- (deferred) Attacking planeswalkers and battles.
- [x] Basic attack requirements and goad target preference.
- (deferred) Attacker-chosen damage assignment order, protection, regeneration, attack taxes, and other prevention/replacement behavior.

## Phase 7 — Permanent interaction and richer state-based actions

- [x] Runtime permanent targets and target legality checks.
- [x] Destroy, exile, bounce, sacrifice, tap/untap, and damage-to-permanent effect primitives.
- [x] Board wipes and mass-effect execution.
- [x] Token creation and token cleanup semantics.
- [x] +1/+1 and -1/-1 counters, including counter cancellation.
- [x] P/T modifications from +1/+1 and -1/-1 counters for combat and lethal-damage SBAs.
- [x] Aura and Equipment skeleton: attach/unattach helpers, attach-on-resolution for targeted permanent spells, basic creature-only legality, and illegal attachment/aura SBAs.
- (deferred) Equip actions and richer attachment legality beyond the implemented Aura/Equipment skeleton.
- [x] Maximum hand size and cleanup discard.
- [x] More state-based actions: 0 toughness, lethal damage, legendary rule, +1/+1/-1/-1 counter cancellation, illegal attachments, aura legality, planeswalker loyalty, battle defense.
- [x] (deferred from Phase 6) Attacking planeswalkers and battles.
- (deferred) Regeneration and other destruction replacement behavior (deferred from Phase 6).

## Phase 8 — Mana, casting, and costs

- [x] Mana abilities as actions, including mana dorks and mana rocks.
- [x] Multicolor, colorless, generic, and variable mana costs.
- (deferred) Hybrid, phyrexian, and snow mana costs.
- [x] X spells and X-cost choice handling.
- [x] Simple sacrifice-as-cost for spells.
- (deferred) Alternative costs, cost reductions/increases, richer additional costs, and cost-choice UI.
- [ ] (deferred from Phase 6) Attack taxes and attack cost payments.
- [x] Modal spells and mode selection for choose-one modal spell abilities.
- [x] Equip actions using activated ability actions and stack resolution.
- (deferred) Kicker, Flashback, Madness, Escape, Foretell, Cycling, Morph/Disguise, and other common non-combat keyword actions.
- (deferred) Richer attachment legality beyond the basic Aura/Equipment skeleton (deferred from Phase 7).
- [x] Flash and instant-speed timing support for non-instant cards.
- [x] (completed in Phase 7) Legal target re-checking on resolution and counter-by-rules for all-targets-illegal spells.

## Phase 9 — Abilities, events, and effects architecture

- [ ] Event system for game events: cast, resolve, ETB, death, damage, attack, block, draw, discard, zone changes.
- [ ] Triggered ability detection, trigger ordering, and stack placement.
- [ ] General activated ability action generation and resolution beyond Phase 8 mana abilities and basic Equip.
- [ ] Static abilities and continuous effect support.
- [ ] Replacement and prevention effects.
- [ ] (deferred from Phase 6) Combat damage prevention and replacement effects.
- [ ] (deferred from Phase 6) Protection restrictions and prevention behavior.
- [ ] (deferred from Phase 8) Alternative costs, cost reductions/increases, richer additional costs, and attack cost/tax framework.
- [ ] (deferred from Phase 8) Kicker, Flashback, Madness, Escape, Foretell, Cycling, Morph/Disguise, and other non-combat keyword actions.
- [ ] Continuous effect layer system, including characteristic-defining abilities and dynamic star P/T.
- [ ] Turn-duration effects and cleanup expiry.
- [ ] Choice framework for may choices, mode choices, ordering, scry/surveil, discard, sacrifice, tutor, and reveal decisions.
- [ ] Hand-written card implementation escape hatch behind the same card implementation interface.

## Phase 10 — Commander format rules

- [ ] Deck legality checks: 100 cards, singleton, commander legality, and color identity.
- [ ] Commander zone replacement for zone changes.
- [ ] Casting commanders from the command zone.
- [ ] Commander tax and commander cast-count tracking.
- [ ] Commander damage from each commander to each player.
- [ ] Commander mulligan flow, including multiplayer first-mulligan behavior.
- [ ] Multiplayer draw rules, seating order, and eliminated-player cleanup hardening.
- [ ] Optional bracket/power-level metadata for simulations and reports.

## Phase 11 — Card data, decklists, and card implementations

- [ ] `mtg/cards` registry package mapping canonical card names to card definitions.
- [ ] Scryfall bulk data ingestion as the source of truth for card metadata.
- [ ] Generated `CardDef` data for supported cards.
- [ ] `mtg/deck` package for Moxfield/MTGO-style text decklist parsing.
- [ ] Commander section parsing (`// Commander`, `COMMANDER:`) and explicit four-deck input.
- [ ] Unsupported-card reporting with actionable messages.
- [ ] Declarative card implementation schema built from effect primitives.
- [ ] LLM-assisted build-time generation pipeline for declarative card implementations from oracle text.
- [ ] Validation suite for generated card implementations.
- [ ] Initial supported-card corpus based on common Commander staples and test decks.

## Phase 12 — Agent and observation system

- [ ] Rich `PlayerObservation` with own hand, public zones, battlefield, stack, life totals, commander state, known information, and legal actions.
- [ ] Hidden-information boundaries: agents never see opponents' hands or library order.
- [ ] Stateful agent hooks for observing actions and maintaining known information.
- [ ] Strategy interface for scoring legal actions.
- [ ] Generic rule-based Commander strategy using board presence, card advantage, mana efficiency, threat removal, and survival.
- [ ] Deck pre-analysis: tags, mana curve, commander profile, archetype classification, and power-level estimate.
- [ ] Threat assessment and target selection for multiplayer.
- [ ] Combat attack/block heuristics beyond `FirstLegal`.
- [ ] (deferred from Phase 6) Attacker-chosen combat damage assignment order.
- [ ] Mana planning and sequencing heuristics.
- [ ] Stack interaction heuristics for removal and counterspells.
- [ ] Personality/skill knobs: aggression, risk tolerance, politics weight, noise, and archetype bias.
- [ ] Random/baseline agent for comparison.
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
- [ ] (deferred from Phase 8) Hybrid, phyrexian, and snow mana costs after the payment model tracks choice, life payment, and snow provenance.
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
