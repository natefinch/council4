package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func transformManaPaymentTargets(payment game.ResolutionPayment, transform targetIndexTransform) (game.ResolutionPayment, bool) {
	if len(payment.AdditionalCosts) != 0 {
		return game.ResolutionPayment{}, false
	}
	if payment.Payer.Exists {
		payer, ok := transformPlayerReference(payment.Payer.Val, transform)
		if !ok {
			return game.ResolutionPayment{}, false
		}
		payment.Payer = opt.Val(payer)
	}
	var ok bool
	payment.DynamicGenericManaCost, ok = transformPaymentAmount(payment.DynamicGenericManaCost, transform)
	if !ok {
		return game.ResolutionPayment{}, false
	}
	payment.ManaCostMultiplier, ok = transformPaymentAmount(payment.ManaCostMultiplier, transform)
	return payment, ok
}

func transformPaymentAmount(amount opt.V[*game.DynamicAmount], transform targetIndexTransform) (opt.V[*game.DynamicAmount], bool) {
	if !amount.Exists {
		return amount, true
	}
	if amount.Val == nil {
		return opt.V[*game.DynamicAmount]{}, false
	}
	value := *amount.Val
	if objectReferenceCarriesTargetIndex(value.Object) {
		object, ok := transformObjectReference(value.Object, transform)
		if !ok {
			return opt.V[*game.DynamicAmount]{}, false
		}
		value.Object = object
	}
	if value.Player != nil {
		player, ok := transformPlayerReference(*value.Player, transform)
		if !ok {
			return opt.V[*game.DynamicAmount]{}, false
		}
		value.Player = &player
	}
	if value.Group.Valid() {
		group, ok := transformGroupReference(value.Group, transform)
		if !ok {
			return opt.V[*game.DynamicAmount]{}, false
		}
		value.Group = group
	}
	return opt.Val(&value), true
}
