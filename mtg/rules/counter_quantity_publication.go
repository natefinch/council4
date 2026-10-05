package rules

import "github.com/natefinch/council4/mtg/game"

func (r *effectResolver) clearCounterQuantityPublication(instruction *game.Instruction) {
	if r.obj == nil || instruction.PublishResult == "" || instruction.Primitive == nil ||
		instruction.Primitive.Kind() != game.PrimitiveRemoveCounter {
		return
	}
	key := string(instruction.PublishResult)
	delete(r.obj.ResolvedAmounts, key)
	delete(r.obj.ResolvedExcessDamage, key)
	delete(r.obj.ResolutionResults, key)
	delete(r.obj.ResolutionResultObjects, key)
}
