package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
)

func sequenceConditionReferenceContext(condition compiler.CompiledCondition, firstOwner int) (referenceLoweringContext, bool) {
	if condition.ObjectBinding != compiler.ReferenceBindingPriorInstructionResult || condition.ObjectReference == nil {
		return referenceLoweringContext{}, true
	}
	prior := condition.ObjectReference.PriorInstruction
	if prior < 0 || prior >= firstOwner {
		return referenceLoweringContext{}, false
	}
	return referenceLoweringContext{
		PriorInstruction: prior,
		PriorLinkedKey:   sequenceProductKey(prior),
	}, true
}

func (plan sequenceConditionPlan) resultReferencesForClause(effects []compiler.CompiledEffect, index int) []compiler.CompiledReference {
	var references []compiler.CompiledReference
	for _, condition := range plan.gateConditions {
		if condition.ObjectBinding != compiler.ReferenceBindingPriorInstructionResult || condition.ObjectReference == nil {
			continue
		}
		owners, reason := conditionClauseIndices(effects, condition)
		if reason == "" && owners[0] == index {
			references = append(references, *condition.ObjectReference)
		}
	}
	return references
}
