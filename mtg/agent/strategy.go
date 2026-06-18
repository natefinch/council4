package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// Strategy is the scoring seam an Agent delegates to. Implementations must be
// deterministic given the same observation and inputs, so simulations are
// reproducible.
//
// The interface is intentionally small so different strategies (a generic
// "good stuff" scorer, archetype-specific strategies) can be swapped without
// changing the Agent that drives the rules engine.
type Strategy interface {
	// ScoreAction returns a desirability score for taking act in the given
	// observation. The Agent plays the highest-scoring legal action, breaking
	// ties toward the earlier action in the engine's legal-action order.
	ScoreAction(obs rules.PlayerObservation, act action.Action) float64

	// ChooseChoice selects option indices answering an engine-mediated choice
	// request (targets, modes, scry, discard, ordering, and similar). Returning
	// an invalid selection is safe: the rules engine validates the result and
	// applies its deterministic fallback when needed.
	ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int
}

// Agent is the unified decision-maker for one seat. It satisfies both
// rules.PlayerAgent (priority and combat actions) and rules.ChoiceAgent
// (non-action choices) by delegating to a single Strategy, so an agent's action
// and choice behaviour stay consistent.
type Agent struct {
	Strategy Strategy
}

// Compile-time checks that an Agent drives both engine decision points.
var (
	_ rules.PlayerAgent = Agent{}
	_ rules.ChoiceAgent = Agent{}
)

// ChooseAction implements rules.PlayerAgent. It plays the highest-scoring legal
// action; ties resolve to the earlier action in the engine's legal order, which
// places productive actions before Pass.
func (a Agent) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	best := legal[0]
	bestScore := a.Strategy.ScoreAction(obs, legal[0])
	for i := 1; i < len(legal); i++ {
		if score := a.Strategy.ScoreAction(obs, legal[i]); score > bestScore {
			best = legal[i]
			bestScore = score
		}
	}
	return best
}

// ChooseChoice implements rules.ChoiceAgent by delegating to the Strategy.
func (a Agent) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	return a.Strategy.ChooseChoice(obs, request)
}

// BaselineStrategy is a trivial Strategy: it scores every action equally, so an
// Agent using it plays the first legal action (productive actions before Pass),
// and answers choices with the request's default or the first required options.
// It exists to exercise the Strategy seam and as a comparison baseline; richer
// strategies replace it without changing Agent.
type BaselineStrategy struct{}

// ScoreAction implements Strategy. Every action scores equally, so the Agent
// keeps the engine's preferred ordering.
func (BaselineStrategy) ScoreAction(_ rules.PlayerObservation, _ action.Action) float64 {
	return 0
}

// ChooseChoice implements Strategy. It selects the request's default selection
// when present, otherwise the first MinChoices options, which is a valid
// selection for ordering and most bounded choices.
func (BaselineStrategy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if len(request.DefaultSelection) > 0 {
		return append([]int(nil), request.DefaultSelection...)
	}
	count := min(request.MinChoices, len(request.Options))
	selected := make([]int, 0, count)
	for i := range count {
		selected = append(selected, request.Options[i].Index)
	}
	return selected
}
