package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// activationCost is the cost of activating an ability with its targets, modes
// and X fixed (CR 602.2b, 601.2f). The request retains IDs, not mutable source
// pointers. payable supports legacy current-resource checks used by casts.
type activationCost struct {
	request payment.Request
	present func(*game.Game) bool
	payable func(*game.Game) bool
}

// payableNow reports whether c is legal and payable from current resources.
func (c activationCost) payableNow(g *game.Game) bool {
	return c.payable != nil && c.payable(g)
}

// abilityRequestCost is the activation cost req describes to the ability-cost
// planner, with manaCost its total mana cost. The total is locked as it is now
// (CR 601.2f); a check finds the ability's source again with source, which
// reports the permanent req.Source names in the game checked, or nil for a card
// source, and whether the source is still where the ability functions.
func abilityRequestCost(manaCost *cost.Mana, req payment.AbilityRequest, source func(*game.Game) (*game.Permanent, bool)) activationCost {
	request := func(g *game.Game) (payment.AbilityRequest, bool) {
		permanent, ok := source(g)
		if !ok {
			return req, false
		}
		locked := req
		if locked.Source != nil {
			locked.Source = permanent
		}
		return locked, true
	}
	return activationCost{
		request: payment.ForAbility(req),
		present: func(g *game.Game) bool { _, ok := source(g); return ok },
		payable: func(g *game.Game) bool {
			req, ok := request(g)
			return ok && paymentOrch.buildAbilityCostPlan(g, req)
		},
	}
}

// genericRequestCost is the activation cost whose locked mana cost req
// describes (CR 601.2f). Its other component, such as discarding the source
// card, can be paid while present reports the source still where the ability
// functions.
func genericRequestCost(req payment.GenericRequest, present func(*game.Game) bool) activationCost {
	return activationCost{
		request: payment.ForGeneric(req),
		present: present,
		payable: func(g *game.Game) bool {
			return present(g) && paymentOrch.canPayGenericCost(g, req)
		},
	}
}

// cardSourcePresent reports whether the card cardID still has the ability
// abilityIndex that find looks for in the zone it functions from.
func cardSourcePresent(find func(*game.Game, game.PlayerID, id.ID, int) (*game.CardInstance, game.ActivatedAbility, bool), player game.PlayerID, cardID id.ID, abilityIndex int) func(*game.Game) bool {
	return func(g *game.Game) bool {
		_, _, ok := find(g, player, cardID, abilityIndex)
		return ok
	}
}

// cardSource adapts present for abilityRequestCost: a card source has no
// permanent.
func cardSource(present func(*game.Game) bool) func(*game.Game) (*game.Permanent, bool) {
	return func(g *game.Game) (*game.Permanent, bool) {
		return nil, present(g)
	}
}

// permanentSource finds the permanent objectID in the game checked.
func permanentSource(objectID id.ID) func(*game.Game) (*game.Permanent, bool) {
	return func(g *game.Game) (*game.Permanent, bool) {
		return permanentByObjectID(g, objectID)
	}
}

// costedActionAvailable requires a joint completion, never a speculative
// activation, resource heuristic or inconclusive clone probe.
func (e *Engine) costedActionAvailable(g *game.Game, player game.PlayerID, c activationCost) bool {
	if c.present == nil || !c.present(g) {
		return false
	}
	_, ok, _ := e.PaymentPlanner(g).Complete(c.request, payment.Partial{})
	return ok
}

// reverseAttempt records the action key identifies as reversed, exhausting it
// when it was already reversed.
func (a *castAttempts) reverseAttempt(key string) {
	if a == nil {
		return
	}
	if a.reversed[key] {
		if a.exhausted == nil {
			a.exhausted = map[string]bool{}
		}
		a.exhausted[key] = true
		return
	}
	if a.reversed == nil {
		a.reversed = map[string]bool{}
	}
	a.reversed[key] = true
}
