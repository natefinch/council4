package rules

import "github.com/natefinch/council4/mtg/game"

func (r *effectResolver) publishOptionalDecision(instruction *game.Instruction, accepted bool) {
	if instruction.PublishOptionalDecision == "" {
		return
	}
	if r.optionalDecisions == nil {
		r.optionalDecisions = make(map[game.OptionalDecisionKey]bool)
	}
	r.optionalDecisions[instruction.PublishOptionalDecision] = accepted
}
