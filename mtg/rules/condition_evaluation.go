package rules

import "github.com/natefinch/council4/mtg/game"

func (r *effectResolver) instructionConditionSatisfied(instr *game.Instruction) bool {
	satisfied := effectConditionSatisfied(r.game, r.obj, instr.Condition)
	if instr.PublishCondition != "" {
		if !instr.Condition.Exists {
			panic("rules: condition publication without a condition")
		}
		if r.conditionEvaluations == nil {
			r.conditionEvaluations = make(map[game.ConditionKey]bool)
		}
		available := true
		if instr.Condition.Val.Condition.Exists {
			available = objectConditionInformationAvailable(r.game, conditionContext{
				controller: stackObjectController(r.obj), sourceObjectID: r.obj.SourceID, obj: r.obj,
			}, &instr.Condition.Val.Condition.Val)
		}
		if available {
			r.conditionEvaluations[instr.PublishCondition] = satisfied
		} else {
			delete(r.conditionEvaluations, instr.PublishCondition)
		}
	}
	if instr.ConditionGate != "" {
		value, published := r.conditionEvaluations[instr.ConditionGate]
		if !published || value == instr.ConditionGateNegate {
			return false
		}
	}
	return satisfied
}
