package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func (flow *scopedResultFlow) planPostfixElse(content compiler.AbilityContent, indices map[int]int) string {
	flow.postfixElse = make(map[int]int)
	for ei, effect := range content.Effects {
		ownership := effect.OptionalPostfixElse
		if ownership.ActionClauseID == 0 {
			continue
		}
		producer, exists := indices[ownership.ActionClauseID]
		ci, count := -1, 0
		for index, condition := range content.Conditions {
			if condition.NodeID == ownership.ConditionNodeID {
				ci, count = index, count+1
			}
		}
		if !exists || producer != ei-1 || count != 1 || effect.Optional ||
			effect.Connection != parser.EffectConnectionOtherwise || effect.ResultElseOfClauseID != 0 ||
			flow.actions.isCompound(producer) || !flow.optional[producer] {
			return "structural — optional postfix Otherwise has no unique singleton action/condition"
		}
		condition := content.Conditions[ci]
		owners, reason := conditionClauseIndices(content.Effects, condition)
		if reason != "" || len(owners) != 1 || owners[0] != producer ||
			condition.Kind != compiler.ConditionIf || condition.Intervening ||
			isResolvingSuccessGate(condition.Predicate) ||
			condition.Predicate == compiler.ConditionPredicatePriorInstructionNotAccepted ||
			content.Effects[producer].DelayedTiming != 0 {
			return "structural — optional postfix Otherwise predicate ownership not modeled"
		}
		flow.postfixElse[ei] = producer
		flow.publishers[producer] = game.ResultKey(fmt.Sprintf("result-clause-%d", ownership.ActionClauseID))
	}
	return ""
}

// The fallback is two disjoint branches: the frozen predicate's complement,
// or the same true predicate plus actual declined acceptance. No failed or
// skipped instruction is relabeled as a published result.
func (flow *scopedResultFlow) appendPostfixElseBranches(
	ranges [][2]int, sequence []game.Instruction,
) ([]game.Instruction, bool) {
	if flow == nil || len(flow.postfixElse) == 0 {
		return sequence, true
	}
	branches := make(map[int][]game.Instruction)
	for ei, producer := range flow.postfixElse {
		first, end := ranges[ei][0], ranges[ei][1]
		action := sequence[ranges[producer][0]]
		if first == end || action.PublishCondition == "" || action.PublishResult != flow.publishers[producer] {
			return nil, false
		}
		branch := slices.Clone(sequence[first:end])
		for i := range branch {
			instruction := &branch[i]
			if instruction.Condition.Exists || instruction.ConditionGate != action.PublishCondition ||
				!instruction.ConditionGateNegate || instruction.Optional ||
				instruction.PublishCondition != "" || instruction.PublishResult != "" ||
				instruction.PublishOptionalDecision != "" || instruction.OptionalDecisionGate != "" {
				return nil, false
			}
			instruction.ConditionGateNegate = false
		}
		gate := game.InstructionResultGate{Key: flow.publishers[producer], Accepted: game.TriFalse}
		var ok bool
		branches[end], ok = appendResultGatedBranch(nil, gate, branch)
		if !ok {
			return nil, false
		}
	}
	var result []game.Instruction
	for i, instruction := range sequence {
		result = append(result, instruction)
		result = append(result, branches[i+1]...)
	}
	return result, true
}
