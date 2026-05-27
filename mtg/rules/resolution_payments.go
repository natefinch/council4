package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) resolveResolutionPayment(g *game.Game, obj *game.StackObject, effect game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (bool, bool) {
	if !effect.Payment.Exists {
		return true, true
	}
	payment := &effect.Payment.Val
	playerID := stackObjectController(obj)
	if !canPayResolutionPayment(g, playerID, payment) {
		return false, false
	}
	prompt := payment.Prompt
	if prompt == "" {
		prompt = "Pay resolution cost?"
	}
	if !e.chooseMay(g, agents, playerID, prompt, log) {
		return false, false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(payment.ManaCost), payment.AdditionalCosts, agents, log)
	if !payResolutionPaymentWithPreferences(g, playerID, payment, prefs) {
		return true, false
	}
	return true, true
}

func canPayResolutionPayment(g *game.Game, playerID game.PlayerID, payment *game.ResolutionPayment) bool {
	if payment == nil {
		return true
	}
	if _, ok := buildResolutionPaymentPlan(g, playerID, payment, nil); !ok {
		return false
	}
	return true
}

func payResolutionPaymentWithPreferences(g *game.Game, playerID game.PlayerID, payment *game.ResolutionPayment, prefs *paymentPreferences) bool {
	plan, ok := buildResolutionPaymentPlan(g, playerID, payment, prefs)
	if !ok {
		return false
	}
	player, ok := playerForCostPayment(g, playerID)
	if !ok || !additionalCostPlanStillValid(g, player, plan.additional) || !paymentPlanStillValid(g, player, plan.mana) {
		return false
	}
	if !applyPaymentPlan(g, playerID, plan.mana) {
		return false
	}
	if !applyAdditionalCostPlan(g, plan.additional) {
		panic("resolution payment plan became invalid while paying additional costs")
	}
	return true
}

func buildResolutionPaymentPlan(g *game.Game, playerID game.PlayerID, payment *game.ResolutionPayment, prefs *paymentPreferences) (spellCostPlan, bool) {
	plan := spellCostPlan{}
	if payment == nil {
		return plan, true
	}
	additional, ok := buildAdditionalCostPlanForCosts(g, playerID, payment.AdditionalCosts, prefs)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(g, playerID, manaCostPtr(payment.ManaCost), payment.XValue, excluded, prefs)
	if !ok {
		return plan, false
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}
