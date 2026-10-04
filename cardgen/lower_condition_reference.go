package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// Condition references arrive in ability-wide Oracle occurrence coordinates,
// unlike a clause primitive's local targets. Remap them only after the clause
// has materialized its targets, including any inherited target.
func remapSequenceConditionTargets(
	i int,
	gates map[int]game.EffectCondition,
	targets []compiler.CompiledTarget,
	indices map[shared.Span]int,
) bool {
	gate, ok := gates[i]
	if !ok || !gate.Condition.Exists || !gate.Condition.Val.Object.Exists {
		return true
	}
	condition := gate.Condition.Val
	object := condition.Object.Val
	kind := object.Kind()
	if kind != game.ObjectReferenceTargetPermanent && kind != game.ObjectReferenceTargetCard {
		return true
	}
	occurrence := object.TargetIndex()
	if occurrence < 0 || occurrence >= len(targets) {
		return false
	}
	index, ok := indices[targets[occurrence].Span]
	if !ok {
		return false
	}
	if kind == game.ObjectReferenceTargetCard {
		object = game.TargetCardReference(index)
	} else {
		object = game.TargetPermanentReference(index)
	}
	condition.Object = opt.Val(object)
	gate.Condition = opt.Val(condition)
	gates[i] = gate
	return true
}
