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

	// Lookahead is how many further full turns a candidate is rolled forward
	// through — opponents taking their turns via the rollout policy — before the
	// position is evaluated, so the agent sees a play's consequences on later
	// turns rather than only its immediate result (whether the board it built gets
	// attacked or wrathed, whether interaction it held up answers a threat). It
	// applies only to main-phase decisions, the sequencing choices where looking
	// ahead matters; zero evaluates the immediate position (a one-ply search).
	// Rolling forward is expensive, so a small horizon (one round) is typical.
	Lookahead int
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
		resolved.BeginStaticSourceFrame()
		value := Evaluate(rules.NewObservation(resolved, me))
		resolved.EndStaticSourceFrame()
		if value > bestValue {
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
	obs := rules.NewObservation(first, me)

	best := candidates[0]
	bestValue := math.Inf(-1)
	for i := range candidates {
		total := 0.0
		for _, world := range worlds {
			total += s.actionValue(sim, world, me, candidates[i], applyPolicies, resolvePolicies)
		}
		average := total/float64(len(worlds)) + searchPolicyPriorWeight*s.Rollout.ScoreAction(obs, candidates[i])
		if average > bestValue {
			bestValue = average
			best = candidates[i]
		}
	}
	return best
}

// searchPolicyPriorWeight scales the rollout policy's per-action score when it is
// added to a candidate's evaluated position value. The position value (Evaluate)
// dominates when it clearly separates candidates, but on the many decisions where
// it does not — playing a land nets about zero (the mana source gained cancels
// the card spent from hand), tapping a land for mana is neutral (a tapped source
// still counts), a small spell barely registers — a value-only search picks
// arbitrarily. The rollout policy's tuned score then breaks the near-tie toward
// the sound play (develop mana, do not float it, cast the better spell), the way
// a policy prior guides a value search. It is deliberately small so a real
// position difference still overrides the prior.
const searchPolicyPriorWeight = 0.01

