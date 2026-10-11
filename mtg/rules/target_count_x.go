package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func specsBindTargetCountsToX(specs []game.TargetSpec) bool {
	for _, spec := range specs {
		if spec.CountEqualsX {
			return true
		}
	}
	return false
}

func targetCountsMatchX(specs []game.TargetSpec, counts []int, xValue int) bool {
	for i, spec := range specs {
		if spec.CountEqualsX && (i >= len(counts) || counts[i] != xValue) {
			return false
		}
	}
	return true
}

func bodyTargetCountsMatchX(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, chosenModes []int, targets []game.Target, xValue int) bool {
	specs := bodyTargetSpecs(body, chosenModes)
	if !specsBindTargetCountsToX(specs) {
		return true
	}
	// Unlike ordinary fixed targets, an ambiguous packed layout cannot prove
	// which target spec is bound to X.
	counts, ok := bodyTargetCountsWithModesAndRecorded(g, controller, source, sourceObjectID, body, chosenModes, nil, targets)
	return ok && targetCountsMatchX(specs, counts, xValue)
}
