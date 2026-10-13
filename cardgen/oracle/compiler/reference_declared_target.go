package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

// A clause's exact owner destination cannot read an unrelated earlier card
// result. Reconcile only that declared subject before issuing lifetime proof.
func bindDeclaredTargetOwnerReferences(content *AbilityContent) {
	for i := range content.References {
		reference := &content.References[i]
		if reference.Binding != ReferenceBindingPriorInstructionResult ||
			reference.PriorInstruction < 0 || reference.PriorInstruction >= len(content.Effects) {
			continue
		}
		prior := content.Effects[reference.PriorInstruction]
		if !prior.Exact || prior.Kind != EffectExile || len(prior.Targets) != 1 ||
			!prior.Targets[0].Exact || targetSubjectDomain(prior.Targets[0]) != ReferenceSubjectCard ||
			prior.Order.End > reference.Order.Start {
			continue
		}
		if occurrence, ok := declaredTargetOwnerOccurrence(*content, *reference); ok {
			reference.Binding = ReferenceBindingTarget
			reference.Occurrence = occurrence
		}
	}
}

func declaredTargetOwnerOccurrence(content AbilityContent, reference CompiledReference) (int, bool) {
	if content.subjectScope == nil || content.subjectScope.owner == nil ||
		content.Source.scope != content.subjectScope ||
		reference.Subject.scope != nil && reference.Subject.scope != content.subjectScope ||
		reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIts ||
		reference.SubjectNoun != parser.ObjectNounUnknown || reference.CardIdentity ||
		reference.PaidCost != nil || reference.LibraryCardObservation || reference.ProducerClauseID != 0 ||
		reference.ReachedCardProducerClauseID != 0 ||
		reference.Order.End <= reference.Order.Start {
		return 0, false
	}
	owner := content.subjectScope.owner
	canonicalCount := 0
	for _, canonical := range owner.References {
		if canonical.NodeID == reference.NodeID {
			if canonical.Order != reference.Order || canonical.Kind != reference.Kind ||
				canonical.Pronoun != reference.Pronoun || canonical.SubjectNoun != reference.SubjectNoun ||
				canonical.CardIdentity != reference.CardIdentity ||
				canonical.ProducerClauseID != reference.ProducerClauseID ||
				canonical.LibraryCardObservation != reference.LibraryCardObservation ||
				canonical.ReachedCardProducerClauseID != reference.ReachedCardProducerClauseID ||
				canonical.PaidCost != nil {
				return 0, false
			}
			canonicalCount++
		}
	}
	if canonicalCount != 1 {
		return 0, false
	}
	var clause *CompiledEffect
	for i := range owner.Effects {
		effect := &owner.Effects[i]
		for _, owned := range effect.References {
			if owned.NodeID != reference.NodeID {
				continue
			}
			if clause != nil || owned.Order != reference.Order || owned.Kind != reference.Kind ||
				owned.Pronoun != reference.Pronoun || owned.SubjectNoun != reference.SubjectNoun ||
				owned.CardIdentity != reference.CardIdentity || owned.ProducerClauseID != 0 ||
				owned.LibraryCardObservation || owned.ReachedCardProducerClauseID != 0 || owned.PaidCost != nil ||
				effect.subjectScope != content.subjectScope || effect.ClauseID <= 0 ||
				!effect.Order.Contains(reference.Order) {
				return 0, false
			}
			clause = effect
		}
	}
	if clause == nil || !clause.Exact || clause.Kind != EffectReturn ||
		clause.Context != parser.EffectContextController || clause.FromZone != zone.None ||
		clause.ToZone != zone.Hand || clause.Negated || clause.Optional || clause.DelayedTiming != 0 ||
		len(clause.Targets) != 1 || len(clause.References) != 1 || len(clause.SubjectTargets) != 0 ||
		len(clause.SubjectReferences) != 0 {
		return 0, false
	}
	clauses := 0
	for _, effect := range owner.Effects {
		if effect.ClauseID == clause.ClauseID {
			clauses++
		}
	}
	if clauses != 1 {
		return 0, false
	}
	declaration := clause.Targets[0]
	if !declaration.Exact || declaration.subjectScope != content.subjectScope ||
		declaration.Cardinality != (TargetCardinality{Min: 1, Max: 1}) ||
		declaration.Order.End <= declaration.Order.Start ||
		!clause.Order.Contains(declaration.Order) || declaration.Order.End > reference.Order.Start ||
		targetSubjectDomain(declaration) != ReferenceSubjectPermanent {
		return 0, false
	}
	for _, alternative := range declaration.Selector.Alternatives {
		if len(alternative.Alternatives) != 0 ||
			targetSubjectDomain(CompiledTarget{Selector: alternative}) != ReferenceSubjectPermanent {
			return 0, false
		}
	}
	occurrence := -1
	for i, target := range owner.Targets {
		if target.Order != declaration.Order {
			continue
		}
		if occurrence >= 0 || !target.Exact || target.subjectScope != content.subjectScope ||
			target.Cardinality != declaration.Cardinality ||
			target.Selector.Kind != declaration.Selector.Kind || target.Selector.Zone != declaration.Selector.Zone {
			return 0, false
		}
		occurrence = i
	}
	return occurrence, occurrence >= 0
}