// deepSearchWidth is how many top candidates, ranked by the cheap one-ply value,
// get rolled forward when Lookahead is set. Rolling a candidate forward is
// several times the cost of a one-ply evaluation, so rolling every candidate
// forward multiplies a decision's cost past the point of being playable in the
// browser. A strong play almost always ranks near the top on the immediate
// position too, so evaluating only the most promising handful deeply keeps nearly
// all of the benefit at a fraction of the cost.
const deepSearchWidth = 3

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
	// Reading the live world to pre-rank candidates and score the policy prior
	// scans the battlefield's static-ability sources repeatedly; a frame builds
	// them once for the whole decision. The world is only read here (each candidate
	// is applied to a clone), so the frame stays valid.
	world.BeginStaticSourceFrame()
	defer world.EndStaticSourceFrame()
	candidates := s.candidateActions(world, me, legal)
	applyPolicies := s.uniformPolicies()
	resolvePolicies := s.resolvePolicies(me)
	obs := rules.NewObservation(world, me)
	priorOf := func(a action.Action) float64 {
		return searchPolicyPriorWeight * s.Rollout.ScoreAction(obs, a)
	}

	if s.Lookahead <= 0 {
		best := candidates[0]
		bestValue := math.Inf(-1)
		for i := range candidates {
			value := s.actionValue(sim, world, me, candidates[i], applyPolicies, resolvePolicies) + priorOf(candidates[i])
			if value > bestValue {
				bestValue = value
				best = candidates[i]
			}
		}
		return best
	}

	// Deeper search: rank every candidate by the cheap one-ply value, then roll
	// only the most promising few forward and choose among those by the deeper
	// value. Two stages bound the cost of the expensive forward roll to a handful
	// of candidates while still letting a play's later consequences decide between
	// the plausible choices.
	type scoredAction struct {
		action  action.Action
		shallow float64
	}
	shortlist := make([]scoredAction, len(candidates))
	for i := range candidates {
		shortlist[i] = scoredAction{
			action:  candidates[i],
			shallow: s.actionValue(sim, world, me, candidates[i], applyPolicies, resolvePolicies) + priorOf(candidates[i]),
		}
	}
	slices.SortStableFunc(shortlist, func(a, b scoredAction) int {
		return cmp.Compare(b.shallow, a.shallow)
	})
	if len(shortlist) > deepSearchWidth {
		shortlist = shortlist[:deepSearchWidth]
	}

	best := shortlist[0].action
	bestValue := math.Inf(-1)
	for i := range shortlist {
		value := s.deepActionValue(sim, world, me, shortlist[i].action, applyPolicies, resolvePolicies) + priorOf(shortlist[i].action)
		if value > bestValue {
			bestValue = value
			best = shortlist[i].action
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

// actionValue scores one candidate action one ply deep: apply it to the world
// (resolving the searcher's own action choices with the rollout policy), let it
// resolve while opponents respond, then evaluate the resulting position for me.
// An action that turns out to be illegal in the determinized world scores as the
// worst option so it is never chosen over a real play.
func (s Searcher) actionValue(sim rules.Simulator, world *game.Game, me game.PlayerID, act action.Action, applyPolicies, resolvePolicies [game.NumPlayers]rules.PlayerAgent) float64 {
	resolved, ok := s.resolveCandidate(sim, world, me, act, applyPolicies, resolvePolicies)
	if !ok {
		return math.Inf(-1)
	}
	return evaluateForSeat(resolved, me)
}

// deepActionValue scores a candidate by rolling the position it produces forward
// through the rest of this turn and the next Lookahead turns — opponents
// responding via the rollout policy — before evaluating it, so the score
// reflects a play's later consequences (whether the board it builds is attacked
// or answered, whether interaction it holds up gets to matter) rather than only
// its immediate result. Rolling forward resumes from the phase after the one
// whose priority just resolved, which is exact only for a main phase, so off a
// main phase (or with no Lookahead) this is the one-ply actionValue.
func (s Searcher) deepActionValue(sim rules.Simulator, world *game.Game, me game.PlayerID, act action.Action, applyPolicies, resolvePolicies [game.NumPlayers]rules.PlayerAgent) float64 {
	resolved, ok := s.resolveCandidate(sim, world, me, act, applyPolicies, resolvePolicies)
	if !ok {
		return math.Inf(-1)
	}
	if s.Lookahead > 0 && isMainPhase(resolved.Turn.Phase) {
		resolved = sim.PlayForward(resolved, s.uniformPolicies(), s.Lookahead)
	}
	return evaluateForSeat(resolved, me)
}

// resolveCandidate applies a candidate action to a clone of the world and
// resolves the priority that follows, returning the resulting position. The
// second result is false when the action is illegal in this determinized world.
func (Searcher) resolveCandidate(sim rules.Simulator, world *game.Game, me game.PlayerID, act action.Action, applyPolicies, resolvePolicies [game.NumPlayers]rules.PlayerAgent) (*game.Game, bool) {
	afterAction, ok := sim.Apply(world, me, act, applyPolicies)
	if !ok {
		return nil, false
	}
	return sim.ResolvePriority(afterAction, resolvePolicies), true
}

// evaluateForSeat scores a resolved position for the seat. The evaluation runs
// inside a static-source frame so that reading every permanent's effective
// characteristics builds the battlefield's static-ability sources once and
// memoizes each permanent's values, rather than rebuilding them per permanent —
// the dominant cost of evaluating a large real-deck board.
func evaluateForSeat(g *game.Game, me game.PlayerID) float64 {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	return Evaluate(rules.NewObservation(g, me))
}

// isMainPhase reports whether phase is a main phase, the only phase a search
// rollout can resume past (see rules.Simulator.PlayForward).
func isMainPhase(phase game.Phase) bool {
	return phase == game.PhasePrecombatMain || phase == game.PhasePostcombatMain
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
