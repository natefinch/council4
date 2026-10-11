package payment

import "github.com/natefinch/council4/mtg/game/counter"

func reserveStunForAdditionalCosts(restrictions *manaSourceRestrictions, additional additionalCostPlan) {
	for key, amount := range plannedCounterRemovalsBySourceKind(additional.counterRemovals) {
		if key.kind == counter.Stun && key.source.Tapped && amount >= key.source.Counters.Get(counter.Stun) {
			restrictions.excluded[key.source.ObjectID] = true
		}
	}
}

// Mana abilities resolve before additional costs are paid (CR 601.2g).
func stunCounterPaymentsCompatible(manaPlan paymentPlan, additional additionalCostPlan) bool {
	for key, amount := range plannedCounterRemovalsBySourceKind(additional.counterRemovals) {
		if key.kind != counter.Stun {
			continue
		}
		available := key.source.Counters.Get(counter.Stun)
		for _, activation := range manaPlan.manaTaps {
			if activation.permanent == key.source && activation.untap && available > 0 {
				available--
			}
		}
		if available < amount {
			return false
		}
	}
	return true
}
