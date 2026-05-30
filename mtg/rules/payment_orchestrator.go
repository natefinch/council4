package rules

import "github.com/natefinch/council4/mtg/game"

// paymentOrchestratorType is the single point in the rules package for all
// spell and ability cost checks and applications. Callers use paymentOrch
// instead of the package-level payment functions directly, so that future
// planner restructuring has a single entry point to update.
//
// The type carries no state; it delegates to the package-level payment
// functions without changing their behaviour.
type paymentOrchestratorType struct{}

// paymentOrch is the package-level singleton for payment orchestration.
var paymentOrch paymentOrchestratorType

// canPaySpellCosts reports whether the player can currently pay all costs for
// the spell described by req.
func (paymentOrchestratorType) canPaySpellCosts(g *game.Game, req spellPaymentRequest) bool {
	return canPaySpellCosts(g, req)
}

// paySpellCosts pays all spell costs described by req and returns the set of
// additional cost names that were paid plus a success flag.
func (paymentOrchestratorType) paySpellCosts(g *game.Game, req spellPaymentRequest) ([]string, bool) {
	return paySpellCosts(g, req)
}

// buildSpellCostPlan constructs (but does not apply) a payment plan for the
// spell described by req.
func (paymentOrchestratorType) buildSpellCostPlan(g *game.Game, req spellPaymentRequest) (spellCostPlan, bool) {
	return buildSpellCostPlan(g, req)
}

// buildAbilityCostPlan constructs (but does not apply) a payment plan for the
// activated ability described by req.
func (paymentOrchestratorType) buildAbilityCostPlan(g *game.Game, req abilityPaymentRequest) (abilityCostPlan, bool) {
	return buildAbilityCostPlan(g, req)
}

// payAbilityCosts pays all ability costs described by req and returns the
// settled cost plan plus a success flag.
func (paymentOrchestratorType) payAbilityCosts(g *game.Game, req abilityPaymentRequest) (abilityCostPlan, bool) {
	return payAbilityCosts(g, req)
}
