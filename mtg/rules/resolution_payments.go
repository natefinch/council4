package rules

import (
	"github.com/natefinch/council4/mtg/game"
	payment "github.com/natefinch/council4/mtg/rules/payment"
)

func (e *Engine) resolveResolutionPayment(g *game.Game, obj *game.StackObject, effect game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (bool, bool) {
	if !effect.Payment.Exists {
		return true, true
	}
	res := &effect.Payment.Val
	playerID := stackObjectController(obj)
	if !canPayResolutionPayment(g, playerID, res) {
		return false, false
	}
	prompt := res.Prompt
	if prompt == "" {
		prompt = "Pay resolution cost?"
	}
	if !e.chooseMay(g, agents, playerID, prompt, log) {
		return false, false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(res.ManaCost), res.AdditionalCosts, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID:        playerID,
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
		Prefs:           prefs,
	}) {
		return true, false
	}
	return true, true
}

func canPayResolutionPayment(g *game.Game, playerID game.PlayerID, res *game.ResolutionPayment) bool {
	if res == nil {
		return true
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{
		PlayerID:        playerID,
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
	})
}
