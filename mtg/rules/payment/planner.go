package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/cost"
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

// SpellCostOptions returns every cost option req offers the caster, whether or
// not it is currently payable: the normal cost and each applicable alternative
// cost (flashback, escape, evoke, a route's alternative), with their additional
// costs, unresolved X symbols and no cost modifiers applied (CR 601.2b, 601.2f).
func (p Planner) SpellCostOptions(req SpellRequest) []SpellOptionSummary {
	options := spellCostOptionsForRequest(p.s, req)
	result := make([]SpellOptionSummary, 0, len(options))
	for _, option := range options {
		result = append(result, SpellOptionSummary{
			Index:           option.index,
			Label:           option.label,
			ManaCost:        option.manaCost,
			AdditionalCosts: option.additionalCosts,
			CastPermission:  option.castPermission,
		})
	}
	return result
}

// PrepareSpellCosts locks each available total before mana abilities change
// cost-modifying sources. Unlike PayableSpellOptions, it includes costs that
// may become payable by activating a manual mana ability.
func (p Planner) PrepareSpellCosts(req SpellRequest) []PreparedSpellCost {
	var result []PreparedSpellCost
	for _, option := range spellCostOptionsForRequest(p.s, req) {
		resolved, ok := resolveManaCostX(option.manaCost, req.XValue)
		if !ok {
			continue
		}
		option.manaCost = resolved
		option = applyCostModifiers(p.s, costModificationContext{
			player: req.PlayerID, card: option.card, cardID: req.CardID, sourceZone: req.SourceZone,
			targets: req.Targets, bargained: option.bargained, bestowed: req.Bestowed, option: option,
		})
		option.additionalCosts = slices.Clone(option.additionalCosts)
		lockedCosts := option.additionalCosts[:0]
		for _, additional := range option.additionalCosts {
			if additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
				additional.Amount = p.s.AdditionalDynamicAmountValue(req.PlayerID, additional.AmountDynamic)
				additional.AmountDynamic = cost.AdditionalDynamicAmountNone
				additional.AmountFromX = false
				if additional.Amount == 0 {
					continue
				}
			}
			lockedCosts = append(lockedCosts, additional)
		}
		option.additionalCosts = lockedCosts
		option.locked = true
		result = append(result, PreparedSpellCost{
			Summary: SpellOptionSummary{Index: option.index, Label: option.label,
				ManaCost: option.manaCost, AdditionalCosts: option.additionalCosts, CastPermission: option.castPermission},
			option: option,
		})
	}
	return result
}

// CanPayPreparedNonManaCosts reports whether the non-mana components of
// req.PreparedCost (sacrifices, discards, life and the other additional costs)
// can be paid from current resources. Mana abilities activated while paying
// (CR 601.2g) only add mana, so a total whose non-mana costs fail this check
// cannot become payable by activating them.
func (p Planner) CanPayPreparedNonManaCosts(req SpellRequest) bool {
	if req.PreparedCost == nil {
		return false
	}
	option := req.PreparedCost.option
	_, ok := buildAdditionalCostPlanForCosts(p.s, req.PlayerID, option.additionalCosts, req.XValue, nil, nil, req.CardID, req.SourceZone, req.CardID)
	return ok
}

// CanPaySpellCostsAsPreferred reports whether the player can currently pay req
// as req.Prefs directs, honoring its cost option and its Phyrexian life and
// mana choices, without applying the payment. With nil Prefs it is
// CanPaySpellCosts.
func (p Planner) CanPaySpellCostsAsPreferred(req SpellRequest) bool {
	_, ok := buildSpellCostPlan(p.s, req)
	return ok
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

// CanPaySpellManaCost reports whether the player could pay req.Cost as the
// total mana cost of casting req.Card, without paying it. The cost is paid as
// the spell planner pays a cast's mana: in the spell's spending context, and
// with the convoke, improvise or delve payment the spell may use once its total
// cost is determined (CR 601.2f-h, 702.51a, 702.66a, 702.126a). It bounds a
// cast before its targets and other choices are made, so req.Cost must already
// include any cost reduction the bound allows. Restricted mana is tested
// against the spell's effective characteristics, as the cast pays it.
func (p Planner) CanPaySpellManaCost(req SpellManaRequest) bool {
	option := spellCostOption{
		card:       req.Card,
		manaCost:   req.Cost,
		spendAs:    req.SpendAs,
		spendSpell: p.s.SpellCharacteristics(req.PlayerID, req.Card, req.Bestowed),
	}
	_, ok := buildSpellManaPlanForOption(p.s, req.PlayerID, req.CardID, req.SourceZone, option, 0, manaSourceRestrictions{}, nil)
	return ok
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
