package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func resultPlayerActorReference(reference compiler.CompiledReference) bool {
	return reference.Kind == compiler.ReferenceThatPlayer ||
		(reference.Kind == compiler.ReferencePronoun && reference.Pronoun == compiler.ReferencePronounThey)
}

func projectedEventControllerPlayer(ctx contentCtx) bool {
	if len(ctx.content.References) != 1 ||
		(ctx.content.Effects[0].Context != parser.EffectContextReferencedPlayer &&
			ctx.content.Effects[0].Context != parser.EffectContextEventPlayer) {
		return false
	}
	reference := ctx.content.References[0]
	return ctx.content.OwnsSubject(reference) && reference.EventControllerResultActor()
}

func bindSequenceOptionalActor(content compiler.AbilityContent, ei int, sequence []game.Instruction, targetIndices map[shared.Span]int) bool {
	effect := content.Effects[ei]
	if !effect.Optional || effect.Context == parser.EffectContextController ||
		content.ControllerObservesTargetLibraryTop(effect) ||
		len(sequence) == 1 && !sequence[0].Optional {
		return true
	}
	if len(sequence) != 1 || sequence[0].OptionalActor.Exists || sequence[0].OptionalActorGroup.Exists {
		return false
	}
	occurrence := -1
	if effect.Context == parser.EffectContextTarget && len(effect.Targets) == 1 {
		for ti, target := range content.Targets {
			if target.Order == effect.Targets[0].Order {
				if occurrence >= 0 {
					return false
				}
				occurrence = ti
			}
		}
	} else if (effect.Context == parser.EffectContextReferencedPlayer || effect.Context == parser.EffectContextEventPlayer) && len(effect.SubjectReferences) == 1 {
		actor, ok := ownedResultActorReference(content, effect.SubjectReferences[0].NodeID)
		if !ok || actor.SubjectDomain() != compiler.ReferenceSubjectPlayer {
			return false
		}
		if actor.Binding == compiler.ReferenceBindingEventPlayer {
			sequence[0].OptionalActor = opt.Val(game.EventPlayerReference())
			return true
		}
		if actor.Binding != compiler.ReferenceBindingTarget {
			return false
		}
		occurrence = actor.Occurrence
	}
	if occurrence < 0 || occurrence >= len(content.Targets) {
		return false
	}
	target := content.Targets[occurrence]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
		(target.Selector.Kind != compiler.SelectorPlayer && target.Selector.Kind != compiler.SelectorOpponent) {
		return false
	}
	index, exists := targetIndices[target.Span]
	if !exists {
		return false
	}
	sequence[0].OptionalActor = opt.Val(game.TargetPlayerReference(index))
	return true
}

func ownedResultActorReference(content compiler.AbilityContent, nodeID int) (compiler.CompiledReference, bool) {
	var actor compiler.CompiledReference
	found := false
	for _, reference := range content.References {
		if reference.NodeID != nodeID {
			continue
		}
		if found || !resultPlayerActorReference(reference) || !content.OwnsSubject(reference) {
			return compiler.CompiledReference{}, false
		}
		actor, found = reference, true
	}
	return actor, found
}

