# Search-based agent architecture

The `mtg/agent` player is being rebuilt from a greedy, single-action heuristic
scorer into a **determinized Monte Carlo search** agent that uses the rules
engine itself as its forward model. The goal is play that resembles a strong
human Commander player at mid power level, grounded in
`docs/research/COMMANDER-AGENT-PLAYBOOK.md`.

## Context

Today `agent.GenericStrategy` scores each legal action in isolation with coarse
additive constants and picks the argmax. This has hard ceilings:

- **Effects are mostly invisible.** The value IR (`mtg/eval`) reduces ability
  bodies to value atoms, but only ~20 of the engine's ~100 resolution
  primitives are modeled; everything else scores as value-neutral, so pumps,
  keyword/static grants, protection, and most non-obvious effects are ignored.
- **No combat or lethal reasoning.** Blocking ignores the agent's life total and
  commander damage; the agent will not chump-block to survive a lethal hit, and
  over-extends attacks into fatal crackbacks.
- **No sequencing or planning.** There is no notion of "hold interaction," "don't
  be the archenemy," end-step timing, or looking a turn ahead.
- **Choices are naive.** Engine-mediated target/mode choices take the first
  option rather than the best one.

A one-ply scorer cannot fix these structurally: correct Magic decisions depend on
what happens *after* the action (combat resolution, opponents' responses, the
next turn). Strong play requires **looking ahead**.

Two engine facts make lookahead practical:

- **`game.Game.Clone()`** is a deep, test-verified copy (`mtg/game/clone.go`,
  extensive `clone_test.go`). Hypothetical states can be branched cheaply enough
  to search.
- **The engine's priority loop is already a complete forward model.**
  `Engine.runPriorityLoop`/`runTurn` drive the entire game purely through the
  `PlayerAgent` / `ChoiceAgent` callback interfaces. Feeding those callbacks with
  policy agents on a cloned game *is* a simulator — the engine needs no
  re-implementation to roll a position forward.

The binding constraint is a core invariant from ADR 0004: **agents observe only a
fog-of-war `PlayerObservation`, never the true `*game.Game`.** A search agent
fundamentally needs a full state model to simulate, so the architecture must
reconcile search with hidden information.

## Decision

Build a **determinized Monte Carlo search** agent with the rules engine as its
forward model. Five components:

**1. Simulator API (`mtg/rules`).** Expose the existing internal machinery
(`legalActions`, `applyActionWithChoices`, the priority/turn loop) behind a
clean, public, clone-oriented API that operates on a caller-owned `*game.Game`:

- enumerate a player's legal actions in a given state,
- apply one action, resolving any intervening choices via caller-supplied policy
  agents,
- advance to the searching player's next decision point, or roll out to a bounded
  horizon / to game end, driven by policy agents.

The engine remains the single source of rules truth; the agent never
re-implements rules. Simulation always runs on a `Clone()`, never the live game.

**2. Determinization (`mtg/rules`).** The engine constructs, on the searching
agent's behalf, a full `*game.Game` consistent with that player's information:
keep all public zones and the searcher's own hidden zones; **re-sample the parts
hidden from the searcher** (opponents' hands, every library's order) with a
seeded RNG. The agent searches over one or more of these *sampled worlds* and
never inspects the true hidden state, so the fog-of-war invariant holds — the
agent sees plausible sampled information sets, not the truth. Sampling fidelity
improves in phases (naive/known-pool first; belief-model-informed later).

**3. Search agent (`mtg/agent`).** A `PlayerAgent` + `ChoiceAgent` that, at each
decision, draws K determinizations and runs bounded search on each — clone →
apply a candidate action → roll out with policy agents to a horizon → evaluate
the leaf — and aggregates action values across determinizations (Perfect
Information Monte Carlo; Information-Set MCTS later). Reproducible via a per-seat
seeded RNG so simulations stay deterministic.

