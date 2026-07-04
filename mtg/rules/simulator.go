package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// This file exposes the engine's forward model as a small public Simulator API
// for search-based agents (see docs/adr/0011-search-based-agent-architecture.md).
// The rules engine is the single source of rules truth: the Simulator lets a
// search agent enumerate, branch, and roll a position forward on cloned states
// without re-implementing any rules.
//
// Every method that advances the game does so on a Clone of the caller's game
// and returns the new state, so the caller's position is never mutated and one
// position can be branched many times. Terminal detection and the winner are
// already public on the game itself (game.Game.IsGameOver, game.Game.Winner), so
// the Simulator does not duplicate them.

// Simulator is a forward-model handle over the rules engine. Obtain one with
// Engine.Simulator. It carries the engine's card implementations and RNG, so
// simulation resolves cards exactly as real play does.
//
// Simulation consumes the underlying engine's RNG stream. A search agent that
// must not perturb the RNG of the engine running the live game should build the
// Simulator from a dedicated engine instance seeded independently.
type Simulator struct {
	engine *Engine
}

// Simulator returns a forward-model handle bound to this engine's rules. It is
// cheap to create.
func (e *Engine) Simulator() Simulator {
	return Simulator{engine: e}
}

// LegalActions returns the actions playerID may legally take in g at the current
// decision point, exactly as the engine offers them when granting priority
// (productive actions before Pass). It is a pure read and does not mutate g, so a
// search agent can enumerate a node's branches without cloning first.
func (s Simulator) LegalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	return s.engine.legalActions(g, playerID)
}

// Apply returns a Clone of g with act applied on playerID's behalf, resolving any
// choices the action itself requires (modes, payment selections, and other
// mid-application decisions) through policies. The original g is never modified.
// ok is false, and the returned game nil, when act is not a legal action for
// playerID in g.
//
// Apply performs only the action itself: for a spell or activated ability it puts
// the object on the stack but does not resolve it. Follow with ResolvePriority to
// let the object resolve and opponents respond, then evaluate the resulting
// position.
func (s Simulator) Apply(g *game.Game, playerID game.PlayerID, act action.Action, policies [game.NumPlayers]PlayerAgent) (*game.Game, bool) {
	if !containsAction(s.engine.legalActions(g, playerID), act) {
		return nil, false
	}
	clone := g.Clone()
	if !s.engine.applyActionWithChoices(clone, playerID, act, policies, &TurnLog{}) {
		return nil, false
	}
	return clone, true
}

// ResolvePriority runs the priority loop forward from g's current point, using
// policies for every seat, until the current step's priority is exhausted — the
// stack has fully resolved and every active player has passed in succession, or
// the game ends. It returns a Clone advanced to that stable point; the original
// g is never modified.
//
// This lets a just-applied spell or activated ability resolve, with opponents
// given the chance to respond through their policies, so the resulting board can
// be evaluated. It advances within the current step only: it does not cross into
// later steps or phases (rolling a whole turn or game forward from an arbitrary
// mid-turn point is a separate capability built on top of this one).
func (s Simulator) ResolvePriority(g *game.Game, policies [game.NumPlayers]PlayerAgent) *game.Game {
	clone := g.Clone()
	s.engine.runPriorityLoop(clone, policies, &TurnLog{})
	return clone
}

// ResolveCombatWithAttackers returns a Clone of g in which playerID has declared
// the given attackers and the rest of the combat phase has been played out —
// opponents declaring blockers and responding via policies, then combat damage
// dealt — so the resulting board can be evaluated. The original g is never
// modified.
//
// It requires g to be at the declare-attackers step of a combat phase with the
// attackers not yet declared (the state a search agent sees when it is asked to
// declare attackers). ok is false, and the returned game nil, when g is not in
// that state or the attackers are not a legal declaration for playerID.
//
// This lets a search agent value an attack by what it actually does — which
// creatures trade, how much damage and commander damage lands, whether it is
// lethal — using the engine's authoritative combat sequence rather than a
// heuristic estimate.
func (s Simulator) ResolveCombatWithAttackers(g *game.Game, playerID game.PlayerID, attackers action.DeclareAttackersAction, policies [game.NumPlayers]PlayerAgent) (*game.Game, bool) {
	if g.Combat == nil || g.Turn.Step != game.StepDeclareAttackers {
		return nil, false
	}
	clone := g.Clone()
	ce := combatEngine{s.engine}
	if !ce.applyAttackers(clone, playerID, attackers) {
		return nil, false
	}
	ce.resolveCombatAfterAttackers(clone, policies, &TurnLog{})
	return clone, true
}
