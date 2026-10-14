package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func (c *programContext) ControlNextTurn(reference game.PlayerReference) (bool, error) {
	if err := c.valid(); err != nil {
		return false, err
	}
	player, available, err := c.Player(reference)
	if err != nil || !available || !isPlayerAlive(c.game, player) || !isPlayerAlive(c.game, c.source.Controller) {
		return false, err
	}
	c.game.SchedulePlayerControlEffect(game.PlayerControlEffect{
		ID: c.game.IDGen.Next(), SourceCardID: c.source.CardID, SourceObjectID: c.source.ObjectID,
		Controller: c.source.Controller, Player: player, CreatedTurn: c.game.Turn.TurnNumber,
	})
	return true, nil
}

// chooseAction is the common priority/combat dispatch. Legality and payment
// always belong to actor; only the agent and information permissions change.
func (e *Engine) chooseAction(g *game.Game, agents [game.NumPlayers]PlayerAgent, actor game.PlayerID, legal []action.Action, fallback action.Action) action.Action {
	e.countDecision(g)
	decider := g.DecisionMaker(actor, game.DecisionRules)
	if agent := agentFor(agents, decider); agent != nil {
		chosen := e.decideAction(g, agent, actor, legal)
		if containsAction(legal, chosen) {
			return chosen
		}
		g.Execution.InvalidActions++
	}
	g.Execution.ActionFallbacks++
	return fallback
}

func decisionObservation(g *game.Game, actor game.PlayerID, kind game.DecisionKind) PlayerObservation {
	obs := NewObservation(g, g.DecisionMaker(actor, kind))
	obs.Player = actor
	return obs
}

// Concede is a personal seat action, never an option offered to a controlling
// agent. The caller identifies the conceding player, not their decision maker.
func (e *Engine) Concede(g *game.Game, player game.PlayerID) bool {
	return e.eliminatePlayer(g, player)
}
