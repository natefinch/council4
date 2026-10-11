package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
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
	if !ok {
		return true
	}
	transform := func(kind targetIndexKind, occurrence int) (int, bool) {
		if kind != targetIndexObject || occurrence < 0 || occurrence >= len(targets) {
			return 0, false
		}
		index, ok := indices[targets[occurrence].Span]
		return index, ok
	}
	gate, ok = transformEffectConditionTargetIndices(gate, transform)
	if !ok {
		return false
	}
	gates[i] = gate
	return true
}
