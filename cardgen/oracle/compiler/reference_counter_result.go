package compiler

import "slices"

func bindCounterResultReferences(conditions []CompiledCondition, references []CompiledReference, targets []CompiledTarget, effects []CompiledEffect) {
	for _, condition := range conditions {
		if condition.Predicate != ConditionPredicateCounterSucceeded {
			continue
		}
		ownership := condition.Ownership
		occurrence := ownership.ResultSubjectTargetOccurrence
		valid := ownership.ResultProducerClauseID > 0 && ownership.ResultSubjectReferenceNodeID >= 0 &&
			slices.Contains(ownership.ReferenceNodeIDs, ownership.ResultSubjectReferenceNodeID) &&
			occurrence >= 0 && occurrence < len(targets)
		producers := 0
		if valid {
			for _, effect := range effects {
				if effect.ClauseID != ownership.ResultProducerClauseID {
					continue
				}
				producers++
				valid = effect.Kind == EffectCounter && len(effect.Targets) == 1 &&
					effect.Targets[0].Order == targets[occurrence].Order &&
					targets[occurrence].Selector.Kind == SelectorSpell &&
					targets[occurrence].Cardinality.Min == 1 && targets[occurrence].Cardinality.Max == 1
				if !valid {
					break
				}
			}
		}
		for ri := range references {
			reference := &references[ri]
			if reference.NodeID != ownership.ResultSubjectReferenceNodeID {
				continue
			}
			// This exact subject has authoritative parser ownership; it cannot
			// name an unrelated earlier product selected by the generic binder.
			reference.Binding = ReferenceBindingUnsupported
			if valid && producers == 1 && reference.Kind == ReferenceThatObject {
				reference.Binding, reference.Occurrence = ReferenceBindingTarget, occurrence
			}
		}
	}
}
