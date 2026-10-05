package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func bindDelayedSubjectReference(reference *CompiledReference, effects []CompiledEffect) bool {
	for _, effect := range effects {
		subject := effect.DelayedSubject
		if !slices.Contains(subject.ReferenceNodeIDs, reference.NodeID) {
			continue
		}
		reference.CardIdentity = reference.CardIdentity || subject.CardIdentity || subject.CardZone != zone.None
		switch subject.Kind {
		case parser.DelayedSubjectSource:
			reference.Binding = ReferenceBindingSource
		case parser.DelayedSubjectTarget:
			reference.Binding = ReferenceBindingTarget
			reference.Occurrence = subject.TargetOccurrence
		case parser.DelayedSubjectProduct:
			reference.Binding = ReferenceBindingUnsupported
			for i, producer := range effects {
				if producer.ClauseID == subject.ProducerClauseID {
					reference.Binding = ReferenceBindingPriorInstructionResult
					reference.PriorInstruction = i
					break
				}
			}
		case parser.DelayedSubjectEvent:
			reference.Binding = ReferenceBindingEventPermanent
		case parser.DelayedSubjectEventRelated:
			reference.Binding = ReferenceBindingEventRelatedPermanent
		default:
			reference.Binding = ReferenceBindingUnsupported
		}
		return true
	}
	return false
}
