package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// activationCost is the cost of activating an ability with its targets, modes
// and X fixed (CR 602.2b, 601.2f). The request retains IDs, not mutable source
// pointers. Feasibility is always a joint Complete query.
type activationCost struct {
	request payment.Request
	present func(*game.Game) bool
}

// abilityRequestCost is the activation cost req describes to the ability-cost
// planner. The total is locked by the shared procedure (CR 601.2f);
// source finds the ability's source again and
// reports the permanent req.Source names in the game checked, or nil for a card
// source, and whether the source is still where the ability functions.
func abilityRequestCost(req payment.AbilityRequest, source func(*game.Game) (*game.Permanent, bool)) activationCost {
	return activationCost{
		request: payment.ForAbility(req),
		present: func(g *game.Game) bool { _, ok := source(g); return ok },
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
// activation or resource heuristic.
func (e *Engine) costedActionAvailable(g *game.Game, player game.PlayerID, c activationCost) bool {
	if c.present == nil || !c.present(g) {
		return false
	}
	_, ok, _ := e.PaymentPlanner(g).Complete(c.request, payment.Partial{})
	return ok
}
