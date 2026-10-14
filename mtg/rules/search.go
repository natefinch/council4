package rules

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// SearchAgent is an optional agent capability (see
// docs/adr/0011-search-based-agent-architecture.md): an agent that decides a
// priority action by searching the game tree instead of scoring the observation
// directly. When the engine detects it, at a priority decision it calls
// ChooseActionBySearch with a SearchContext — a forward model of the current
// decision — instead of ChooseAction. An agent that does not implement it is
// unaffected. Combat declarations use the same action dispatch; non-action
// choices still go through ChoiceAgent.
type SearchAgent interface {
	// ChooseActionBySearch returns the action to take from legal, having looked
	// ahead with ctx. It must return one of the legal actions; the engine
	// validates the result and falls back to passing otherwise.
	ChooseActionBySearch(ctx SearchContext, legal []action.Action) action.Action
}

// DefaultSearchDecisionLimit bounds hypothetical per-branch decisions, separate
// from live work and Searcher's candidate-width/lookahead policy.
const DefaultSearchDecisionLimit = 10000

// SearchContext gives a SearchAgent a forward model of the current decision
// without exposing the live game. It is created by the engine and passed to
// ChooseActionBySearch. The agent branches on states from Determinize, rolls them
// forward with Simulator, and evaluates the results; it never mutates the live
// game.
type SearchContext struct {
	engine        *Engine
	game          *game.Game
	player        game.PlayerID
	decisionMaker game.PlayerID
	rng           *rand.Rand
}

// Player returns the seat the search is deciding for.
func (c SearchContext) Player() game.PlayerID {
	return c.player
}

// DecisionMaker identifies the agent's seat, which may differ from the actor.
func (c SearchContext) DecisionMaker() game.PlayerID { return c.decisionMaker }

// Determinize returns a full, self-consistent game state the search agent may
// read and simulate on. It re-samples the information hidden from the searching
// player — each library's order and every opponent's hand, re-dealt from that
// opponent's own deck (see determinize) — so the agent only ever sees plausible
// sampled worlds and never an opponent's true hand, preserving fog of war. Each
// call returns an independent sample, so a caller can average a decision over
// several determinizations (PIMC).
func (c SearchContext) Determinize() *game.Game {
	clone := c.game.Clone()
	clone.Execution = game.ExecutionDiagnostics{}
	determinize(clone, c.decisionMaker, c.rng)
	return clone
}

// Simulator returns a forward model whose randomness is isolated from the live
// game: it runs on a dedicated engine seeded from the search RNG stream, so
// rolling positions forward never perturbs the live game's RNG and the real game
// stays reproducible. It shares the engine's card implementations so simulated
// cards resolve exactly as in real play. Obtain one per decision and reuse it.
func (c SearchContext) Simulator() Simulator {
	seed := c.rng.Uint64()
	simEngine := &Engine{
		rng:                 rand.New(rand.NewPCG(seed, seed^searchSeedLo)),
		searchRNG:           rand.New(rand.NewPCG(seed^searchSeedHi, seed)),
		cardImplementations: c.engine.cardImplementations,
		programs:            c.engine.programs,
		DecisionLimit:       DefaultSearchDecisionLimit,
	}
	return Simulator{engine: simEngine}
}

func (e *Engine) newSearchContext(g *game.Game, playerID game.PlayerID) SearchContext {
	return SearchContext{engine: e, game: g, player: playerID,
		decisionMaker: g.DecisionMaker(playerID, game.DecisionRules), rng: e.searchRNG}
}

// decideAction asks the agent for its action at a priority or combat decision,
// routing through the search capability when the agent implements SearchAgent and
// through the observation-scoring path otherwise. The observation path runs
// inside a static-source frame so the agent's evaluation reuses one
// static-ability source scan; the frame is closed via defer so a panicking agent
// cannot leak it. The search path branches on cloned states and needs no frame on
// the live game.
func (e *Engine) decideAction(g *game.Game, agent PlayerAgent, playerID game.PlayerID, legal []action.Action) action.Action {
	if searcher, ok := agent.(SearchAgent); ok {
		return e.searchDecision(g, searcher, playerID, legal)
	}
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	return agent.ChooseAction(decisionObservation(g, playerID, game.DecisionRules), legal)
}

func (e *Engine) searchDecision(g *game.Game, searcher SearchAgent, playerID game.PlayerID, legal []action.Action) (chosen action.Action) {
	defer func() {
		if failure := recover(); failure != nil {
			if _, exhausted := failure.(WorkBudgetError); !exhausted {
				panic(failure)
			}
			// A hypothetical nonterminating branch is not live-game exhaustion.
			g.Execution.SearchBudgetStops++
			g.Execution.ActionFallbacks++
			chosen = legal[0]
			if containsAction(legal, action.Pass()) {
				chosen = action.Pass()
			}
		}
	}()
	return searcher.ChooseActionBySearch(e.newSearchContext(g, playerID), legal)
}
