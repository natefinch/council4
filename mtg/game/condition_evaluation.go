package game

import "fmt"

func validateConditionEvaluations(seq []Instruction) error {
	published := make(map[ConditionKey]int)
	for i := range seq {
		instr := &seq[i]
		if instr.ConditionGateNegate && instr.ConditionGate == "" {
			return fmt.Errorf("instruction[%d]: ConditionGateNegate requires ConditionGate", i)
		}
		if instr.ConditionGate != "" {
			if _, ok := published[instr.ConditionGate]; !ok {
				return fmt.Errorf("instruction[%d]: ConditionGate references key %q not yet published", i, instr.ConditionGate)
			}
		}
		if instr.PublishCondition != "" {
			if instr.ConditionGate != "" {
				return fmt.Errorf("instruction[%d]: condition publication and consumption cannot share an instruction", i)
			}
			if !instr.Condition.Exists {
				return fmt.Errorf("instruction[%d]: PublishCondition requires Condition", i)
			}
			if previous, duplicate := published[instr.PublishCondition]; duplicate {
				return fmt.Errorf("instruction[%d]: duplicate condition key %q (first used at index %d)", i, instr.PublishCondition, previous)
			}
			published[instr.PublishCondition] = i
		}
		// The Tempting offer body dispatches primitives directly, not instruction
		// envelopes. It cannot represent scoped condition evaluations.
		for _, body := range instr.TemptingOfferBody {
			if body.PublishCondition != "" || body.ConditionGate != "" || body.ConditionGateNegate {
				return fmt.Errorf("instruction[%d]: condition evaluation inside TemptingOfferBody not supported", i)
			}
		}
	}
	return nil
}
