package rules

import "github.com/natefinch/council4/mtg/game"

func sequencePublishesResult(sequence []game.Instruction, key game.ResultKey) bool {
	if key == "" {
		return false
	}
	for _, instruction := range sequence {
		if instruction.PublishResult == key {
			return true
		}
	}
	return false
}

func isolateResultProducts(obj *game.StackObject, key string) func() {
	restoreReceipt := isolateProductCell(&obj.ResolutionResults, key)
	restoreObjects := isolateProductCell(&obj.ResolutionResultObjects, key)
	restoreAmount := isolateProductCell(&obj.ResolvedAmounts, key)
	restoreExcess := isolateProductCell(&obj.ResolvedExcessDamage, key)
	return func() {
		restoreExcess()
		restoreAmount()
		restoreObjects()
		restoreReceipt()
	}
}
