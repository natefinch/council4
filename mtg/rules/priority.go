package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func (e *Engine) runPriorityLoop(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	consecutivePasses := 0

	for {
		e.applyStateBasedActionsWithLog(g, log)
		if g.IsGameOver() {
			return
		}
		if e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
			consecutivePasses = 0
			g.Turn.PriorityPlayer = g.Turn.ActivePlayer
			continue
		}

		activePlayers := g.TurnOrder.ActivePlayerCount()
		if activePlayers <= 0 {
			return
		}

		playerID := g.Turn.PriorityPlayer
		if !canAct(g, playerID) {
			g.Turn.PriorityPlayer = g.TurnOrder.NextPriority(playerID)
			continue
		}

		legal := e.legalActions(g, playerID)
		if len(legal) == 0 {
			legal = []action.Action{action.Pass()}
		}

		chosen := action.Pass()
		if agent := agentFor(agents, playerID); agent != nil {
			chosen = agent.ChooseAction(observe(g, playerID), legal)
		}
		if !containsAction(legal, chosen) {
			chosen = action.Pass()
		}

		log.addAction(ActionLog{
			Player: playerID,
			Action: chosen,
		})

		if !e.applyActionWithChoices(g, playerID, chosen, agents, log) {
			panic("applyAction failed for validated action")
		}
		if chosen.Kind == action.ActionPass {
			consecutivePasses++
			if consecutivePasses >= activePlayers {
				if g.Stack.IsEmpty() {
					return
				}
				e.resolveTopOfStackWithChoices(g, agents, log)
				consecutivePasses = 0
				g.Turn.PriorityPlayer = g.Turn.ActivePlayer
				continue
			}
			g.Turn.PriorityPlayer = g.TurnOrder.NextPriority(playerID)
			continue
		}

		consecutivePasses = 0
		g.Turn.PriorityPlayer = playerID
	}
}

func observe(g *game.Game, playerID game.PlayerID) PlayerObservation {
	return PlayerObservation{
		Player: playerID,
		Turn: TurnObservation{
			TurnNumber:     g.Turn.TurnNumber,
			ActivePlayer:   g.Turn.ActivePlayer,
			PriorityPlayer: g.Turn.PriorityPlayer,
			Phase:          g.Turn.Phase,
			Step:           g.Turn.Step,
		},
	}
}

func agentFor(agents [game.NumPlayers]PlayerAgent, playerID game.PlayerID) PlayerAgent {
	if playerID < 0 || int(playerID) >= len(agents) {
		return nil
	}
	return agents[playerID]
}

func containsAction(actions []action.Action, want action.Action) bool {
	for _, act := range actions {
		if actionsEqual(act, want) {
			return true
		}
	}
	return false
}

func actionsEqual(a, b action.Action) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		return a.PlayLand == b.PlayLand
	case action.ActionCastSpell:
		return a.CastSpell.CardID == b.CastSpell.CardID &&
			a.CastSpell.XValue == b.CastSpell.XValue &&
			slices.Equal(a.CastSpell.Targets, b.CastSpell.Targets) &&
			slices.Equal(a.CastSpell.ChosenModes, b.CastSpell.ChosenModes)
	case action.ActionActivateAbility:
		return a.ActivateAbility.SourceID == b.ActivateAbility.SourceID &&
			a.ActivateAbility.AbilityIndex == b.ActivateAbility.AbilityIndex &&
			a.ActivateAbility.XValue == b.ActivateAbility.XValue &&
			slices.Equal(a.ActivateAbility.Targets, b.ActivateAbility.Targets)
	case action.ActionDeclareAttackers:
		return slices.Equal(a.DeclareAttackers.Attackers, b.DeclareAttackers.Attackers)
	case action.ActionDeclareBlockers:
		return slices.Equal(a.DeclareBlockers.Blockers, b.DeclareBlockers.Blockers)
	default:
		return false
	}
}
