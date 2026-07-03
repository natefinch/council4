package agent

import (
	"math"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// Searcher is a search-based agent (docs/adr/0011-search-based-agent-architecture.md).
// At a priority decision it looks one ply ahead: for each legal action it
// simulates the action, lets it resolve — with opponents responding via the
// rollout policy — and evaluates the resulting position with Evaluate, then plays
// the highest-valued action. Combat declarations and non-action choices fall back
// to the rollout policy.
//
// This is milestone S1: the search is one ply deep with perfect information and
// evaluates positions with the heuristic Evaluate. Deeper lookahead, rollouts to
// terminal, and determinized hidden information layer on top without changing the
// agent's shape.
type Searcher struct {
	// Rollout is the policy that drives opponents while a candidate action
	// resolves, and the fallback the agent uses for combat declarations and
	// non-action choices. It should be a fast, sensible strategy such as
	// GenericStrategy.
	Rollout Strategy
}

// Compile-time checks that a Searcher drives every engine decision point and is
// detected as a search agent.
var (
	_ rules.PlayerAgent = Searcher{}
	_ rules.ChoiceAgent = Searcher{}
	_ rules.SearchAgent = Searcher{}
)

// ChooseAction implements rules.PlayerAgent. It is the fallback used for combat
// declarations and anywhere the engine does not route through search: it plays
// the rollout policy's choice.
func (s Searcher) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	return Agent{Strategy: s.Rollout}.ChooseAction(obs, legal)
}

// ChooseChoice implements rules.ChoiceAgent by delegating to the rollout policy,
// so the searching agent's non-action choices stay consistent with the policy
// that drives its simulations.
func (s Searcher) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	return s.Rollout.ChooseChoice(obs, request)
}

// ChooseActionBySearch implements rules.SearchAgent with one-ply lookahead and
// position evaluation. For each legal action it applies the action to a
// determinized world, resolves it (opponents responding via the rollout policy
// while the searcher itself only passes), and scores the resulting position for
// the searching seat. It plays the highest-scoring action, breaking ties toward
// the earlier action in the engine's order (productive actions before Pass).
func (s Searcher) ChooseActionBySearch(ctx rules.SearchContext, legal []action.Action) action.Action {
	return s.searchBestAction(ctx.Simulator(), ctx.Determinize(), ctx.Player(), legal)
}

// searchBestAction is the search core, separated from the SearchContext so it can
// be driven directly in tests with a Simulator and a constructed world. It plays
// the highest-scoring action, breaking ties toward the earlier action in the
// engine's order (productive actions before Pass).
func (s Searcher) searchBestAction(sim rules.Simulator, world *game.Game, me game.PlayerID, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	applyPolicies := s.uniformPolicies()
	resolvePolicies := s.resolvePolicies(me)

	best := legal[0]
	bestValue := math.Inf(-1)
	for i := range legal {
		value := s.actionValue(sim, world, me, legal[i], applyPolicies, resolvePolicies)
		if value > bestValue {
			bestValue = value
			best = legal[i]
		}
	}
	return best
}

// actionValue scores one candidate action: apply it to the world (resolving the
// searcher's own action choices with the rollout policy), let it resolve while
// opponents respond, then evaluate the resulting position for me. An action that
// turns out to be illegal in the determinized world scores as the worst option
// so it is never chosen over a real play.
func (Searcher) actionValue(sim rules.Simulator, world *game.Game, me game.PlayerID, act action.Action, applyPolicies, resolvePolicies [game.NumPlayers]rules.PlayerAgent) float64 {
	afterAction, ok := sim.Apply(world, me, act, applyPolicies)
	if !ok {
		return math.Inf(-1)
	}
	resolved := sim.ResolvePriority(afterAction, resolvePolicies)
	return Evaluate(rules.NewObservation(resolved, me))
}

// uniformPolicies drives every seat with the rollout policy. It is used while a
// single action is applied, where only the acting seat makes any choices (its
// modes, targets, and payment selections).
func (s Searcher) uniformPolicies() [game.NumPlayers]rules.PlayerAgent {
	rollout := Agent{Strategy: s.Rollout}
	var policies [game.NumPlayers]rules.PlayerAgent
	for i := range policies {
		policies[i] = rollout
	}
	return policies
}

// resolvePolicies drives opponents with the rollout policy but makes the
// searching seat pass (while still making its own in-resolution choices), so the
// candidate action resolves without the searcher chaining further plays that
// would blur the comparison between candidates.
func (s Searcher) resolvePolicies(me game.PlayerID) [game.NumPlayers]rules.PlayerAgent {
	policies := s.uniformPolicies()
	policies[me] = passThenChoose{inner: s.Rollout}
	return policies
}

// passThenChoose is the searcher's own policy while resolving a candidate action:
// it always passes priority — so it does not chain further plays — but still
// makes real non-action choices, so the candidate's own targets, modes, and
// discards resolve as they would in real play.
type passThenChoose struct {
	inner Strategy
}

func (passThenChoose) ChooseAction(_ rules.PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func (p passThenChoose) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	return p.inner.ChooseChoice(obs, request)
}
