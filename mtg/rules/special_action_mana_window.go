package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func handSpecialActionCost(player game.PlayerID, cardID id.ID, manaCost *cost.Mana) activationCost {
	return genericRequestCost(payment.GenericRequest{PlayerID: player, Cost: manaCost, SourceCardID: cardID, SourceZone: zone.Hand}, handCardPresent(player, cardID))
}

// handCardPresent reports whether the card cardID is in player's hand in the
// game checked.
func handCardPresent(player game.PlayerID, cardID id.ID) func(*game.Game) bool {
	return func(g *game.Game) bool {
		p, ok := playerByID(g, player)
		return ok && p.Hand.Contains(cardID)
	}
}

// attemptSpecialAction takes act when it is a special action with a mana
// cost: suspending, plotting or foretelling a card, or turning a face-down
// permanent face up (CR 116.2b, 116.2f, 116.2h, 116.2k). It reports false for
// any other action.
func (e *Engine) attemptSpecialAction(g *game.Game, player game.PlayerID, act action.Action, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (actionOutcome, bool) {
	switch act.Kind {
	case action.ActionSuspendCard:
		suspend, ok := act.SuspendCardPayload()
		if !ok {
			return actionFailed, true
		}
		return e.applySuspendCard(g, player, suspend.CardID, agents, log), true
	case action.ActionPlotCard:
		plot, ok := act.PlotCardPayload()
		if !ok {
			return actionFailed, true
		}
		return e.applyPlotCard(g, player, plot.CardID, agents, log), true
	case action.ActionForetellCard:
		foretell, ok := act.ForetellCardPayload()
		if !ok {
			return actionFailed, true
		}
		return e.applyForetellCard(g, player, foretell.CardID, agents, log), true
	case action.ActionTurnFaceUp:
		turnFaceUp, ok := act.TurnFaceUpPayload()
		if !ok {
			return actionFailed, true
		}
		return e.applyTurnFaceUpWithChoices(g, player, turnFaceUp.PermanentID, agents, log), true
	default:
		return actionFailed, false
	}
}
