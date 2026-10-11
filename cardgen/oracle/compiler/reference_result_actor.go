package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func resultPlayerActorReference(reference CompiledReference) bool {
	return reference.Kind == ReferenceThatPlayer ||
		(reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounThey)
}

func resultReferencedActorProducer(content AbilityContent, reference CompiledReference) (CompiledEffect, bool) {
	var producer CompiledEffect
	found := false
	for _, condition := range content.Conditions {
		if (condition.Predicate != ConditionPredicatePriorInstructionAccepted &&
			condition.Predicate != ConditionPredicatePriorInstructionNotAccepted) ||
			condition.Ownership.ResultProducerClauseID <= 0 ||
			condition.Ownership.ResultSubjectTargetOccurrence != -1 ||
			len(condition.Ownership.ReferenceNodeIDs) != 1 ||
			condition.Ownership.ReferenceNodeIDs[0] != condition.Ownership.ResultSubjectReferenceNodeID {
			continue
		}
		actors := 0
		for _, actor := range content.References {
			if actor.NodeID == condition.Ownership.ResultSubjectReferenceNodeID && resultPlayerActorReference(actor) {
				actors++
			}
		}
		if actors != 1 {
			return CompiledEffect{}, false
		}
		owns := condition.Ownership.ResultSubjectReferenceNodeID == reference.NodeID &&
			slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID)
		for _, effect := range content.Effects {
			if slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) &&
				(effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer) &&
				slices.ContainsFunc(effect.SubjectReferences, func(subject CompiledReference) bool { return subject.NodeID == reference.NodeID }) {
				owns = true
			}
		}
		if !owns || !resultPlayerActorReference(reference) {
			continue
		}
		matches := 0
		for _, effect := range content.Effects {
			if effect.ClauseID != condition.Ownership.ResultProducerClauseID {
				continue
			}
			matches++
			producer = effect
		}
		controller := producer.Context == parser.EffectContextReferencedObjectController
		player := producer.Context == parser.EffectContextReferencedPlayer || producer.Context == parser.EffectContextEventPlayer
		if found || matches != 1 || (!controller && !player) ||
			producer.Order.End > condition.Order.Start || len(producer.SubjectReferences) != 1 {
			return CompiledEffect{}, false
		}
		sourceNode := producer.SubjectReferences[0].NodeID
		sources := 0
		for _, source := range content.References {
			valid := controller && source.Binding == ReferenceBindingEventPermanent &&
				source.Kind == ReferencePronoun && source.Pronoun == ReferencePronounIts ||
				player && source.Binding == ReferenceBindingEventPlayer && resultPlayerActorReference(source)
			if source.NodeID == sourceNode && valid &&
				producer.Order.Contains(source.Order) && source.Order.End <= reference.Order.Start {
				sources++
			}
		}
		if sources != 1 {
			return CompiledEffect{}, false
		}
		found = true
	}
	return producer, found
}

// EventControllerResultActor checks the owned singular action-result proof for
// an event permanent's controller, retaining the permanent's object projection
// rather than borrowing a zone-event player.
func (reference CompiledReference) EventControllerResultActor() bool {
	if !reference.SubjectSupported() || reference.Binding != ReferenceBindingEventPermanent ||
		reference.SubjectDomain() != ReferenceSubjectPlayer ||
		reference.ObjectProjectionDomain() != ReferenceSubjectPermanent ||
		reference.Subject.scope.owner == nil {
		return false
	}
	producer, ok := resultReferencedActorProducer(*reference.Subject.scope.owner, reference)
	return ok && producer.Context == parser.EffectContextReferencedObjectController
}

func bindResultActorReferences(content *AbilityContent, trigger *CompiledTrigger) {
	if trigger == nil || trigger.Pattern.OneOrMore {
		return
	}
	for ri := range content.References {
		reference := &content.References[ri]
		if producer, ok := resultReferencedActorProducer(*content, *reference); ok {
			if producer.Context == parser.EffectContextReferencedObjectController {
				if !triggerEventBindsPermanent(trigger.Pattern.Event) {
					continue
				}
				reference.Binding = ReferenceBindingEventPermanent
			} else {
				if !triggerPatternBindsThatPlayer(&trigger.Pattern) {
					continue
				}
				reference.Binding = ReferenceBindingEventPlayer
			}
			reference.SubjectNoun = parser.ObjectNounPlayer
		}
	}
}
