package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (content AbilityContent) TargetSubject(index int) (CompiledReference, bool) {
	if content.subjectScope == nil || index < 0 || index >= len(content.Targets) ||
		content.Targets[index].subjectScope != content.subjectScope {
		return CompiledReference{}, false
	}
	reference := CompiledReference{Binding: ReferenceBindingTarget, Occurrence: index}
	issueReferenceSubject(&reference, content.Targets, content.Effects, content.Source)
	if !reference.SubjectSupported() {
		return CompiledReference{}, false
	}
	return reference, true
}

// DelayedSubjectReference validates parser-owned scheduling facts at the same
// binding seam as immediate subjects. It never invents a reference NodeID.
func (content AbilityContent) DelayedSubjectReference(effect CompiledEffect, effects []CompiledEffect) (CompiledReference, bool) {
	if content.subjectScope == nil || effect.subjectScope != content.subjectScope ||
		effect.DelayedTiming == 0 {
		return CompiledReference{}, false
	}
	subject := effect.DelayedSubject
	var selected *CompiledReference
	for _, reference := range content.References {
		if !slices.Contains(subject.ReferenceNodeIDs, reference.NodeID) {
			continue
		}
		if !content.OwnsSubject(reference) || selected != nil &&
			(selected.Binding != reference.Binding || selected.ProducerClauseID != reference.ProducerClauseID ||
				selected.Occurrence != reference.Occurrence || selected.SubjectDomain() != reference.SubjectDomain()) {
			return CompiledReference{}, false
		}
		copy := reference
		selected = &copy
	}
	switch subject.Kind {
	case parser.DelayedSubjectProduct:
		index := -1
		for i, producer := range effects {
			if producer.ClauseID != subject.ProducerClauseID {
				continue
			}
			if index >= 0 || producer.subjectScope != content.subjectScope {
				return CompiledReference{}, false
			}
			index = i
		}
		if index < 0 {
			return CompiledReference{}, false
		}
		if selected == nil {
			if !effect.CreatedTokensReference || effects[index].Kind != EffectCreate {
				return CompiledReference{}, false
			}
			reference := issuedIntrinsicSubject(ReferenceBindingPriorInstructionResult,
				ReferenceSubjectPermanent, ReferenceLifetimeActualProduct, content.subjectScope)
			reference.PriorInstruction, reference.ProducerClauseID = index, subject.ProducerClauseID
			reference.Subject.clauseID = reference.ProducerClauseID
			stampSubjectProducer(&reference.Subject, effects[index])
			return reference, true
		}
		if !selected.SubjectProducerMatches(effects[index]) {
			return CompiledReference{}, false
		}
		selected.PriorInstruction = index
		return *selected, true
	case parser.DelayedSubjectTarget:
		if subject.DirectTarget {
			original := content.subjectScope.owner.Targets
			if subject.TargetOccurrence < 0 || subject.TargetOccurrence >= len(original) {
				return CompiledReference{}, false
			}
			target := original[subject.TargetOccurrence]
			index := -1
			for i, candidate := range content.Targets {
				if candidate.Order != target.Order || candidate.subjectScope != content.subjectScope {
					continue
				}
				if index >= 0 {
					return CompiledReference{}, false
				}
				index = i
			}
			return content.TargetSubject(index)
		}
		if selected == nil || selected.Binding != ReferenceBindingTarget {
			return CompiledReference{}, false
		}
	case parser.DelayedSubjectSource:
		if effect.SubjectSourceAttached {
			return content.Source.AttachedObjectSubject(effect)
		}
		if subject.CardZone != zone.None {
			reference, capturedZone, ok := content.Source.CardSubject()
			if ok && capturedZone == subject.CardZone {
				return reference, true
			}
			if selected != nil && selected.Binding == ReferenceBindingSource &&
				selected.CardIdentity && selected.SubjectDomain() == ReferenceSubjectCard {
				return *selected, true
			}
			return CompiledReference{}, false
		}
		if selected == nil {
			return content.Source.OriginalObjectSubject()
		}
		if selected.Binding != ReferenceBindingSource && selected.Binding != ReferenceBindingEventCard {
			return CompiledReference{}, false
		}
	case parser.DelayedSubjectEvent:
		if selected == nil || selected.Binding != ReferenceBindingEventPermanent &&
			selected.Binding != ReferenceBindingEventCard {
			return CompiledReference{}, false
		}
	case parser.DelayedSubjectEventRelated:
		if selected == nil || selected.Binding != ReferenceBindingEventRelatedPermanent {
			return CompiledReference{}, false
		}
	default:
		return CompiledReference{}, false
	}
	return *selected, true
}
