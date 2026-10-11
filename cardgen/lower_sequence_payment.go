package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

const sequencePaymentIdentityCategory = "structural — embedded payment has no unique clause-owned result identity"

func sequencePaymentKey(clauseID int) game.ResultKey {
	return game.ResultKey(fmt.Sprintf("payment-clause-%d", clauseID))
}

func counterPaymentActorCondition(content compiler.AbilityContent, producer int, condition compiler.CompiledCondition) bool {
	if producer < 0 || producer >= len(content.Effects) ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
		return false
	}
	effect := content.Effects[producer]
	if !effectOwnsCounterTax(effect) || !effect.Exact || effect.Negated ||
		effect.Optional || effect.DelayedTiming != 0 || effect.ClauseID <= 0 ||
		condition.Ownership.ResultProducerClauseID != effect.ClauseID ||
		condition.Ownership.ResultSubjectReferenceNodeID < 0 ||
		len(effect.Targets) != 1 || len(condition.Ownership.ReferenceNodeIDs) != 1 {
		return false
	}
	target := effect.Targets[0]
	if target.Selector.Kind != compiler.SelectorSpell ||
		target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return false
	}
	matches := 0
	for _, reference := range content.References {
		if !slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID) {
			continue
		}
		if reference.Kind != compiler.ReferenceThatPlayer &&
			(reference.Kind != compiler.ReferencePronoun || reference.Pronoun != compiler.ReferencePronounThey) {
			return false
		}
		if !reference.SubjectSupported() || reference.Binding != compiler.ReferenceBindingTarget ||
			reference.NodeID != condition.Ownership.ResultSubjectReferenceNodeID ||
			reference.Occurrence < 0 || reference.Occurrence >= len(content.Targets) ||
			reference.Occurrence != condition.Ownership.ResultSubjectTargetOccurrence ||
			content.Targets[reference.Occurrence].Order != target.Order {
			return false
		}
		for ti, other := range content.Targets {
			if ti != reference.Occurrence && other.Order == target.Order {
				return false
			}
		}
		matches++
	}
	return matches == 1
}

func (plan sequenceConditionPlan) bindPaymentResult(content compiler.AbilityContent, flow optionalFlowPlan, ei int, sequence []game.Instruction) bool {
	if _, owned := plan.clauseConditions[ei]; !owned {
		return true
	}
	paymentActor := flow.scoped != nil && flow.scoped.paymentPublishers[ei] != ""
	if len(plan.clauseConditions) <= 1 && !paymentActor {
		return true
	}
	index, ok := counterActionInstruction(sequence)
	if !ok || index != 1 || sequence[0].Optional || sequence[0].PublishResult != unlessPaidResultKey {
		return false
	}
	clauseID := content.Effects[ei].ClauseID
	if clauseID <= 0 {
		return false
	}
	key := sequencePaymentKey(clauseID)
	sequence[0].PublishResult = key
	sequence[0].LocalProducts.Results = append(sequence[0].LocalProducts.Results, key)
	sequence[1].ResultGate.Val.Key = key
	return true
}