func resultActorCondition(content compiler.AbilityContent, producer int, condition compiler.CompiledCondition, controllerCaused bool) bool {
	nodeID := condition.Ownership.ResultSubjectReferenceNodeID
	controllerObservation := content.ControllerObservesTargetLibraryTop(content.Effects[producer])
	if !slices.ContainsFunc(content.References, func(reference compiler.CompiledReference) bool {
		return slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID) &&
			resultPlayerActorReference(reference)
	}) {
		if nodeID != 0 || len(condition.Ownership.ReferenceNodeIDs) != 0 {
			return false
		}
		return controllerCaused || controllerObservation || resultProducerActsForController(content.Effects[producer])
	}
	if controllerObservation {
		return false
	}
	if len(condition.Ownership.ReferenceNodeIDs) != 1 || condition.Ownership.ReferenceNodeIDs[0] != nodeID {
		return false
	}
	actor, ok := ownedResultActorReference(content, nodeID)
	occurrence := condition.Ownership.ResultSubjectTargetOccurrence
	if !ok || actor.SubjectDomain() != compiler.ReferenceSubjectPlayer {
		return false
	}
	effect := content.Effects[producer]
	if effect.Order.End > condition.Order.Start {
		return false
	}
	for _, later := range content.Effects[producer+1:] {
		if later.Order.Start >= condition.Order.Start {
			break
		}
		if later.Context == parser.EffectContextController {
			continue
		}
		if later.Context == parser.EffectContextTarget && len(later.Targets) == 1 &&
			actor.Binding == compiler.ReferenceBindingTarget && occurrence >= 0 && occurrence < len(content.Targets) &&
			later.Targets[0].Order != content.Targets[occurrence].Order {
			continue
		}
		if len(later.SubjectReferences) == 1 {
			subject, ok := ownedResultActorReference(content, later.SubjectReferences[0].NodeID)
			if ok && actor.Binding == compiler.ReferenceBindingTarget &&
				subject.Binding == compiler.ReferenceBindingTarget && actor.Occurrence != subject.Occurrence {
				continue
			}
		}
		return false
	}
	if actor.Binding == compiler.ReferenceBindingEventPermanent &&
		actor.ObjectProjectionDomain() == compiler.ReferenceSubjectPermanent {
		if occurrence != -1 || effect.Context != parser.EffectContextReferencedObjectController ||
			len(effect.SubjectReferences) != 1 {
			return false
		}
		subject := effect.SubjectReferences[0]
		return content.OwnsSubject(subject) && subject.Binding == compiler.ReferenceBindingEventPermanent &&
			subject.SubjectDomain() == compiler.ReferenceSubjectPermanent &&
			subject.Kind == compiler.ReferencePronoun && subject.Pronoun == compiler.ReferencePronounIts &&
			effect.Order.Contains(subject.Order) && subject.Order.End <= actor.Order.Start
	}
	if actor.Binding == compiler.ReferenceBindingEventPlayer {
		if occurrence != -1 || (effect.Context != parser.EffectContextReferencedPlayer && effect.Context != parser.EffectContextEventPlayer) ||
			len(effect.SubjectReferences) != 1 || !content.OwnsSubject(effect.SubjectReferences[0]) {
			return false
		}
		subject, ok := ownedResultActorReference(content, effect.SubjectReferences[0].NodeID)
		return ok && subject.Binding == compiler.ReferenceBindingEventPlayer &&
			subject.SubjectDomain() == compiler.ReferenceSubjectPlayer && subject.Order.End <= actor.Order.Start &&
			effect.Order.Contains(subject.Order)
	}
	if actor.Binding != compiler.ReferenceBindingTarget || actor.Occurrence != occurrence ||
		occurrence < 0 || occurrence >= len(content.Targets) {
		return false
	}
	target := content.Targets[occurrence]
	_, targetMatches := actor.LocalizeSubjectOccurrence(target, occurrence)
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
		(target.Selector.Kind != compiler.SelectorPlayer && target.Selector.Kind != compiler.SelectorOpponent) ||
		actor.SubjectDomain() != compiler.ReferenceSubjectPlayer || !targetMatches {
		return false
	}
	if effect.Context == parser.EffectContextTarget {
		return len(effect.Targets) == 1 && effect.Targets[0].Order == target.Order
	}
	if (effect.Context != parser.EffectContextReferencedPlayer && effect.Context != parser.EffectContextEventPlayer) || len(effect.SubjectReferences) != 1 ||
		!content.OwnsSubject(effect.SubjectReferences[0]) {
		return false
	}
	subject, ok := ownedResultActorReference(content, effect.SubjectReferences[0].NodeID)
	_, targetMatches = subject.LocalizeSubjectOccurrence(target, occurrence)
	return ok && subject.Binding == actor.Binding && subject.Occurrence == occurrence &&
		subject.SubjectDomain() == compiler.ReferenceSubjectPlayer && targetMatches &&
		subject.Order.End <= actor.Order.Start && effect.Order.Contains(subject.Order)
}
