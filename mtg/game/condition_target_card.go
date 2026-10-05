package game

// CardReference indexes only card slots; ObjectReference indexes all slots.
func targetCardConditionSlotMatches(card CardReference, object ObjectReference, targets []TargetSpec) bool {
	slot := object.TargetIndex()
	cardSlot := 0
	for i := range targets {
		target := &targets[i]
		allowsCard := targetSpecAllowedKinds(target)&TargetAllowCard != 0
		if slot < target.MaxTargets {
			return allowsCard && card.TargetIndex == cardSlot+slot
		}
		slot -= target.MaxTargets
		if allowsCard {
			cardSlot += target.MaxTargets
		}
	}
	return false
}
