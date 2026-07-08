package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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
// canPaySpellCosts reports whether the player could pay the full cost of casting
// the spell described by req, without paying it. It is the feasibility check for
// the pay-the-cost step of casting: a total cost that can't be paid in full makes
// the cast illegal and is rewound (CR 601.2h, CR 118.3).
func (o paymentOrchestratorType) canPaySpellCosts(g *game.Game, req payment.SpellRequest) bool {
	return o.planner(g).CanPaySpellCosts(req)
}

// paySpellCosts pays all spell costs described by req and returns the payment
// details, including the selected casting permission. This is the "pay the total
// cost" step of casting (CR 601.2h): the locked-in total cost (CR 601.2f) is paid
// in full, with mana abilities already available to have been activated
// (CR 601.2g); partial payments are not allowed.
func (o paymentOrchestratorType) paySpellCosts(g *game.Game, req payment.SpellRequest) (payment.SpellPaymentResult, bool) {
	return o.planner(g).PaySpellCosts(req)
}

// buildAbilityCostPlan reports whether a plan can be built for the ability
// described by req, without applying it.
func (o paymentOrchestratorType) buildAbilityCostPlan(g *game.Game, req payment.AbilityRequest) bool {
	return o.planner(g).BuildAbilityCostPlan(req)
}

// abilityCostPayment carries the identities of objects and cards consumed to pay
// an ability's activation cost so the caller can record them on the resolving
// stack object: the object IDs of permanents sacrificed and the card-instance
// IDs of cards exiled from a zone.
type abilityCostPayment struct {
	sacrificedIDs []id.ID
	exiledIDs     []id.ID
}

// payAbilityCosts pays all ability costs described by req. Paying an ability cost
// is never a spell cast, so any tagged mana-spend rider units it consumes are
// dropped without firing, keeping rider provenance exact for later payments. It
// returns the object IDs of any permanents sacrificed and the card-instance IDs
// of any cards exiled as a cost so the caller can record them on the resolving
// stack object.
func (o paymentOrchestratorType) payAbilityCosts(g *game.Game, req payment.AbilityRequest) (abilityCostPayment, bool) {
	before, hasRiders := manaSpendRiderSnapshot(g, req.PlayerID)
	paid, ok := o.planner(g).PayAbilityCosts(req)
	if !ok {
		return abilityCostPayment{}, false
	}
	if hasRiders {
		consumeManaSpendRidersForPayment(g, req.PlayerID, req.Source, before, paid.PoolSpend)
	}
	return abilityCostPayment{sacrificedIDs: paid.SacrificedIDs, exiledIDs: paid.ExiledIDs}, true
}

// canPayGenericCost reports whether the player can pay the mana cost described by req.
func (o paymentOrchestratorType) canPayGenericCost(g *game.Game, req payment.GenericRequest) bool {
	return o.planner(g).CanPayGenericCost(req)
}

// payGenericCostForSpell pays a mana cost that is part of casting a spell (such
// as a madness cost) and returns the per-unit pool mana consumed so the caller
// can resolve mana-spend riders as a spell cast after the spell is on the stack.
// Unlike payGenericCost it does not itself consume rider units, because the
// payment is a spell cast and any tagged mana spent must be evaluated against
// the qualifying spell rather than dropped without firing.
func (o paymentOrchestratorType) payGenericCostForSpell(g *game.Game, req payment.GenericRequest) (poolSpend map[mana.Unit]int, ok bool) {
	return o.planner(g).PayGenericCost(req)
}

// payGenericCost builds, validates, and applies the mana cost described by req.
// A generic cost is never a spell cast, so any tagged mana-spend rider units it
// consumes are dropped without firing, keeping rider provenance exact for later
// payments.
func (o paymentOrchestratorType) payGenericCost(g *game.Game, req payment.GenericRequest) bool {
	before, hasRiders := manaSpendRiderSnapshot(g, req.PlayerID)
	poolSpend, ok := o.planner(g).PayGenericCost(req)
	if !ok {
		return false
	}
	if hasRiders {
		consumeManaSpendRidersForPayment(g, req.PlayerID, nil, before, poolSpend)
	}
	return true
}
