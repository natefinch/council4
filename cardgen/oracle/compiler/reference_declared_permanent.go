package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func bindDeclaredPermanentSubject(reference *CompiledReference, targets []CompiledTarget, effects []CompiledEffect, source ReferenceSourceContext) bool {
	reference.Binding = ReferenceBindingUnsupported
	if source.scope == nil || source.scope.owner == nil {
		return false
	}
	occurrence, ok := declaredPermanentOccurrence(*reference, targets, effects, source.scope)
	if !ok {
		return false
	}
	reference.Binding, reference.Occurrence = ReferenceBindingTarget, occurrence
	return true
}

func declaredPermanentOccurrence(reference CompiledReference, targets []CompiledTarget, effects []CompiledEffect, scope *referenceSubjectScope) (int, bool) {
	if reference.DeclaredPermanent == nil || reference.DeclaredPermanent.ClauseID <= 0 ||
		reference.ProducerClauseID != 0 || reference.LibraryCardObservation || reference.CardIdentity ||
		reference.ReachedCardProducerClauseID != 0 ||
		reference.Order.End <= reference.Order.Start ||
		(reference.Kind != ReferenceThatObject || reference.Pronoun != ReferencePronounUnknown) &&
			(reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIt && reference.Pronoun != ReferencePronounIts) ||
		reference.SubjectNoun == parser.ObjectNounCard || reference.SubjectNoun == parser.ObjectNounSpell ||
		reference.SubjectNoun == parser.ObjectNounPlayer {
		return 0, false
	}
	owner := scope.owner
	canonical := 0
	for _, candidate := range owner.References {
		if candidate.NodeID == reference.NodeID {
			if candidate.DeclaredPermanent == nil || *candidate.DeclaredPermanent != *reference.DeclaredPermanent ||
				candidate.ProducerClauseID != 0 || candidate.LibraryCardObservation || candidate.CardIdentity ||
				candidate.ReachedCardProducerClauseID != 0 ||
				candidate.Order != reference.Order || candidate.Kind != reference.Kind ||
				candidate.Pronoun != reference.Pronoun || candidate.SubjectNoun != reference.SubjectNoun {
				return 0, false
			}
			canonical++
		}
	}
	if canonical != 1 {
		return 0, false
	}
	var declaration *CompiledEffect
	consumers := 0
	permanents := 0
	for i := range effects {
		effect := &effects[i]
		if effect.Order.End <= reference.Order.Start {
			for _, target := range effect.Targets {
				if targetSubjectDomain(target) == ReferenceSubjectPermanent {
					permanents++
				}
			}
		}
		if effectOwnsReference(*effect, reference) && effect.Order.Contains(reference.Order) {
			if effect.ClauseID <= 0 || effect.subjectScope != scope ||
				(effect.Context != parser.EffectContextReferencedObjectOwner &&
					effect.Context != parser.EffectContextReferencedObjectController) ||
				len(effect.References) == 0 || effect.References[0].NodeID != reference.NodeID ||
				reference.Subject.scope != nil && (reference.Subject.declaredConsumerClauseID != effect.ClauseID ||
					reference.Subject.declaredConsumerOrder != effect.Order ||
					reference.Subject.declaredConsumerContext != effect.Context) {
				return 0, false
			}
			owned := 0
			for _, candidate := range effect.References {
				if candidate.NodeID == reference.NodeID {
					if candidate.Order != reference.Order || candidate.Kind != reference.Kind ||
						candidate.Pronoun != reference.Pronoun || candidate.DeclaredPermanent == nil ||
						candidate.ProducerClauseID != 0 || candidate.LibraryCardObservation || candidate.CardIdentity ||
						candidate.ReachedCardProducerClauseID != 0 ||
						*candidate.DeclaredPermanent != *reference.DeclaredPermanent {
						return 0, false
					}
					owned++
				}
			}
			if owned != 1 {
				return 0, false
			}
			consumers++
		}
		if effect.ClauseID == reference.DeclaredPermanent.ClauseID {
			if declaration != nil || effect.subjectScope != scope {
				return 0, false
			}
			declaration = effect
		}
	}
	if permanents != 1 || consumers != 1 || declaration == nil || !declaration.Exact || declaration.Negated ||
		declaration.DelayedTiming != 0 || len(declaration.Targets) != 1 ||
		declaration.Order == (shared.SourceOrder{}) || declaration.Order.End > reference.Order.Start {
		return 0, false
	}
	if reference.Subject.scope != nil &&
		(reference.Subject.declaredPermanentOrder != declaration.Order ||
			reference.Subject.producerKind != declaration.Kind ||
			reference.Subject.producerFromZone != declaration.FromZone ||
			reference.Subject.producerToZone != declaration.ToZone ||
			reference.Subject.producerCardSource != declaration.CardSource) {
		return 0, false
	}
	target := declaration.Targets[0]
	if target.subjectScope != scope || !target.Exact || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
		target.Order.End <= target.Order.Start || !declaration.Order.Contains(target.Order) ||
		target.Order.End > reference.Order.Start || targetSubjectDomain(target) != ReferenceSubjectPermanent {
		return 0, false
	}
	found, occurrence := 0, 0
	for i, candidate := range targets {
		if candidate.Order == target.Order {
			if candidate.subjectScope != scope || !candidate.Exact || candidate.Selector.Kind != target.Selector.Kind ||
				candidate.Selector.Zone != target.Selector.Zone || candidate.Cardinality != target.Cardinality {
				return 0, false
			}
			found++
			occurrence = i
		}
	}
	return occurrence, found == 1
}

func declaredPermanentProofMatches(reference CompiledReference) bool {
	proof := reference.Subject
	if reference.DeclaredPermanent == nil {
		return proof.declaredPermanentClauseID == 0
	}
	if proof.declaredPermanentClauseID <= 0 || reference.DeclaredPermanent.ClauseID != proof.declaredPermanentClauseID ||
		proof.scope == nil || proof.scope.owner == nil || reference.Binding != ReferenceBindingTarget {
		return false
	}
	owner := proof.scope.owner
	occurrence, ok := declaredPermanentOccurrence(reference, owner.Targets, owner.Effects, proof.scope)
	if !ok || occurrence != proof.targetOccurrence {
		return false
	}
	target := owner.Targets[occurrence]
	return proof.targetOrder == target.Order && proof.targetSelector == target.Selector.Kind &&
		proof.targetZone == target.Selector.Zone && proof.targetMin == target.Cardinality.Min &&
		proof.targetMax == target.Cardinality.Max
}
