package compiler

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// bindContextualObjectCondition uses the parser-owned subject identity, not
// other references in the selection or consequence, to resolve the condition.
func bindContextualObjectCondition(
	condition *CompiledCondition,
	references []CompiledReference,
	targets []CompiledTarget,
	effects []CompiledEffect,
	trigger *CompiledTrigger,
) bool {
	for _, subject := range references {
		if subject.NodeID != condition.SubjectRefID {
			continue
		}
		wasSource := subject.Binding == ReferenceBindingSource
		sourceOrder := -1
		if wasSource && subject.Kind == ReferencePronoun {
			for _, reference := range references {
				if (reference.Kind == ReferenceThisObject || reference.Kind == ReferenceSelfName) &&
					reference.Order.Start < subject.Order.Start && reference.Order.Start > sourceOrder {
					sourceOrder = reference.Order.Start
				}
			}
			for _, target := range targets {
				if target.Order.Start > sourceOrder && target.Order.Start < subject.Order.Start {
					occurrence, ok, ambiguous := targetAntecedent(subject, targets)
					if !ok || ambiguous {
						return false
					}
					subject.Binding = ReferenceBindingTarget
					subject.Occurrence = occurrence
					sourceOrder = targets[occurrence].Order.Start
					break
				}
			}
		}
		if !wasSource || subject.Kind == ReferencePronoun {
			for i, owner := range effects {
				if !owner.Order.Contains(subject.Order) || subject.Order.Start >= owner.VerbOrder.Start {
					continue
				}
				// A leading condition belongs to the following instruction.
				// Resolve its antecedent in that instruction's context rather
				// than treating the producer's verb as the consuming verb.
				first := i
				if wasSource {
					first = 0
				}
				for current := first; current <= i; current++ {
					if prior, ok := priorInstructionAntecedentAt(subject, effects, current); ok &&
						(!wasSource || effects[prior].Order.Start > sourceOrder) {
						subject.Binding = ReferenceBindingPriorInstructionResult
						subject.PriorInstruction = prior
					}
				}
				break
			}
		}
		if subject.Binding == ReferenceBindingPriorInstructionResult {
			prior := subject.PriorInstruction
			if prior < 0 || prior >= len(effects) {
				return false
			}
			producer := effects[prior]
			// A targeted non-battlefield card remains readable by card identity
			// after exile. Do not extend this to a new battlefield incarnation
			// (blink), or to cards chosen/revealed during resolution.
			if producer.Kind != EffectExile || len(producer.Targets) != 1 ||
				producer.Targets[0].Selector.Kind != SelectorCard ||
				producer.Targets[0].Selector.Zone == zone.Battlefield ||
				producer.Targets[0].Selector.Zone == zone.None {
				return false
			}
			occurrence, ok, ambiguous := targetAntecedent(subject, targets)
			if !ok || ambiguous || targets[occurrence].Order != producer.Targets[0].Order {
				return false
			}
			subject.Binding = ReferenceBindingTarget
			subject.Occurrence = occurrence
		}
		if subject.Binding == ReferenceBindingTarget && len(condition.SubjectTypes) > 0 {
			if subject.Occurrence < 0 || subject.Occurrence >= len(targets) {
				return false
			}
			if !conditionSubjectMatchesTarget(condition.SubjectTypes, targets[subject.Occurrence]) {
				// A typed "that land" cannot acquire a creature antecedent
				// merely because its target is closer than the land event.
				if trigger == nil || !conditionSubjectTypesContained(condition.SubjectTypes, trigger.Pattern.SubjectSelection.RequiredTypes) {
					return false
				}
				for _, target := range targets {
					if target.Order.Start < subject.Order.Start && conditionSubjectMatchesTarget(condition.SubjectTypes, target) {
						return false
					}
				}
				subject.Binding = ReferenceBindingEventPermanent
			}
		}
		switch subject.Binding {
		case ReferenceBindingTarget:
			if subject.Occurrence < 0 || subject.Occurrence >= len(targets) {
				return false
			}

			target := targets[subject.Occurrence]
			if target.Cardinality.Max != 1 {
				return false
			}
			for _, preceding := range targets[:subject.Occurrence] {
				if preceding.Cardinality.Min != 1 || preceding.Cardinality.Max != 1 {
					return false
				}
			}
			for _, effect := range effects {
				if effect.Order.Start < subject.Order.Start && len(effect.Targets) > 1 {
					for _, owned := range effect.Targets {
						if owned.Order == target.Order {
							return false
						}
					}
				}
			}
			condition.ObjectTarget = &target
		case ReferenceBindingEventPermanent:
			if trigger == nil || trigger.Pattern.OneOrMore ||
				!triggerEventBindsPermanent(trigger.Pattern.Event) {
				return false
			}
		case ReferenceBindingSource:
		default:
			return false
		}
		condition.ObjectBinding = subject.Binding
		condition.ObjectReference = &subject
		return true
	}
	return false
}

func conditionSubjectMatchesTarget(subjectTypes []types.Card, target CompiledTarget) bool {
	targetTypes := target.Selector.RequiredTypesAny()
	implied := map[SelectorKind]types.Card{
		SelectorArtifact: types.Artifact, SelectorCreature: types.Creature,
		SelectorEnchantment: types.Enchantment, SelectorLand: types.Land,
		SelectorPlaneswalker: types.Planeswalker, SelectorBattle: types.Battle,
	}[target.Selector.Kind]
	for _, subjectType := range subjectTypes {
		if subjectType != implied && !slices.Contains(targetTypes, subjectType) {
			return false
		}
	}
	return true
}

func conditionSubjectTypesContained(subjectTypes, candidateTypes []types.Card) bool {
	for _, subjectType := range subjectTypes {
		if !slices.Contains(candidateTypes, subjectType) {
			return false
		}
	}
	return true
}
