package agent

import (
	"cmp"
	"math"
	"slices"

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

	// Budget caps how many legal actions are evaluated by simulation at each
	// decision. When more actions are legal, the agent pre-ranks them cheaply with
	// the rollout policy and simulates only the most promising, plus Pass, so a
	// decision's cost stays bounded regardless of how many actions are legal.
	// Zero or negative means no cap: every legal action is simulated.
	Budget int

	// Determinizations is how many sampled worlds a decision is averaged over
	// (Perfect Information Monte Carlo). Each world re-samples the opponents'
	// hidden hands and every library order, so averaging reduces the chance of
	// deciding against one unlucky sample. Zero or one searches a single sampled
	// world. Higher values play more robustly at a proportional cost.
	Determinizations int
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

// ChooseActionBySearch implements rules.SearchAgent. It handles two decision
// kinds: a declare-attackers decision, which it evaluates by playing each
// candidate attack out through combat (see searchAttackers), and an ordinary
// priority decision, which it evaluates one ply by applying and resolving the
// action. In both it plays the action whose resulting position scores highest for
// the searching seat.
func (s Searcher) ChooseActionBySearch(ctx rules.SearchContext, legal []action.Action) action.Action {
	me := ctx.Player()
	if isAttackDeclaration(legal) {
		// Attack declarations happen on the searcher's own turn; evaluate each
		// candidate by its actual combat result.
		return s.searchAttackers(ctx.Simulator(), ctx.Determinize(), me, legal)
	}

	world := ctx.Determinize()
	// Search is expensive, so spend it where it matters: the searcher's own turn,
	// where the meaningful sequencing, deployment, and combat decisions are made.
	// On opponents' turns the decision is a reactive window (hold or answer), which
	// the rollout policy already handles well (it holds interaction and answers the
	// biggest threat), so use the fast heuristic there instead of a full search.
	if world.Turn.ActivePlayer != me {
		return Agent{Strategy: s.Rollout}.ChooseAction(rules.NewObservation(world, me), legal)
	}
	if s.Determinizations > 1 {
		return s.searchAcrossWorlds(ctx, world, me, legal)
	}
	return s.searchBestAction(ctx.Simulator(), world, me, legal)
}

// searchAttackers chooses a declare-attackers action by simulating each candidate
// attack all the way through combat — opponents declaring blockers and responding
// via the rollout policy — and evaluating the resulting position. It plays the
// attack (or the no-attack option) whose post-combat position scores highest for
// the searching seat, so the agent attacks for real value: profitable trades,
// connecting damage, and lethal, while declining attacks that a good defender
// punishes. Ties resolve to the earlier action in the engine's order.
func (s Searcher) searchAttackers(sim rules.Simulator, world *game.Game, me game.PlayerID, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	policies := s.resolvePolicies(me)

	best := legal[0]
	bestValue := math.Inf(-1)
	for i := range legal {
		payload, ok := legal[i].DeclareAttackersPayload()
		if !ok {
			continue
		}
		resolved, ok := sim.ResolveCombatWithAttackers(world, me, payload, policies)
		if !ok {
			continue
		}
		if value := Evaluate(rules.NewObservation(resolved, me)); value > bestValue {
			bestValue = value
			best = legal[i]
		}
	}
	return best
}

// isAttackDeclaration reports whether the legal actions are declare-attackers
// options, so the searcher can route them through combat rollout rather than the
// priority-action search.
func isAttackDeclaration(legal []action.Action) bool {
	for i := range legal {
		if legal[i].Kind == action.ActionDeclareAttackers {
			return true
		}
	}
	return false
}

// searchAcrossWorlds averages each candidate action's value over several
// determinized worlds (Perfect Information Monte Carlo), so a decision is robust
// to any single unlucky sample of the opponents' hidden hands. The first world is
// reused for the own-turn check and candidate pre-ranking; the rest are sampled
// here.
func (s Searcher) searchAcrossWorlds(ctx rules.SearchContext, first *game.Game, me game.PlayerID, legal []action.Action) action.Action {
	if len(legal) == 1 {
		return legal[0]
	}
	worlds := make([]*game.Game, 0, s.Determinizations)
	worlds = append(worlds, first)
	for len(worlds) < s.Determinizations {
		worlds = append(worlds, ctx.Determinize())
	}

	sim := ctx.Simulator()
	applyPolicies := s.uniformPolicies()
	resolvePolicies := s.resolvePolicies(me)
	candidates := s.candidateActions(first, me, legal)

	best := candidates[0]
	bestValue := math.Inf(-1)
	for i := range candidates {
		total := 0.0
		for _, world := range worlds {
			total += s.actionValue(sim, world, me, candidates[i], applyPolicies, resolvePolicies)
		}
		if average := total / float64(len(worlds)); average > bestValue {
			bestValue = average
			best = candidates[i]
		}
	}
	return best
}

// searchBestAction is the search core, separated from the SearchContext so it can
// be driven directly in tests with a Simulator and a constructed world. It plays
// the highest-scoring action, breaking ties toward the earlier action in the
// engine's order (productive actions before Pass).
func (s Searcher) searchBestAction(sim rules.Simulator, world *game.Game, me game.PlayerID, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	if len(legal) == 1 {
		// A forced decision (typically only Pass): nothing to search.
		return legal[0]
	}
	candidates := s.candidateActions(world, me, legal)
	applyPolicies := s.uniformPolicies()
	resolvePolicies := s.resolvePolicies(me)

	best := candidates[0]
	bestValue := math.Inf(-1)
	for i := range candidates {
		value := s.actionValue(sim, world, me, candidates[i], applyPolicies, resolvePolicies)
		if value > bestValue {
			bestValue = value
			best = candidates[i]
		}
	}
	return best
}

// candidateActions is the set of legal actions the search will simulate. With no
// budget, or when the budget covers every action, it is the legal set unchanged.
// Otherwise it pre-ranks the legal actions with the rollout policy's cheap scorer
// and keeps the highest-scoring Budget of them, always including Pass so the
// "do nothing" baseline is evaluated. Pruning the least promising actions bounds
// a decision's simulation cost without discarding the plays worth considering.
func (s Searcher) candidateActions(world *game.Game, me game.PlayerID, legal []action.Action) []action.Action {
	if s.Budget <= 0 || len(legal) <= s.Budget {
		return legal
	}
	obs := rules.NewObservation(world, me)
	ranked := append([]action.Action(nil), legal...)
	slices.SortStableFunc(ranked, func(a, b action.Action) int {
		return cmp.Compare(s.Rollout.ScoreAction(obs, b), s.Rollout.ScoreAction(obs, a))
	})
	candidates := ranked[:s.Budget]
	if !containsPass(candidates) {
		candidates = append(candidates, action.Pass())
	}
	return candidates
}

func containsPass(actions []action.Action) bool {
	for i := range actions {
		if actions[i].Kind == action.ActionPass {
			return true
		}
	}
	return false
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
