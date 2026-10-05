package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func targetCardConditionResultKey(clauseID int) game.ResultKey {
	return game.ResultKey(fmt.Sprintf("condition-target-card-%d", clauseID))
}

func publishConditionTargetCards(content compiler.AbilityContent, ranges [][2]int, sequence []game.Instruction) bool {
	for _, condition := range content.Conditions {
		producer := condition.TargetCardProducerClauseID
		if producer == 0 {
			continue
		}
		index := -1
		for ei, effect := range content.Effects {
			if effect.ClauseID == producer {
				if index >= 0 {
					return false
				}
				index = ei
			}
		}
		if index < 0 || index >= len(ranges) || ranges[index][1] != ranges[index][0]+1 ||
			ranges[index][0] < 0 || ranges[index][1] > len(sequence) {
			return false
		}
		instruction := &sequence[ranges[index][0]]
		move, ok := instruction.Primitive.(game.MoveCard)
		if !ok || move.Card.Kind != game.CardReferenceTarget {
			return false
		}
		expected := targetCardConditionResultKey(producer)
		key := instruction.PublishResult
		if key == "" {
			key = expected
			instruction.PublishResult = key
		}
		found := false
		for i := range sequence {
			gate := &sequence[i].Condition
			if !gate.Exists || !gate.Val.Condition.Exists ||
				(gate.Val.Condition.Val.TargetCardResultKey != expected && gate.Val.Condition.Val.TargetCardResultKey != key) {
				continue
			}
			if i <= ranges[index][0] {
				return false
			}
			gate.Val.Condition.Val.TargetCardResultKey = key
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}
