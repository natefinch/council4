package cardgen

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const conditionEvaluationCategory = "structural — condition evaluation ownership not modeled"

func contentWithoutOwnedConditionReferences(content compiler.AbilityContent) compiler.AbilityContent {
	result := content
	result.References = referencesOutsideOwnedConditions(content.References, content.Conditions)
	result.Effects = slices.Clone(content.Effects)
	for i := range result.Effects {
		result.Effects[i].References = referencesOutsideOwnedConditions(result.Effects[i].References, content.Conditions)
		result.Effects[i].SubjectReferences = referencesOutsideOwnedConditions(result.Effects[i].SubjectReferences, content.Conditions)
	}
	return result
}

func conditionsOwnedByResolvingBody(content compiler.AbilityContent) bool {
	if len(content.Conditions) == 0 ||
		(abilityContentHasAddManaEffect(content) && !abilityContentHasTargets(content)) {
		return false
	}
	for _, condition := range content.Conditions {
		if condition.Intervening || condition.Resolving ||
			(condition.Kind != compiler.ConditionIf && condition.Kind != compiler.ConditionUnless) {
			return false
		}
		if _, reason := conditionClauseIndices(content.Effects, condition); reason != "" {
			return false
		}
	}
	return true
}

func contentHasResolutionCount(content compiler.AbilityContent) bool {
	if slices.ContainsFunc(content.Conditions, func(condition compiler.CompiledCondition) bool {
		return condition.Predicate == compiler.ConditionPredicateSourceAbilityResolutionOrdinalThisTurn
	}) {
		return true
	}
	for _, mode := range content.Modes {
		if contentHasResolutionCount(mode.Content) {
			return true
		}
	}
	return false
}

func conditionClauseIndices(effects []compiler.CompiledEffect, condition compiler.CompiledCondition) ([]int, string) {
	ownership := condition.Ownership
	switch ownership.Scope {
	case parser.ConditionScopeClause:
		if len(ownership.ClauseIDs) != 1 {
			return nil, conditionEvaluationCategory
		}
	case parser.ConditionScopeGroup:
		if len(ownership.ClauseIDs) < 2 {
			return nil, conditionEvaluationCategory
		}
	case parser.ConditionScopeUnsupported:
		return nil, effectGateCategoryMultiClause
	default:
		return nil, effectGateCategoryNoClause
	}
	var indices []int
	for _, id := range ownership.ClauseIDs {
		found := -1
		for ei := range effects {
			if id > 0 && effects[ei].ClauseID == id {
				if found >= 0 {
					return nil, conditionEvaluationCategory
				}
				found = ei
			}
		}
		if found < 0 || len(indices) > 0 && found != indices[len(indices)-1]+1 {
			return nil, conditionEvaluationCategory
		}
		indices = append(indices, found)
	}
	return indices, ""
}

// The generic planner consumes parser-owned scope; specialized matchers keep
// their existing acceptance boundaries until individually migrated.
func matchOrderedSequenceEffectConditions(effects []compiler.CompiledEffect, conditions []compiler.CompiledCondition) (map[int]game.EffectCondition, string, bool) {
	gates := make(map[int]game.EffectCondition)
	for _, condition := range conditions {
		owners, reason := conditionClauseIndices(effects, condition)
		if reason != "" {
			return nil, reason, false
		}
		ctx := effectGateLoweringContext(condition)
		references, ok := sequenceConditionReferenceContext(condition, owners[0])
		if !ok {
			return nil, effectGateCategoryLowering, false
		}
		lowered, ok := lowerConditionWithReferences(condition, ctx, references)
		if !ok {
			return nil, effectGateRejectCategory(condition, ctx), false
		}
		for _, ei := range owners {
			if _, exists := gates[ei]; exists {
				return nil, effectGateCategoryMultiCondition, false
			}
			if condition.Kind == compiler.ConditionUnless &&
				(effects[ei].Replacement.Kind != parser.EffectReplacementNone ||
					ei+1 < len(effects) && effects[ei+1].Connection == parser.EffectConnectionOtherwise) {
				return nil, effectGateCategoryUnlessBranch, false
			}
			if effects[ei].Context == parser.EffectContextTarget && conditionTestsSourceObjectCharacteristics(lowered) {
				return nil, effectGateCategoryPredicate, false
			}
			gates[ei] = game.EffectCondition{Condition: opt.Val(lowered)}
		}
	}
	return gates, "", true
}

type conditionEvaluationMember struct {
	index  int
	negate bool
}

func (plan sequenceConditionPlan) captureEvaluations(
	effects []compiler.CompiledEffect,
	ranges [][2]int,
	sequence []game.Instruction,
	insteadGates, otherwiseGates map[int]game.EffectCondition,
) bool {
	for _, condition := range plan.gateConditions {
		owners, reason := conditionClauseIndices(effects, condition)
		if reason != "" {
			return false
		}
		var members []conditionEvaluationMember
		appendClause := func(ei int, negate bool) bool {
			if ei < 0 || ei >= len(ranges) || ranges[ei][0] == ranges[ei][1] {
				return false
			}
			for i := ranges[ei][0]; i < ranges[ei][1]; i++ {
				members = append(members, conditionEvaluationMember{index: i, negate: negate})
			}
			return true
		}
		firstOwner := owners[0]
		_, replaced := insteadGates[firstOwner-1]
		if replaced && !appendClause(firstOwner-1, false) {
			return false
		}
		for _, ei := range owners {
			if !appendClause(ei, replaced) {
				return false
			}
		}
		after := owners[len(owners)-1] + 1
		if _, otherwise := otherwiseGates[after]; otherwise && !appendClause(after, !replaced) {
			return false
		}
		if len(members) < 2 {
			continue
		}
		first := &sequence[members[0].index]
		if !first.Condition.Exists || first.PublishCondition != "" || first.ConditionGate != "" {
			return false
		}
		key := game.ConditionKey(fmt.Sprintf("condition-%d", condition.NodeID))
		complement, complementOK := negatedEffectCondition(&first.Condition.Val)
		for _, member := range members[1:] {
			instr := &sequence[member.index]
			expected := first.Condition
			if member.negate {
				if !complementOK {
					return false
				}
				expected = opt.Val(complement)
			}
			if !reflect.DeepEqual(instr.Condition, expected) ||
				instr.PublishCondition != "" || instr.ConditionGate != "" {
				return false
			}
			instr.Condition = opt.V[game.EffectCondition]{}
			instr.ConditionGate = key
			instr.ConditionGateNegate = member.negate
		}
		first.PublishCondition = key
	}
	return true
}

func conditionOwnsReference(condition compiler.CompiledCondition, nodeID int) bool {
	return slices.Contains(condition.Ownership.ReferenceNodeIDs, nodeID)
}

func referencesOutsideOwnedConditions(references []compiler.CompiledReference, conditions []compiler.CompiledCondition) []compiler.CompiledReference {
	result := make([]compiler.CompiledReference, 0, len(references))
	for _, reference := range references {
		owned := false
		for _, condition := range conditions {
			if conditionOwnsReference(condition, reference.NodeID) {
				owned = true
				break
			}
		}
		if !owned {
			result = append(result, reference)
		}
	}
	return result
}
