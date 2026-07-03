package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// This file exposes the engine's forward model as a small public Simulator API
// for search-based agents (see docs/adr/0011-search-based-agent-architecture.md).
// The rules engine is the single source of rules truth: these methods let a
// search agent enumerate, branch, and roll a position forward on cloned states
// without re-implementing any rules.
//
// Every method that advances the game does so on a Clone of the caller's game
// and returns the new state, so the caller's position is never mutated and one
// position can be branched many times. Terminal detection and the winner are
// already public on the game itself (game.Game.IsGameOver, game.Game.Winner), so
// the Simulator API does not duplicate them.

// LegalActions returns the actions playerID may legally take in g at the current
// decision point, exactly as the engine offers them when granting priority
// (productive actions before Pass). It is a pure read and does not mutate g, so a
// search agent can enumerate a node's branches without cloning first.
func (e *Engine) LegalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	return e.legalActions(g, playerID)
}

// SimulateAction returns a Clone of g with act applied on playerID's behalf,
// resolving any choices the action itself requires (modes, payment selections,
// and other mid-application decisions) through policies. The original g is never
// modified. ok is false, and the returned game nil, when act is not a legal
// action for playerID in g.
//
// SimulateAction performs only the action itself: for a spell or activated
// ability it puts the object on the stack but does not resolve it. Follow with
// ResolvePriority to let the object resolve and opponents respond, then evaluate
// the resulting position.
func (e *Engine) SimulateAction(g *game.Game, playerID game.PlayerID, act action.Action, policies [game.NumPlayers]PlayerAgent) (*game.Game, bool) {
	if !containsAction(e.legalActions(g, playerID), act) {
		return nil, false
	}
	clone := g.Clone()
	if !e.applyActionWithChoices(clone, playerID, act, policies, &TurnLog{}) {
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
// This lets a just-cast spell or activated ability resolve, with opponents given
// the chance to respond through their policies, so the resulting board can be
// evaluated. It advances within the current step only: it does not cross into
// later steps or phases (rolling a whole turn or game forward from an arbitrary
// mid-turn point is a separate capability built on top of this one).
func (e *Engine) ResolvePriority(g *game.Game, policies [game.NumPlayers]PlayerAgent) *game.Game {
	clone := g.Clone()
	e.runPriorityLoop(clone, policies, &TurnLog{})
	return clone
}