**4. Leaf / rollout evaluation (`mtg/agent`).** A heuristic position-value
function implementing the playbook §6 factors (board, card advantage, life,
tempo, threat distribution, archenemy paint). It evaluates non-terminal search
leaves and drives fast rollout policies; terminal states use the true engine
result. This is required even for search and is developed as its own phase.

**5. Rollout policy (`mtg/agent`).** A fast heuristic agent (the evolved
`GenericStrategy`) drives all seats during simulation, keeping rollouts cheap and
realistic. Opponent modeling deepens later.

### Search milestones

- **S1 — Simulator + perfect-information search.** Public Simulator API plus
  bounded search from the true state (a single determinization equal to the
  truth), with heuristic leaf evaluation. Yields immediate tactical strength
  (lethal detection, combat, sequencing). It may "see" hidden information; that
  is an accepted engineering milestone, not the end state.
- **S2 — Determinized Monte Carlo (PIMC).** Sample hidden information, search each
  world, aggregate. Removes the hidden-information cheat → realistic play that
  respects what a human could actually know.
- **S3 — ISMCTS, opponent models, budgets, tuning.** Information-set MCTS with
  transposition, node/time budgets, opponent belief models, and eval-weight
  tuning.

### Relationship to the heuristic roadmap (P0–P7)

The heuristic and search tracks converge rather than compete:

- **P0** rich observation + **P0b** effect-model (IR) expansion feed the leaf
  evaluation and the rollout policy (effects must be *visible* to be valued).
- **P1** evaluation function *is* the leaf/rollout evaluation.
- **P2–P5** combat, sequencing/timing, choices, and threat/politics heuristics
  become rollout-policy shortcuts and eval terms (fast, good default play makes
  search cheaper and more accurate).
- **P6** is the search itself (S1–S3).
- **P7** deck/archetype and personality tune the eval weights.

Strong heuristics and search are complementary: better heuristics shrink the
search needed for the same strength, and search covers the tactical cases
heuristics get wrong.

## Key design points

- **Engine as forward model** — no rules logic is duplicated in the agent.
- **Determinization is engine-side** so the "agents never see the true hidden
  state" invariant is preserved; the agent only ever sees sampled worlds.
- **Bounded, budgeted search** — mid-power casual is the target (not cEDH), and
  the WASM playtester constrains memory and time. Search uses shallow horizons,
  node/determinization caps, and clone reuse; it must degrade gracefully.
- **Deterministic & reproducible** — per-seat seeded RNG; a fixed seed reproduces
  a game exactly, so simulations and tests stay stable.
- **Heuristic fallback** — when the search budget is unavailable or exhausted
  (e.g. WASM), the agent falls back to the heuristic policy, which is also the
  rollout policy, so there is always a sane action.

## Consequences

- Enables decisions a one-ply scorer cannot make: chump-to-survive, holding
  interaction for the right window, attacking for lethal, playing around the
  crackback, and combo/timing lines.
- Adds a public Simulator API and determinization to `mtg/rules`, and a larger,
  layered `mtg/agent`. The engine stays authoritative and untouched in its rules.
- Performance and memory cost is the main risk, especially under WASM. Mitigate
  with shallow search, budgets, clone pooling, and the heuristic fallback; measure
  before widening.
- Hidden-information realism arrives with S2; S1 briefly plays with perfect
  information as a stepping stone.

## Implementation order

1. This ADR.
2. **P0 / P0b** — rich observation and effect-model IR expansion (feed eval and
   rollouts; also the "pipe full information, not narrow fields" convention).
3. **P1** — heuristic evaluation / position-value function (leaf eval + rollout
   core).
4. **S1** — Simulator API on clones + perfect-information bounded search + leaf
   eval; wire an optional search agent into `sim`/playtester behind a flag.
5. **P2–P5** — combat, sequencing/timing, choices, threat/politics heuristics
   (improve eval and rollout policy).
6. **S2** — determinization + PIMC.
7. **S3 / P7** — ISMCTS, opponent models, budgets, and deck/personality weight
   tuning.
