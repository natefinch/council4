package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveResolutionPaymentValue(g *game.Game, obj *game.StackObject, res *game.ResolutionPayment, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (accepted, succeeded bool) {
	resolved, ok := materializeResolutionPayment(g, obj, nil, res)
	if !ok {
		return false, false
	}
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

func materializeResolutionPayment(g *game.Game, obj *game.StackObject, source *game.Permanent, res *game.ResolutionPayment) (game.ResolutionPayment, bool) {
	if res == nil {
		return game.ResolutionPayment{}, true
	}
	resolved := *res
	if res.ManaCost.Exists {
		resolved.ManaCost.Val = slices.Clone(res.ManaCost.Val)
	}
	resolved.AdditionalCosts = slices.Clone(res.AdditionalCosts)
	switch {
	case res.ManaCostMultiplier.Exists && res.ManaCostMultiplier.Val != nil && res.ManaCost.Exists:
		amount, ok := resolutionPaymentDynamicAmountValue(g, obj, source, res.ManaCostMultiplier.Val)
		if !ok {
			return resolved, false
		}
		amount = max(0, amount)
		resolved.ManaCost = opt.Val(res.ManaCost.Val.Multiply(amount))
		resolved.Prompt = "Pay " + resolved.ManaCost.Val.String() + "?"
		resolved.ManaCostMultiplier = opt.V[*game.DynamicAmount]{}
	case res.DynamicGenericManaCost.Exists && res.DynamicGenericManaCost.Val != nil:
		amount, ok := resolutionPaymentDynamicAmountValue(g, obj, source, res.DynamicGenericManaCost.Val)
		if !ok {
			return resolved, false
		}
		amount = max(0, amount)
		resolved.ManaCost = opt.Val(cost.Mana{cost.O(amount)})
		resolved.Prompt = "Pay " + resolved.ManaCost.Val.String() + "?"
		resolved.DynamicGenericManaCost = opt.V[*game.DynamicAmount]{}
	default:
	}
	return resolved, true
}

func resolutionPaymentDynamicAmountValue(g *game.Game, obj *game.StackObject, source *game.Permanent, dynamic *game.DynamicAmount) (int, bool) {
	if source != nil {
		return enterBattlefieldResolutionPaymentDynamicAmountValue(g, obj, source, dynamic)
	}
	controller := stackObjectController(obj)
	if obj == nil ||
		dynamic.Kind != game.DynamicAmountObjectPower ||
		dynamic.Object != game.SourcePermanentReference() {
		return dynamicAmountValue(g, obj, controller, *dynamic), true
	}
	resolved, ok := resolvePermanentOrLastKnown(g, obj.SourceID)
	if !ok {
		return 0, true
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return resolvedObjectPower(g, &resolved) * multiplier, true
}

func enterBattlefieldResolutionPaymentDynamicAmountValue(g *game.Game, obj *game.StackObject, source *game.Permanent, dynamic *game.DynamicAmount) (int, bool) {
	switch dynamic.Kind {
	case game.DynamicAmountConstant,
		game.DynamicAmountX,
		game.DynamicAmountControllerLife,
		game.DynamicAmountControllerHandSize,
		game.DynamicAmountControllerGraveyardSize,
		game.DynamicAmountControllerBasicLandTypeCount,
		game.DynamicAmountOpponentCount:
		return dynamicAmountValue(g, obj, obj.Controller, *dynamic), true
	case game.DynamicAmountObjectManaValue,
		game.DynamicAmountObjectCounters:
		if dynamic.Object != game.SourcePermanentReference() {
			return 0, false
		}
	default:
		return 0, false
	}

	resolved := resolvedObjectReference{permanent: source}
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountObjectManaValue:
		amount = resolvedObjectManaValue(g, &resolved)
	case game.DynamicAmountObjectCounters:
		amount = source.Counters.Get(dynamic.CounterKind)
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier, true
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
