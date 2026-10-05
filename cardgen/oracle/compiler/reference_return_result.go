package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func bindReturnedConditionConsequences(content *AbilityContent) {
	bindReturnedConditionReferences(content.Conditions, content.References, content.Targets, content.Effects)
}

func bindReturnedConditionReferences(conditions []CompiledCondition, references []CompiledReference, targets []CompiledTarget, effects []CompiledEffect) {
	for _, condition := range conditions {
		if condition.Predicate != ConditionPredicateObjectMatches ||
			condition.ObjectBinding != ReferenceBindingPriorInstructionResult || condition.ObjectReference == nil {
			continue
		}
		prior := condition.ObjectReference.PriorInstruction
		if !singularReturnedPermanentProducer(effects, prior) {
			continue
		}
		producer := effects[prior]
		for i := range references {
			reference := &references[i]
			if slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID) ||
				(reference.Kind != ReferenceThatObject &&
					(reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIt && reference.Pronoun != ReferencePronounIts)) {
				continue
			}
			owned := slices.ContainsFunc(effects, func(effect CompiledEffect) bool {
				return slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) && effect.Order.Contains(reference.Order)
			})
			if !owned {
				continue
			}
			sameSubject := reference.Binding == ReferenceBindingPriorInstructionResult && reference.PriorInstruction == prior
			if len(producer.Targets) == 1 && reference.Binding == ReferenceBindingTarget &&
				reference.Occurrence >= 0 && reference.Occurrence < len(targets) {
				sameSubject = producer.Targets[0].Order == targets[reference.Occurrence].Order
			}
			if len(producer.Targets) == 0 {
				sameSubject = reference.Binding == ReferenceBindingPriorInstructionResult && reference.PriorInstruction == prior-1 ||
					reference.Kind == ReferencePronoun && reference.Binding == ReferenceBindingSource &&
						exileUsesSource(effects[prior-1])
				if reference.Binding == ReferenceBindingTarget && len(effects[prior-1].Targets) == 1 &&
					reference.Occurrence >= 0 && reference.Occurrence < len(targets) {
					sameSubject = effects[prior-1].Targets[0].Order == targets[reference.Occurrence].Order
				}
			}
			if sameSubject {
				reference.Binding = ReferenceBindingPriorInstructionResult
				reference.PriorInstruction = prior
			}
		}
	}
}

// A later object subject names the latest returned incarnation, even across
// clauses that introduce no object. Opaque producers must supersede it too.
func priorReturnedPermanentAntecedent(reference CompiledReference, effects []CompiledEffect) (int, bool) {
	if reference.Kind != ReferenceThatObject &&
		(reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIt &&
			reference.Pronoun != ReferencePronounIts) {
		return 0, false
	}
	prior := -1
	for i, effect := range effects {
		if effect.Order.End > reference.Order.Start {
			continue
		}
		switch effect.Kind {
		case EffectReturn, EffectPut:
			if effect.ToZone != zone.Battlefield {
				continue
			}
		case EffectExile, EffectReveal, EffectDig, EffectSearch, EffectChoosePermanent,
			EffectCreate, EffectManifestDread, EffectSacrifice, EffectDestroy:
		default:
			if len(effect.Targets) == 0 || effect.Targets[0].Selector.Kind == SelectorPlayer {
				continue
			}
		}
		if prior < 0 || effect.VerbOrder.Start > effects[prior].VerbOrder.Start {
			prior = i
		}
	}
	if prior < 0 || (effects[prior].Kind != EffectReturn && effects[prior].Kind != EffectPut) ||
		effects[prior].ToZone != zone.Battlefield {
		return 0, false
	}
	return prior, true
}

func singularReturnedPermanentProducer(effects []CompiledEffect, index int) bool {
	if index < 0 || index >= len(effects) {
		return false
	}
	effect := effects[index]
	if (effect.Kind != EffectReturn && effect.Kind != EffectPut) ||
		effect.ToZone != zone.Battlefield || effect.Optional ||
		effect.Negated || effect.DelayedTiming != 0 {
		return false
	}
	if len(effect.Targets) > 0 {
		return effect.Exact && len(effect.Targets) == 1 && effect.Targets[0].Selector.Zone != zone.None &&
			effect.Targets[0].Selector.Zone != zone.Battlefield &&
			effect.Targets[0].Cardinality.Min == 1 && effect.Targets[0].Cardinality.Max == 1
	}
	// The modeled targetless return is the singular immediate blink of a
	// target or the source. Choice/group returns are not singular by inference.
	if effect.Kind != EffectReturn || index == 0 || effects[index-1].Kind != EffectExile ||
		effects[index-1].Optional || effects[index-1].Negated || effects[index-1].DelayedTiming != 0 {
		return false
	}
	exile := effects[index-1]
	if len(exile.Targets) == 1 {
		return exile.Targets[0].Cardinality.Min == 1 && exile.Targets[0].Cardinality.Max == 1
	}
	return len(exile.Targets) == 0 && exileUsesSource(exile)
}

func exileUsesSource(exile CompiledEffect) bool {
	return exile.Context == parser.EffectContextSource ||
		slices.ContainsFunc(exile.References, func(reference CompiledReference) bool {
			return reference.Kind == ReferenceThisObject || reference.Kind == ReferenceSelfName
		})
}
