package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func recurringSourceCardFunctionZone(ability CompiledAbility) zone.Type {
	if ability.Kind != AbilityTriggered || ability.Trigger == nil ||
		ability.Trigger.Pattern.Kind != TriggerAt ||
		ability.Trigger.Pattern.Event != TriggerEventBeginningOfStep ||
		len(ability.Content.Modes) != 0 {
		return zone.Battlefield
	}
	switch ability.Trigger.Pattern.Step {
	case TriggerStepUpkeep, TriggerStepBeginningOfCombat, TriggerStepEnd:
	default:
		return zone.Battlefield
	}
	for i := range ability.Content.Effects {
		if !exactRecurringSourceCardEffect(ability.Content.Effects, i) {
			continue
		}
		for _, preceding := range ability.Content.Effects[:i] {
			for _, prior := range preceding.References {
				if prior.Binding == ReferenceBindingSource {
					return zone.Battlefield
				}
			}
		}
		return zone.Graveyard
	}
	return zone.Battlefield
}

func exactRecurringSourceCardEffect(effects []CompiledEffect, i int) bool {
	effect := effects[i]
	if effect.Kind != EffectReturn && effect.Kind != EffectPut ||
		!effect.Exact || effect.Negated || effect.FromZone != zone.Graveyard ||
		effect.ToZone != zone.Hand && effect.ToZone != zone.Battlefield ||
		len(effect.Targets) != 0 || len(effect.References) != 1 {
		return false
	}
	reference := effect.References[0]
	return reference.Binding == ReferenceBindingSource &&
		(reference.Kind == ReferenceSelfName ||
			reference.Kind == ReferenceThisObject && reference.SubjectNoun == parser.ObjectNounCard) &&
		referenceFollowsEffectVerbInClause(i, effects, reference.Order)
}

func recurringSourceCardReference(source ReferenceSourceContext, reference CompiledReference, effects []CompiledEffect) bool {
	if source.kind != AbilityTriggered || source.event != TriggerEventBeginningOfStep || source.zone != zone.Graveyard {
		return false
	}
	for i := range effects {
		if exactRecurringSourceCardEffect(effects, i) && effectOwnsReference(effects[i], reference) {
			return true
		}
	}
	return false
}

// TriggerFunctionZone exposes only an owned, finalized recurring source shell.
func (content AbilityContent) TriggerFunctionZone() (zone.Type, bool) {
	source := content.Source
	if source.scope == nil || source.scope.owner == nil || !source.known ||
		content.subjectScope != source.scope ||
		source.kind != AbilityTriggered || source.event != TriggerEventBeginningOfStep ||
		source.scope.owner.Source.scope != source.scope ||
		source.zone != zone.Battlefield && source.zone != zone.Graveyard {
		return zone.None, false
	}
	for _, reference := range content.References {
		if reference.Binding == ReferenceBindingSource && !content.OwnsSubject(reference) {
			return zone.None, false
		}
	}
	if source.zone == zone.Battlefield {
		return zone.None, true
	}
	return source.zone, true
}
