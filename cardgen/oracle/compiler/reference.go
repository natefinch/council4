package compiler

import "github.com/natefinch/council4/cardgen/oracle/shared"

// bindReferences assigns each recognized reference phrase one conservative
// referent. It never guesses between multiple target occurrences or an
// unsupported antecedent.
func bindReferences(
	references []CompiledReference,
	targets []CompiledTarget,
	effects []CompiledEffect,
	trigger *CompiledTrigger,
) []CompiledReference {
	bound := append([]CompiledReference(nil), references...)
	for i := range bound {
		reference := &bound[i]
		switch reference.Kind {
		case ReferenceSelfName, ReferenceThisObject:
			reference.Binding = ReferenceBindingSource
			continue
		case ReferencePronoun, ReferenceThatObject, ReferenceThatPlayer:
		default:
			reference.Binding = ReferenceBindingUnsupported
			continue
		}

		if trigger != nil &&
			precedingSourceReferenceAfter(bound[:i], reference.Span, trigger.Span.End.Offset) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if prior, ok := priorInstructionAntecedent(*reference, effects); ok {
			if precedingSourceReferenceAfter(
				bound[:i],
				reference.Span,
				effects[prior].VerbSpan.Start.Offset,
			) {
				reference.Binding = ReferenceBindingSource
			} else {
				reference.Binding = ReferenceBindingPriorInstructionResult
				reference.PriorInstruction = prior
			}
			continue
		}
		if trigger != nil &&
			reference.Kind == ReferenceThatPlayer &&
			reference.Span.Start.Offset >= trigger.Span.Start.Offset &&
			triggerEventBindsPlayer(trigger.Pattern.Event) {
			reference.Binding = ReferenceBindingEventPlayer
			continue
		}
		if occurrence, ok, ambiguous := targetAntecedent(*reference, targets); ok || ambiguous {
			if ambiguous {
				reference.Binding = ReferenceBindingAmbiguous
			} else {
				reference.Binding = ReferenceBindingTarget
				reference.Occurrence = occurrence
			}
			continue
		}
		if trigger == nil && precedingSourceReference(bound[:i], reference.Span) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if delayedEffectBindsSource(*reference, effects) &&
			(trigger == nil || trigger.Pattern.Source == TriggerSourceSelf) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if trigger != nil && triggerReferenceBindsEventCard(&trigger.Pattern, *reference, effects) {
			reference.Binding = ReferenceBindingEventCard
			continue
		}
		if trigger != nil &&
			reference.Span.Start.Offset >= trigger.Span.Start.Offset &&
			!trigger.Pattern.OneOrMore &&
			triggerEventBindsPermanent(trigger.Pattern.Event) {
			reference.Binding = ReferenceBindingEventPermanent
			continue
		}
		if trigger != nil &&
			reference.Kind == ReferencePronoun &&
			reference.Span.Start.Offset >= trigger.Span.Start.Offset &&
			triggerEventBindsPlayer(trigger.Pattern.Event) &&
			(reference.Pronoun == ReferencePronounThey ||
				reference.Pronoun == ReferencePronounTheir ||
				reference.Pronoun == ReferencePronounThem) {
			reference.Binding = ReferenceBindingEventPlayer
			continue
		}
		if reference.Kind == ReferencePronoun {
			reference.Binding = ReferenceBindingAmbiguous
		} else {
			reference.Binding = ReferenceBindingUnsupported
		}

	}
	return bound
}

func bindActivationCostReferences(kind AbilityKind, cost *CompiledCost, references []CompiledReference) []CompiledReference {
	if kind != AbilityActivated || cost == nil {
		return references
	}
	bound := append([]CompiledReference(nil), references...)
	for i := range bound {
		reference := &bound[i]
		if reference.Kind == ReferencePronoun &&
			reference.Pronoun == ReferencePronounIt &&
			spanContains(cost.Span, reference.Span) {
			if activationCostPronounBindsSource(cost, reference.Span) {
				reference.Binding = ReferenceBindingSource
			} else {
				reference.Binding = ReferenceBindingAmbiguous
			}
		}
	}
	return bound
}

func activationCostPronounBindsSource(cost *CompiledCost, reference shared.Span) bool {
	componentIndex := -1
	for i, component := range cost.Components {
		if spanContains(component.Span, reference) {
			componentIndex = i
			break
		}
	}
	if componentIndex < 0 {
		return false
	}
	switch cost.Components[componentIndex].Kind {
	case CostRemoveCounter, CostPutCounter, CostExert:
	default:
		return false
	}
	for _, component := range cost.Components[:componentIndex] {
		switch component.Kind {
		case CostMana, CostTap, CostUntap, CostPayLife, CostEnergy, CostLoyalty:
		default:
			return false
		}
	}
	return true
}

