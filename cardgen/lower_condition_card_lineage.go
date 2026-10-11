package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const cardLineageConditionCategory = "structural — card predicate has no exact available incarnation alternatives"

type sequenceCardLineageCondition struct {
	owners    []int
	reference compiler.CompiledReference
	condition game.Condition
	key       game.ConditionKey
}

func planSequenceCardLineageConditions(content compiler.AbilityContent, conditions []compiler.CompiledCondition) ([]sequenceCardLineageCondition, string, bool) {
	var lineages []sequenceCardLineageCondition
	claimed := make(map[int]bool)
	for _, condition := range conditions {
		if condition.ObjectReference == nil || condition.ObjectReference.ReachedCardProducerClauseID == 0 {
			continue
		}
		reached, ok := condition.ObjectReference.CardLineageSubject(content.Effects)
		owners, reason := conditionClauseIndices(content.Effects, condition)
		if !ok || reason != "" || len(owners) == 0 || reached.PriorInstruction >= owners[0] ||
			!content.OwnsSubject(*condition.ObjectReference) {
			return nil, cardLineageConditionCategory, false
		}
		alternative := condition
		alternative.ObjectReference = &reached
		references, ok := sequenceConditionReferenceContext(alternative, owners[0])
		if !ok {
			return nil, cardLineageConditionCategory, false
		}
		lowered, ok := lowerConditionWithReferences(alternative, effectGateLoweringContext(alternative), references)
		if !ok || lowered.Negate || !lowered.Object.Exists || !lowered.ObjectMatches.Exists {
			return nil, cardLineageConditionCategory, false
		}
		for _, index := range owners {
			if claimed[index] || len(referencesOutsideOwnedConditions(content.Effects[index].References, content.Conditions)) != 0 ||
				len(content.Effects[index].Targets) != 0 {
				return nil, cardLineageConditionCategory, false
			}
			claimed[index] = true
		}
		lineages = append(lineages, sequenceCardLineageCondition{
			owners: owners, reference: reached, condition: lowered,
			key: game.ConditionKey(fmt.Sprintf("condition-%d-reached-card", condition.NodeID)),
		})
	}
	return lineages, "", true
}

func (plan sequenceConditionPlan) publishCardLineageForClause(index int, sequence []game.Instruction, ranges [][2]int) bool {
	for _, lineage := range plan.cardLineages {
		if lineage.owners[0] != index {
			continue
		}
		if _, _, ok := sequencePriorInstructionLink([]compiler.CompiledReference{lineage.reference}, sequence, ranges); !ok {
			return false
		}
	}
	return true
}

// The input's versioned link and the actual entered product are disjoint.
// Each branch freezes its own predicate once, without inventing a success result.
func (plan sequenceConditionPlan) expandCardLineageConditions(sequence []game.Instruction, ranges [][2]int) ([]game.Instruction, bool) {
	branches := make(map[int][]game.Instruction)
	for _, lineage := range plan.cardLineages {
		start := ranges[lineage.owners[0]][0]
		end := ranges[lineage.owners[len(lineage.owners)-1]][1]
		if end <= start || end > len(sequence) {
			return nil, false
		}
		branch := slices.Clone(sequence[start:end])
		originalKey := branch[0].PublishCondition
		for i := range branch {
			instruction := &branch[i]
			if instruction.Optional || instruction.PublishResult != "" || instruction.ResultGate.Exists ||
				instruction.PublishOptionalDecision != "" || instruction.OptionalDecisionGate != "" ||
				instruction.ClearLinkedBeforeGate || len(instruction.LocalProducts.Links) != 0 ||
				len(instruction.LocalProducts.Results) != 0 {
				return nil, false
			}
			if i == 0 {
				if !instruction.Condition.Exists || instruction.ConditionGate != "" {
					return nil, false
				}
				instruction.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(lineage.condition)})
				if len(branch) > 1 {
					if originalKey == "" {
						return nil, false
					}
					instruction.PublishCondition = lineage.key
				}
			} else {
				if instruction.ConditionGate != originalKey || instruction.ConditionGateNegate ||
					instruction.PublishCondition != "" {
					return nil, false
				}
				instruction.ConditionGate = lineage.key
			}
		}
		branches[end] = append(branches[end], branch...)
	}
	var expanded []game.Instruction
	for i, instruction := range sequence {
		expanded = append(expanded, instruction)
		expanded = append(expanded, branches[i+1]...)
	}
	return expanded, true
}
