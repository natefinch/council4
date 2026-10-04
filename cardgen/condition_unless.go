package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// resolvingStateUnless admits only controller-state predicates whose existing
// Selection/aggregate adapters implement the complete negated resolving gate.
// Payment outcomes, result antecedents and object bindings are not state gates.
func resolvingStateUnless(condition compiler.CompiledCondition) bool {
	if condition.Kind != compiler.ConditionUnless || !condition.Negated ||
		condition.Intervening || condition.SourceInGraveyard || condition.Selection.ExcludeSource ||
		condition.Threshold < 0 {
		return false
	}

	switch condition.Predicate {
	case compiler.ConditionPredicateControllerControls:
		return true
	case compiler.ConditionPredicateControllerGraveyardCardCountAtLeast,
		compiler.ConditionPredicateControllerGraveyardManaValueCountAtLeast:
		return conditionSelectionEmpty(condition.Selection)
	default:
		return false
	}
}

func effectGateLoweringContext(condition compiler.CompiledCondition) conditionLoweringContext {
	if condition.Kind == compiler.ConditionUnless {
		return conditionContextUnlessEffectGate
	}
	return conditionContextEffectGate
}

func matchSequenceEffectConditions(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) (map[int]game.EffectCondition, string, bool) {
	return matchEffectConditions(effects, conditions, false)
}

func matchOrderedSequenceEffectConditions(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) (map[int]game.EffectCondition, string, bool) {
	return matchEffectConditions(effects, conditions, true)
}

func effectHasUnlessGate(effect compiler.CompiledEffect, conditions []compiler.CompiledCondition) bool {
	for _, condition := range conditions {
		if condition.Kind == compiler.ConditionUnless && spanCovered(condition.Span, []shared.Span{effect.Span}) {
			return true
		}
	}
	return false
}
