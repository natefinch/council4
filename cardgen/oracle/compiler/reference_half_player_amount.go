package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

func halfPlayerAmountOwnsReference(effect CompiledEffect, reference CompiledReference) bool {
	kind := effect.Kind == EffectLose && effect.LifeObject && effect.Amount.DynamicKind == DynamicAmountHalfPlayerLife ||
		effect.Kind == EffectMill && effect.Amount.DynamicKind == DynamicAmountHalfPlayerLibrary
	if !kind || !effect.Exact || effect.Negated || effect.ClauseID <= 0 ||
		effect.Amount.ReferenceNodeID != reference.NodeID ||
		reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounTheir ||
		reference.Order.End <= reference.Order.Start {
		return false
	}
	for _, owned := range effect.References {
		if owned.NodeID == reference.NodeID && owned.Order == reference.Order {
			return true
		}
	}
	return false
}

func defendingHalfPlayerAmountPolicy(reference CompiledReference, effects []CompiledEffect, source ReferenceSourceContext) bool {
	if source.kind != AbilityTriggered ||
		(source.event != TriggerEventAttackerDeclared && source.event != TriggerEventAttackerBecameBlocked &&
			source.event != TriggerEventAttackerBecameUnblocked) ||
		reference.Binding != ReferenceBindingEventPermanent {
		return false
	}
	for _, effect := range effects {
		if effect.Context == parser.EffectContextDefendingPlayer && len(effect.Targets) == 0 &&
			halfPlayerAmountOwnsReference(effect, reference) {
			return true
		}
	}
	return false
}

// HalfPlayerAmountRecipient permits consuming only this clause's proved
// recipient-owned amount anaphor, not any other player-looking reference.
func (reference CompiledReference) HalfPlayerAmountRecipient(effect CompiledEffect, target *CompiledTarget) bool {
	if !reference.SubjectSupported() || reference.Subject.scope != effect.subjectScope ||
		reference.Subject.scope.owner == nil || !halfPlayerAmountOwnsReference(effect, reference) {
		return false
	}
	matches := 0
	for _, owned := range reference.Subject.scope.owner.Effects {
		if owned.ClauseID == effect.ClauseID && owned.Kind == effect.Kind && owned.Context == effect.Context &&
			owned.Amount.DynamicKind == effect.Amount.DynamicKind && owned.Amount.RoundUp == effect.Amount.RoundUp &&
			halfPlayerAmountOwnsReference(owned, reference) {
			matches++
		}
	}
	if matches != 1 {
		return false
	}
	if target != nil {
		if effect.Context != parser.EffectContextTarget && effect.Context != parser.EffectContextPriorSubject &&
			effect.Context != parser.EffectContextReferencedPlayer && effect.Context != parser.EffectContextEventPlayer {
			return false
		}
		if reference.Subject.domain != ReferenceSubjectPlayer || reference.Binding != ReferenceBindingTarget ||
			!target.Exact || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
			(target.Selector.Kind != SelectorPlayer && target.Selector.Kind != SelectorOpponent) {
			return false
		}
		_, ok := reference.LocalizeSubjectOccurrence(*target, reference.Occurrence)
		return ok
	}
	switch effect.Context {
	case parser.EffectContextEventPlayer, parser.EffectContextReferencedPlayer:
		return reference.Subject.domain == ReferenceSubjectPlayer && reference.Binding == ReferenceBindingEventPlayer
	case parser.EffectContextDefendingPlayer:
		return reference.Subject.domain == ReferenceSubjectPolicy && reference.Binding == ReferenceBindingEventPermanent
	default:
		return false
	}
}
