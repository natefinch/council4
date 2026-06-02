package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// paymentOrchestratorType is the single point in the rules package for all
// spell and ability cost checks and applications. Callers use paymentOrch
// instead of the payment subpackage directly, so rules-local state is adapted
// to payment.State in one place.
//
// Each method creates a short-lived payment.Planner bound to the current
// *game.Game and forwards the call. The type itself carries no state.
type paymentOrchestratorType struct{}

// paymentOrch is the package-level singleton for payment orchestration.
var paymentOrch paymentOrchestratorType

func (paymentOrchestratorType) planner(g *game.Game) payment.Planner {
	return payment.New(&rulesPaymentState{g: g})
}

// canPaySpellCosts reports whether the player can currently pay all costs for
// the spell described by req.
func (o paymentOrchestratorType) canPaySpellCosts(g *game.Game, req payment.SpellRequest) bool {
	return o.planner(g).CanPaySpellCosts(req)
}

// paySpellCosts pays all spell costs described by req and returns the set of
// additional cost names that were paid plus a success flag.
func (o paymentOrchestratorType) paySpellCosts(g *game.Game, req payment.SpellRequest) ([]string, bool) {
	return o.planner(g).PaySpellCosts(req)
}

// buildAbilityCostPlan reports whether a plan can be built for the ability
// described by req, without applying it.
func (o paymentOrchestratorType) buildAbilityCostPlan(g *game.Game, req payment.AbilityRequest) bool {
	return o.planner(g).BuildAbilityCostPlan(req)
}

// payAbilityCosts pays all ability costs described by req.
func (o paymentOrchestratorType) payAbilityCosts(g *game.Game, req payment.AbilityRequest) bool {
	return o.planner(g).PayAbilityCosts(req)
}

// canPayGenericCost reports whether the player can pay the mana cost described by req.
func (o paymentOrchestratorType) canPayGenericCost(g *game.Game, req payment.GenericRequest) bool {
	return o.planner(g).CanPayGenericCost(req)
}

// payGenericCost builds, validates, and applies the mana cost described by req.
func (o paymentOrchestratorType) payGenericCost(g *game.Game, req payment.GenericRequest) bool {
	return o.planner(g).PayGenericCost(req)
}
