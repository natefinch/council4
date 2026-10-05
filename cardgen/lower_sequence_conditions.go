package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const (
	sequenceTaxOwnerCategory      = "structural — counter payment condition has no unique typed owner"
	sequenceMultipleTaxesCategory = "structural — multiple embedded counter payments not modeled"
)

type sequenceConditionPlan struct {
	gates              map[int]game.EffectCondition
	clauseConditions   map[int][]compiler.CompiledCondition
	gateConditions     []compiler.CompiledCondition
	externalConditions []compiler.CompiledCondition
}

func planSequenceConditions(
	content compiler.AbilityContent,
	optionalFlow optionalFlowPlan,
) (sequenceConditionPlan, string, bool) {
	plan := sequenceConditionPlan{
		clauseConditions: make(map[int][]compiler.CompiledCondition),
	}
	owners := make(map[int]int)
	for ei := range content.Effects {
		effect := &content.Effects[ei]
		if !effectOwnsCounterTax(*effect) {
			continue
		}
		nodeID := effect.Payment.FailureConditionNodeID
		if nodeID < 0 {
			return sequenceConditionPlan{}, sequenceTaxOwnerCategory, false
		}
		if _, exists := owners[nodeID]; exists {
			return sequenceConditionPlan{}, sequenceTaxOwnerCategory, false
		}
		owners[nodeID] = ei
	}
	// The existing clause lowerer uses one fixed result key. Until payment
	// results have scoped identities, composing two would alias their outcomes.
	if len(owners) > 1 {
		return sequenceConditionPlan{}, sequenceMultipleTaxesCategory, false
	}
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		ei, owned := owners[condition.NodeID]
		if !owned {
			if condition.Predicate == compiler.ConditionPredicateTargetControllerDoesNotPay {
				return sequenceConditionPlan{}, sequenceTaxOwnerCategory, false
			}
			plan.externalConditions = append(plan.externalConditions, condition)
			if !optionalFlow.consumesCondition(ci) {
				plan.gateConditions = append(plan.gateConditions, condition)
			}
			continue
		}
		if !counterTaxConditionMatches(content.Effects[ei].Payment, condition) ||
			len(plan.clauseConditions[ei]) != 0 {
			return sequenceConditionPlan{}, sequenceTaxOwnerCategory, false
		}
		plan.clauseConditions[ei] = []compiler.CompiledCondition{condition}
	}
	if len(plan.clauseConditions) != len(owners) {
		return sequenceConditionPlan{}, sequenceTaxOwnerCategory, false
	}
	// A payment's "if they do" cannot be attributed to the counter action.
	// Keep such outcome-dependent flows closed until payment provenance is modeled.
	if optionalFlow.enabled {
		if optionalFlow.scoped != nil {
			for producer := range optionalFlow.scoped.publishers {
				if _, owned := plan.clauseConditions[producer]; owned {
					return sequenceConditionPlan{}, "structural — counter payment outcome flow not modeled", false
				}
			}
		}
		if _, owned := plan.clauseConditions[optionalFlow.optionalIndex]; owned {
			return sequenceConditionPlan{}, "structural — counter payment outcome flow not modeled", false
		}
	}
	gates, reason, ok := matchOrderedSequenceEffectConditions(content.Effects, plan.gateConditions)
	plan.gates = gates
	return plan, reason, ok
}

func effectOwnsCounterTax(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCounter &&
		effect.Payment.Form == parser.EffectPaymentFormUnless &&
		effect.Payment.Payer == parser.EffectPaymentPayerTargetController
}

func counterTaxConditionMatches(payment compiler.CompiledEffectPayment, condition compiler.CompiledCondition) bool {
	// Unless carries the compiler's grammatical negation. The clause lowerer
	// realizes it as a failed payment result, not a negated state predicate.
	return payment.Form == parser.EffectPaymentFormUnless &&
		payment.Payer == parser.EffectPaymentPayerTargetController &&
		payment.AdditionalCost == nil &&
		payment.FailureConditionNodeID >= 0 &&
		condition.NodeID == payment.FailureConditionNodeID &&
		condition.Kind == compiler.ConditionUnless &&
		condition.Predicate == compiler.ConditionPredicateTargetControllerDoesNotPay &&
		condition.Negated &&
		!condition.Intervening
}

func (plan sequenceConditionPlan) referencesForClause(content compiler.AbilityContent, ei int) []compiler.CompiledReference {
	references := referencesOutsideOwnedConditions(content.Effects[ei].References, plan.externalConditions)
	// The compiler may attribute a trailing payment's references only to the
	// content. Restore them to its NodeID-selected owner exactly once.
	for _, condition := range plan.clauseConditions[ei] {
		for _, reference := range content.References {
			if !conditionOwnsReference(condition, reference.NodeID) {
				continue
			}
			alreadyOwned := false
			for _, existing := range references {
				if existing.NodeID == reference.NodeID {
					alreadyOwned = true
					break
				}
			}
			if !alreadyOwned {
				references = append(references, reference)
			}
		}
	}
	return references
}
