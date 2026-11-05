package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// activationCost is the cost of activating an ability with its targets, modes
// and X fixed (CR 602.2b, 601.2f). Its checks take the game to check in, and
// find the ability's source there again by identity, so a clone in which
// payment-time mana abilities were activated can be checked too. The zero
// value is an activation that isn't legal.
type activationCost struct {
	// includesMana reports whether the total cost includes a mana payment, so
	// mana abilities may be activated while paying it (CR 601.2g).
	includesMana bool
	// payable reports whether the whole cost can be paid from current
	// resources; nonManaPayable whether all but its mana cost can.
	payable, nonManaPayable func(*game.Game) bool
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
		includesMana: activationCostIncludesMana(manaCost, req.XValue),
		payable: func(g *game.Game) bool {
			req, ok := request(g)
			return ok && paymentOrch.buildAbilityCostPlan(g, req)
		},
		nonManaPayable: func(g *game.Game) bool {
			req, ok := request(g)
			return ok && paymentOrch.canPayAbilityNonManaCosts(g, req)
		},
	}
}

// genericRequestCost is the activation cost whose locked mana cost req
// describes (CR 601.2f). Its other component, such as discarding the source
// card, can be paid while present reports the source still where the ability
// functions.
func genericRequestCost(req payment.GenericRequest, present func(*game.Game) bool) activationCost {
	return activationCost{
		includesMana: activationCostIncludesMana(req.Cost, req.XValue),
		payable: func(g *game.Game) bool {
			return present(g) && paymentOrch.canPayGenericCost(g, req)
		},
		nonManaPayable: func(g *game.Game) bool {
			return present(g) && len(req.AdditionalCosts) == 0
		},
	}
}

// activationPermanentPresent reports whether the permanent objectID is on the
// battlefield in the game checked.
func activationPermanentPresent(objectID id.ID) func(*game.Game) bool {
	return func(g *game.Game) bool {
		_, ok := permanentByObjectID(g, objectID)
		return ok
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

// activationCostIncludesMana reports whether manaCost, with X announced as
// xValue, includes a mana payment. An activation announces no Phyrexian
// payments, so each Phyrexian symbol may still be paid with mana.
func activationCostIncludesMana(manaCost *cost.Mana, xValue int) bool {
	return xValue > 0 && costHasVariableMana(manaCost) || manaCostIncludesManaPayment(manaCost, nil)
}

// activationFeasible reports whether player may begin activating an ability
// whose cost is c: c is legal and payable from current resources, or its cost
// includes a mana payment, everything but that mana can be paid now, and the
// payment-time mana abilities player may activate while paying it can make it
// payable (CR 602.2b, 601.2g-h, 605.3a; see manualActivationManaMayPay). An
// activation begun that can't then be paid is reversed (CR 733.1).
func (e *Engine) activationFeasible(g *game.Game, player game.PlayerID, c activationCost) bool {
	if c.payable == nil {
		return false
	}
	if c.payable(g) {
		return true
	}
	return c.includesMana && c.nonManaPayable(g) && e.hasManualPaymentManaAbility(g, player) &&
		e.manualActivationManaMayPay(g, player, c.payable)
}

// activationPaymentManaWindow gives player the chance to activate mana
// abilities once an activation's total cost c is locked, when it includes a
// mana payment (CR 602.2b, 601.2g). It returns nil when no window opened, and
// otherwise the undo that reverses the reversible mana abilities activated if
// the activation then goes unpaid (CR 733.1).
func (e *Engine) activationPaymentManaWindow(g *game.Game, player game.PlayerID, c activationCost, agents [game.NumPlayers]PlayerAgent, log *TurnLog) func() {
	if !c.includesMana || !e.hasManualPaymentManaAbility(g, player) {
		return nil
	}
	return e.activateManualPaymentManaAbilities(g, player, agents, log)
}

// reverseActivation reverses an activation whose cost went unpaid after
// payment-time mana abilities were activated for it: undo reverses those
// abilities (CR 733.1), and the player keeps priority (CR 733.2).
func reverseActivation(undo func()) actionOutcome {
	undo()
	return actionReversed
}

// activationAttemptKey identifies an activation an attempt redoes.
func activationAttemptKey(activate action.ActivateAbilityAction) string {
	return fmt.Sprintf("activate|%+v", activate)
}

// reverseActivationAttempt records an activation reversed while its player
// holds priority. They may redo it (CR 733.2), but an attempt repeating one
// already reversed exhausts it, so it's no longer offered until an action is
// taken and a deterministic player always moves on.
func (a *castAttempts) reverseActivationAttempt(activate action.ActivateAbilityAction) {
	a.reverseAttempt(activationAttemptKey(activate))
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
