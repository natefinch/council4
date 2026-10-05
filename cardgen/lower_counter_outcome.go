package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const counterOwnershipCategory = "structural — counter outcome has no unique owned target occurrence"
const counterDestinationCategory = "structural — counter destination replacement not modeled"

func counterHasActualSuccessCondition(content compiler.AbilityContent, ei int) bool {
	if ei < 0 || ei >= len(content.Effects) {
		return false
	}
	found := false
	for _, condition := range content.Conditions {
		if condition.Ownership.ResultProducerClauseID != content.Effects[ei].ClauseID {
			continue
		}
		if condition.Predicate != compiler.ConditionPredicateCounterSucceeded {
			return false
		}
		found = true
	}
	return found
}

type counterDestinationPlan struct {
	modifiers  map[int]game.CounterObject
	absorbed   map[int]bool
	conditions []compiler.CompiledCondition
}

func counterOutcomeProducer(content compiler.AbilityContent, condition compiler.CompiledCondition) (int, bool) {
	producer := -1
	for ei, effect := range content.Effects {
		if effect.ClauseID == condition.Ownership.ResultProducerClauseID && effect.ClauseID > 0 {
			if producer >= 0 {
				return -1, false
			}
			producer = ei
		}
	}
	if producer < 0 {
		return -1, false
	}
	if condition.Ownership.ResultSubjectReferenceNodeID < 0 ||
		!slices.Contains(condition.Ownership.ReferenceNodeIDs, condition.Ownership.ResultSubjectReferenceNodeID) {
		return -1, false
	}
	effect := content.Effects[producer]
	if effect.Kind != compiler.EffectCounter || !effect.Exact || effect.Negated ||
		effect.DelayedTiming != 0 || len(effect.Targets) != 1 {
		return -1, false
	}
	target := effect.Targets[0]
	if target.Selector.Kind != compiler.SelectorSpell ||
		target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return -1, false
	}
	subjects := 0
	for _, reference := range content.References {
		if reference.NodeID != condition.Ownership.ResultSubjectReferenceNodeID {
			continue
		}
		subjects++
		if reference.Kind != compiler.ReferenceThatObject || reference.Binding != compiler.ReferenceBindingTarget ||
			reference.Occurrence < 0 || reference.Occurrence >= len(content.Targets) ||
			reference.Occurrence != condition.Ownership.ResultSubjectTargetOccurrence ||
			content.Targets[reference.Occurrence].Order != target.Order {
			return -1, false
		}
		for ti, other := range content.Targets {
			if ti != reference.Occurrence && other.Order == target.Order {
				return -1, false
			}
		}
	}
	return producer, subjects == 1
}

func planCounterDestinations(content compiler.AbilityContent) (counterDestinationPlan, compiler.AbilityContent, string) {
	plan := counterDestinationPlan{
		modifiers: make(map[int]game.CounterObject), absorbed: make(map[int]bool),
	}
	planning := content
	planning.Conditions = slices.Clone(content.Conditions)
	for ei, effect := range content.Effects {
		if !effect.CounteredSpellExileReplacement && !effect.CounteredSpellDestinationReplacement {
			continue
		}
		if !effect.Exact || effect.Optional || effect.Negated || effect.DelayedTiming != 0 ||
			len(effect.Targets) != 0 || effect.ClauseID <= 0 {
			return plan, planning, counterDestinationCategory
		}
		matches, producer := 0, -1
		for _, condition := range content.Conditions {
			if !slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) {
				continue
			}
			matches++
			if condition.Predicate != compiler.ConditionPredicateCounterSucceeded ||
				condition.Kind != compiler.ConditionIf || condition.Negated || condition.Intervening ||
				condition.Ownership.Scope != parser.ConditionScopeClause ||
				len(condition.Ownership.ClauseIDs) != 1 {
				return plan, planning, counterOwnershipCategory
			}
			var ok bool
			producer, ok = counterOutcomeProducer(content, condition)
			if !ok || producer >= ei {
				return plan, planning, counterOwnershipCategory
			}
			plan.conditions = append(plan.conditions, condition)
		}
		if matches != 1 {
			return plan, planning, counterOwnershipCategory
		}
		if _, exists := plan.modifiers[producer]; exists {
			return plan, planning, counterDestinationCategory
		}
		modifier := game.CounterObject{}
		switch {
		case effect.Kind == compiler.EffectExile && effect.CounteredSpellExileReplacement &&
			!effect.CounteredSpellDestinationReplacement:
			modifier.ExileInstead = true
		case effect.Kind == compiler.EffectPut && effect.CounteredSpellDestinationReplacement &&
			!effect.CounteredSpellExileReplacement:
			destination, ok := counteredSpellRedirectDestination(&effect)
			if !ok {
				return plan, planning, counterDestinationCategory
			}
			modifier.Destination = destination
		default:
			return plan, planning, counterDestinationCategory
		}
		plan.modifiers[producer], plan.absorbed[ei] = modifier, true
	}
	planning.Conditions = slices.DeleteFunc(planning.Conditions, func(condition compiler.CompiledCondition) bool {
		return slices.ContainsFunc(plan.conditions, func(owned compiler.CompiledCondition) bool {
			return owned.NodeID == condition.NodeID
		})
	})
	return plan, planning, ""
}

// A tax clause's Pay result controls whether its terminal counter runs. Only
// that counter can publish success or own an intrinsic destination modifier.
func counterActionInstruction(sequence []game.Instruction) (int, bool) {
	if len(sequence) == 0 || len(sequence) > 2 {
		return -1, false
	}
	index := len(sequence) - 1
	if _, ok := sequence[index].Primitive.(game.CounterObject); !ok {
		return -1, false
	}
	if index == 0 {
		return index, !sequence[index].ResultGate.Exists
	}
	payment, counter := sequence[0], sequence[1]
	if _, ok := payment.Primitive.(game.Pay); !ok || payment.PublishResult == "" ||
		!counter.ResultGate.Exists || counter.ResultGate.Val.Key != payment.PublishResult ||
		counter.ResultGate.Val.Succeeded != game.TriFalse ||
		counter.ResultGate.Val.Negate || counter.ResultGate.Val.ObjectSelection.Exists ||
		counter.ResultGate.Val.ObjectCountRange.Exists {
		return -1, false
	}
	return index, true
}

func (plan counterDestinationPlan) apply(ei int, sequence []game.Instruction) bool {
	modifier, exists := plan.modifiers[ei]
	if !exists {
		return true
	}
	index, ok := counterActionInstruction(sequence)
	if !ok {
		return false
	}
	counter := sequence[index].Primitive.(game.CounterObject)
	if counter.ExileInstead || counter.Destination != game.CounteredSpellGraveyard {
		return false
	}
	counter.ExileInstead, counter.Destination = modifier.ExileInstead, modifier.Destination
	sequence[index].Primitive = counter
	return true
}
