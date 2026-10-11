package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

func combatStunPossessiveBindsRelated(reference CompiledReference, preceding []CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIts {
		return false
	}
	for _, effect := range effects {
		if effect.Kind != EffectUntap || !effect.Negated || !effect.Exact ||
			effect.Context != parser.EffectContextReferencedObject || len(effect.Targets) != 0 ||
			len(effect.SubjectReferences) != 1 || len(effect.References) != 2 ||
			!effectOwnsReference(effect, reference) || reference.Subject.scope != effect.subjectScope {
			continue
		}
		subject := effect.SubjectReferences[0]
		for _, antecedent := range preceding {
			if antecedent.NodeID == subject.NodeID && antecedent.Kind == ReferenceThatObject &&
				antecedent.Binding == ReferenceBindingEventRelatedPermanent &&
				antecedent.Order.End <= reference.Order.Start && effectOwnsReference(effect, antecedent) &&
				antecedent.Subject.scope == effect.subjectScope {
				return true
			}
		}
	}
	return false
}
