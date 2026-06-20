package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

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
			precedingSourceReferenceAfter(bound[:i], reference.Order, trigger.Order.End) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if prior, ok := priorInstructionAntecedent(*reference, effects); ok {
			if precedingSourceReferenceAfter(
				bound[:i],
				reference.Order,
				effects[prior].VerbOrder.Start,
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
			reference.Order.Start >= trigger.Order.Start &&
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
		if trigger == nil && precedingSourceReference(bound[:i], reference.Order) {
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
			reference.Order.Start >= trigger.Order.Start &&
			!trigger.Pattern.OneOrMore &&
			triggerEventBindsPermanent(trigger.Pattern.Event) {
			reference.Binding = ReferenceBindingEventPermanent
			continue
		}
		if trigger != nil &&
			reference.Kind == ReferencePronoun &&
			reference.Order.Start >= trigger.Order.Start &&
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
			cost.Order.Contains(reference.Order) {
			if activationCostPronounBindsSource(cost, reference.Order) {
				reference.Binding = ReferenceBindingSource
			} else {
				reference.Binding = ReferenceBindingAmbiguous
			}
		}
	}
	return bound
}

func activationCostPronounBindsSource(cost *CompiledCost, reference shared.SourceOrder) bool {
	componentIndex := -1
	for i, component := range cost.Components {
		if component.Order.Contains(reference) {
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
		reference.Order.Start < effects[0].VerbOrder.End {
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
		if effect.VerbOrder.Start >= reference.Order.Start {
			continue
		}
		if current < 0 || effect.VerbOrder.Start > effects[current].VerbOrder.Start {
			current = i
		}
	}
	if current < 0 {
		return 0, false
	}
	prior := -1
	for i := range effects {
		effect := &effects[i]
		if effect.VerbOrder.Start >= effects[current].VerbOrder.Start {
			continue
		}
		if prior < 0 || effect.VerbOrder.Start > effects[prior].VerbOrder.Start {
			prior = i
		}
	}
	if prior < 0 {
		return 0, false
	}
	if search, ok := priorSearchMoveAntecedent(current, effects); ok {
		return search, true
	}
	switch effects[prior].Kind {
	case EffectDig, EffectExile, EffectManifestDread, EffectReveal, EffectSearch:
		return prior, true
	case EffectPut, EffectReturn:
		effect := effects[prior]
		if reference.Kind == ReferenceThatObject &&
			effect.FromZone == zone.Graveyard &&
			effect.ToZone == zone.Battlefield &&
			effect.UnderYourControl {
			return prior, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func priorSearchMoveAntecedent(current int, effects []CompiledEffect) (int, bool) {
	if effects[current].Kind == EffectPut && current >= 2 &&
		effects[current-1].Kind == EffectShuffle {
		search := current - 2
		if effects[search].Kind == EffectReveal {
			search--
		}
		if search >= 0 && effects[search].Kind == EffectSearch &&
			effects[search].SearchDestination == parser.EffectDestinationTop {
			return search, true
		}
	}
	if current < 3 ||
		effects[current-1].Kind != EffectShuffle ||
		effects[current-2].Kind != EffectPut ||
		effects[current-3].Kind != EffectSearch {
		return 0, false
	}
	return current - 3, true
}

func targetAntecedent(reference CompiledReference, targets []CompiledTarget) (occurrence int, ok, ambiguous bool) {
	closest := -1
	for i, target := range targets {
		if target.Order.Start >= reference.Order.Start {
			continue
		}
		if closest < 0 || target.Order.Start > targets[closest].Order.Start {
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
		if targets[i].Order.Start < reference.Order.Start &&
			targets[i].Order.Start == target.Order.Start {
			return 0, false, true
		}
	}
	return closest, true, false
}

func precedingSourceReference(references []CompiledReference, order shared.SourceOrder) bool {
	return precedingSourceReferenceAfter(references, order, 0)
}

func precedingSourceReferenceAfter(references []CompiledReference, order shared.SourceOrder, after int) bool {
	for _, reference := range references {
		if reference.Order.Start >= after &&
			reference.Order.Start < order.Start &&
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
		if !effect.Order.Contains(reference.Order) {
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
	case TriggerEventSpellCast,
		TriggerEventCardDrawn,
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