func delayedEffectBindsSource(reference CompiledReference, effects []CompiledEffect) bool {
	if len(effects) != 1 ||
		effects[0].DelayedTiming == 0 ||
		reference.Span.Start.Offset < effects[0].VerbSpan.End.Offset {
		return false
	}
	switch effects[0].Kind {
	case EffectExile, EffectReturn, EffectSacrifice:
		return true
	default:
		return false
	}
}

func priorInstructionAntecedent(reference CompiledReference, effects []CompiledEffect) (int, bool) {
	current := -1
	for i := range effects {
		effect := &effects[i]
		if effect.VerbSpan.Start.Offset >= reference.Span.Start.Offset {
			continue
		}
		if current < 0 || effect.VerbSpan.Start.Offset > effects[current].VerbSpan.Start.Offset {
			current = i
		}
	}
	if current < 0 {
		return 0, false
	}
	prior := -1
	for i := range effects {
		effect := &effects[i]
		if effect.VerbSpan.Start.Offset >= effects[current].VerbSpan.Start.Offset {
			continue
		}
		if prior < 0 || effect.VerbSpan.Start.Offset > effects[prior].VerbSpan.Start.Offset {
			prior = i
		}
	}
	if prior < 0 {
		return 0, false
	}
	switch effects[prior].Kind {
	case EffectExile, EffectManifestDread, EffectReveal, EffectSearch:
		return prior, true
	default:
		return 0, false
	}
}

func targetAntecedent(reference CompiledReference, targets []CompiledTarget) (occurrence int, ok, ambiguous bool) {
	closest := -1
	for i, target := range targets {
		if target.Span.Start.Offset >= reference.Span.Start.Offset {
			continue
		}
		if closest < 0 || target.Span.Start.Offset > targets[closest].Span.Start.Offset {
			closest = i
		}
	}
	if closest < 0 {
		return 0, false, false
	}
	target := targets[closest]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return 0, false, true
	}
	for i := closest + 1; i < len(targets); i++ {
		if targets[i].Span.Start.Offset < reference.Span.Start.Offset &&
			targets[i].Span.Start.Offset == target.Span.Start.Offset {
			return 0, false, true
		}
	}
	return closest, true, false
}

func precedingSourceReference(references []CompiledReference, span shared.Span) bool {
	return precedingSourceReferenceAfter(references, span, 0)
}

func precedingSourceReferenceAfter(references []CompiledReference, span shared.Span, after int) bool {
	for _, reference := range references {
		if reference.Span.Start.Offset >= after &&
			reference.Span.Start.Offset < span.Start.Offset &&
			reference.Binding == ReferenceBindingSource {
			return true
		}
	}
	return false
}

func triggerEventBindsPermanent(event TriggerEvent) bool {
	switch event {
	case TriggerEventPermanentEnteredBattlefield,
		TriggerEventPermanentDied,
		TriggerEventZoneChanged,
		TriggerEventCountersAdded,
		TriggerEventDamageDealt,
		TriggerEventAttackerDeclared,
		TriggerEventBlockerDeclared,
		TriggerEventPermanentTapped,
		TriggerEventPermanentUntapped,
		TriggerEventPermanentTurnedFaceUp,
		TriggerEventPermanentSacrificed,
		TriggerEventObjectBecameTarget,
		TriggerEventPermanentMutated,
		TriggerEventAttackerBecameBlocked:
		return true
	default:
		return false
	}
}

func triggerReferenceBindsEventCard(
	trigger *TriggerPattern,
	reference CompiledReference,
	effects []CompiledEffect,
) bool {
	if trigger.OneOrMore ||
		(trigger.Event != TriggerEventPermanentDied &&
			(trigger.Event != TriggerEventZoneChanged ||
				!trigger.MatchToZone ||
				trigger.ToZone != TriggerZoneGraveyard)) {
		return false
	}
	for i := range effects {
		effect := &effects[i]
		if !spanContains(effect.Span, reference.Span) {
			continue
		}
		switch effect.Kind {
		case EffectReturn, EffectExile, EffectCast:
			return true
		default:
			return false
		}
	}
	return false
}

// triggerEventBindsPlayer reports whether the trigger event has an authoritative
// player subject. When true, player pronouns in the trigger body (they/their/them)
// are conservatively bound to ReferenceBindingEventPlayer.
func triggerEventBindsPlayer(event TriggerEvent) bool {
	switch event {
	case TriggerEventCardDrawn,
		TriggerEventCardDiscarded,
		TriggerEventCycled,
		TriggerEventScry,
		TriggerEventSurveil,
		TriggerEventLifeGained,
		TriggerEventLifeLost:
		return true
	default:
		return false
	}
}
