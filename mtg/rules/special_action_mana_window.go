package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// specialActionCost is the mana cost manaCost that player pays to take a
// special action (CR 116.2), as an activationCost: it can be paid while
// present reports the action's card or permanent still where the action
// takes it from. Feasibility and the payment-time mana window follow an
// activation's (see activationFeasible), since a player may activate mana
// abilities whenever a rule asks for a mana payment (CR 605.3a).
func specialActionCost(player game.PlayerID, manaCost *cost.Mana, present func(*game.Game) bool) activationCost {
	return genericRequestCost(payment.GenericRequest{PlayerID: player, Cost: manaCost}, present)
}

// handCardPresent reports whether the card cardID is in player's hand in the
// game checked.
func handCardPresent(player game.PlayerID, cardID id.ID) func(*game.Game) bool {
	return func(g *game.Game) bool {
		p, ok := playerByID(g, player)
		return ok && p.Hand.Contains(cardID)
	}
}

// paySpecialActionCost pays the mana cost manaCost of a special action begun
// by player, whose cost c was feasible as the action began. If the cost
// includes a mana payment, player first may activate mana abilities
// (CR 605.3a, 601.2g); then the cost is paid (CR 601.2h). It returns
// actionApplied once the cost is paid. If the cost can't then be paid, the
// action is reversed along with the reversible mana abilities activated for
// it (CR 733.1) and it returns actionReversed, or actionFailed when no mana
// ability was activated.
func (e *Engine) paySpecialActionCost(g *game.Game, player game.PlayerID, c activationCost, manaCost *cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	undoMana := e.activationPaymentManaWindow(g, player, c, agents, log)
	if undoMana != nil && !c.payableNow(g) {
		return reverseActivation(undoMana)
	}
	prefs := e.paymentPreferencesForCost(g, player, manaCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: player, Cost: manaCost, Prefs: prefs}) {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	return actionApplied
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

// isManaCostSpecialAction reports whether act is a special action with a
// mana cost (see attemptSpecialAction).
func isManaCostSpecialAction(act action.Action) bool {
	switch act.Kind {
	case action.ActionSuspendCard, action.ActionPlotCard, action.ActionForetellCard, action.ActionTurnFaceUp:
		return true
	default:
		return false
	}
}

// specialActionAttemptKey identifies a special action an attempt redoes.
func specialActionAttemptKey(act action.Action) string {
	return fmt.Sprintf("special|%+v", act)
}

// reverseSpecialActionAttempt records a special action reversed while its
// player holds priority, as reverseActivationAttempt records an activation:
// a repeat of one already reversed exhausts it (CR 733.2).
func (a *castAttempts) reverseSpecialActionAttempt(act action.Action) {
	a.reverseAttempt(specialActionAttemptKey(act))
}
