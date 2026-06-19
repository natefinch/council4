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
			legal = []action.Action{actionBuild.pass()}
		}

		chosen := actionBuild.pass()
		if agent := agentFor(agents, playerID); agent != nil {
			// An agent inspects the game through a read-only observation and must
			// not mutate it, so a static-source frame lets its evaluation reuse
			// one static-ability source scan instead of rescanning per permanent.
			// The frame is closed via defer so a panicking agent cannot leak it.
			func() {
				g.BeginStaticSourceFrame()
				defer g.EndStaticSourceFrame()
				chosen = agent.ChooseAction(observe(g, playerID), legal)
			}()
		}
		if !containsAction(legal, chosen) {
			chosen = actionBuild.pass()
		}

		log.addAction(&ActionLog{
			Player: playerID,
			Action: chosen,
		})

		if !e.applyActionWithChoices(g, playerID, chosen, agents, log) {
			panic("applyAction failed for validated action")
		}
		if chosen.Kind != action.ActionPass {
			e.notifyActionObservers(g, agents, playerID, chosen)
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
	return NewObservation(g, playerID)
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
		aPayload, aOK := a.PlayLandPayload()
		bPayload, bOK := b.PlayLandPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionCastSpell:
		aPayload, aOK := a.CastSpellPayload()
		bPayload, bOK := b.CastSpellPayload()
		return aOK && bOK &&
			aPayload.CardID == bPayload.CardID &&
			aPayload.SourceZone == bPayload.SourceZone &&
			aPayload.Face == bPayload.Face &&
			aPayload.XValue == bPayload.XValue &&
			aPayload.KickerPaid == bPayload.KickerPaid &&
			aPayload.Mutate == bPayload.Mutate &&
			aPayload.MutateTargetID == bPayload.MutateTargetID &&
			slices.Equal(aPayload.Targets, bPayload.Targets) &&
			slices.Equal(aPayload.ChosenModes, bPayload.ChosenModes)
	case action.ActionActivateAbility:
		aPayload, aOK := a.ActivateAbilityPayload()
		bPayload, bOK := b.ActivateAbilityPayload()
		return aOK && bOK &&
			aPayload.SourceID == bPayload.SourceID &&
			aPayload.AbilityIndex == bPayload.AbilityIndex &&
			aPayload.XValue == bPayload.XValue &&
			slices.Equal(aPayload.Targets, bPayload.Targets) &&
			slices.Equal(aPayload.TargetCounts, bPayload.TargetCounts) &&
			slices.Equal(aPayload.ChosenModes, bPayload.ChosenModes)
	case action.ActionSuspendCard:
		aPayload, aOK := a.SuspendCardPayload()
		bPayload, bOK := b.SuspendCardPayload()
		return aOK && bOK && aPayload == bPayload
	case action.ActionDeclareAttackers:
		aPayload, aOK := a.DeclareAttackersPayload()
		bPayload, bOK := b.DeclareAttackersPayload()
		return aOK && bOK && slices.Equal(aPayload.Attackers, bPayload.Attackers)
	case action.ActionDeclareBlockers:
		aPayload, aOK := a.DeclareBlockersPayload()
		bPayload, bOK := b.DeclareBlockersPayload()
		return aOK && bOK && slices.Equal(aPayload.Blockers, bPayload.Blockers)
	default:
		return false
	}
}
