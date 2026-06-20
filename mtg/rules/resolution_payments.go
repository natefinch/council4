package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveResolutionPaymentValue(g *game.Game, obj *game.StackObject, res *game.ResolutionPayment, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (accepted, succeeded bool) {
	resolved := materializeResolutionPayment(g, obj, res)
	res = &resolved
	playerID, ok := resolutionPaymentPayer(g, obj, res)
	if !ok {
		return false, false
	}
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
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(res.ManaCost), res.AdditionalCosts, res.XValue, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID:        playerID,
		SourceCardID:    stackObjectSourceID(obj),
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
		Prefs:           prefs,
	}) {
		return true, false
	}
	return true, true
}

func materializeResolutionPayment(g *game.Game, obj *game.StackObject, res *game.ResolutionPayment) game.ResolutionPayment {
	if res == nil {
		return game.ResolutionPayment{}
	}
	resolved := *res
	if !res.DynamicGenericManaCost.Exists {
		return resolved
	}
	amount := max(0, dynamicAmountValue(g, obj, stackObjectController(obj), res.DynamicGenericManaCost.Val))
	resolved.ManaCost = opt.Val(cost.Mana{cost.O(amount)})
	resolved.Prompt = "Pay " + resolved.ManaCost.Val.String() + "?"
	resolved.DynamicGenericManaCost = opt.V[game.DynamicAmount]{}
	return resolved
}

func resolutionPaymentPayer(g *game.Game, obj *game.StackObject, res *game.ResolutionPayment) (game.PlayerID, bool) {
	if res != nil && res.Payer.Exists {
		return resolvePlayerReference(g, obj, res.Payer.Val)
	}
	return stackObjectController(obj), true
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
