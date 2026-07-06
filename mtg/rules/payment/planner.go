package payment

import (
	"github.com/natefinch/council4/mtg/game/mana"
)

// Planner is the payment planner. It exposes the public entry points for cost
// checking and payment. Construction is via New; the zero value is not usable.
type Planner struct {
	s State
}

// New returns a Planner bound to the provided State implementation.
func New(s State) Planner {
	return Planner{s: s}
}

// CanPaySpellCosts reports whether the player can currently pay all costs for
// the spell described by req.
func (p Planner) CanPaySpellCosts(req SpellRequest) bool {
	return canPaySpellCosts(p.s, req)
}

// PaySpellCosts pays all spell costs described by req and returns the payment
// details, including the selected casting permission.
func (p Planner) PaySpellCosts(req SpellRequest) (SpellPaymentResult, bool) {
	return paySpellCosts(p.s, req)
}

// PayableSpellOptions returns the set of cost options that the player can currently pay,
// with enough detail for the Engine's choice layer to present and track them.
func (p Planner) PayableSpellOptions(req SpellRequest) []SpellOptionSummary {
	return payableSpellOptionsFromState(p.s, req)
}

// BuildAbilityCostPlan reports whether a payment plan can be built for the
// activated ability described by req without applying it.
func (p Planner) BuildAbilityCostPlan(req AbilityRequest) bool {
	_, ok := buildAbilityCostPlan(p.s, req)
	return ok
}

// PayAbilityCosts pays all costs for the activated ability described by req. It
// returns an AbilityCostPayment carrying the pool mana consumed (for mana-spend
// rider resolution), the object IDs of permanents sacrificed as a cost, and the
// card-instance IDs of cards exiled as a cost, plus a success flag.
func (p Planner) PayAbilityCosts(req AbilityRequest) (AbilityCostPayment, bool) {
	return payAbilityCosts(p.s, req)
}

// CanPayGenericCost reports whether the player can pay the mana (and optional
// additional) costs described by req.
func (p Planner) CanPayGenericCost(req GenericRequest) bool {
	return canPayGenericCost(p.s, req)
}

// PayGenericCost builds, validates, and applies the mana (and optional
// additional) costs described by req. It returns the per-unit amount of pool
// mana consumed (for mana-spend rider resolution) plus a success flag.
func (p Planner) PayGenericCost(req GenericRequest) (poolSpend map[mana.Unit]int, ok bool) {
	return payGenericCost(p.s, req)
}
