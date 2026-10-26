package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// Retargeting preserves the number and position of targets (CR 115.7d).
// An unchanged target may remain illegal; only new targets must be legal. The
// host stack object is never offered as one of its own new targets (CR 115.5).
func retargetChoices(g *game.Game, controller game.PlayerID, specs stackObjectSpecs, event game.Event,
	targets []game.Target, recorded []int, xValue int, keepIllegal bool) targetChoiceResult {
	counts, ok := resolutionTargetCounts(specs.specs, recorded, len(targets))
	if !ok {
		return targetChoiceResult{kind: targetInvalidSpec}
	}
	var choices [][]game.Target
	var appendSlots func(int, int, []game.Target)
	appendSlots = func(specIndex, targetIndex int, prefix []game.Target) {
		if specIndex == len(specs.specs) {
			choices = append(choices, slices.Clone(prefix))
			return
		}
		spec := normalizeTargetSpecForEvent(&specs.specs[specIndex], event)
		candidates := targetCandidatesForSpec(g, controller, specs.def, specs.sourceObjectID, event, &spec)
		if specs.hostObjectID != 0 {
			candidates = slices.DeleteFunc(candidates, func(target game.Target) bool {
				return target.Kind == game.TargetStackObject && target.StackObjectID == specs.hostObjectID
			})
		}
		var appendMembers func(int, []game.Target)
		appendMembers = func(member int, next []game.Target) {
			if member == counts[specIndex] {
				if targetsSatisfySameGraveyard(g, &spec, next[len(prefix):]) {
					appendSlots(specIndex+1, targetIndex+member, next)
				}
				return
			}
			original := targets[targetIndex+member]
			options := slices.Clone(candidates)
			if keepIllegal {
				options = append(options, original)
			}
			for i, target := range options {
				if slices.Contains(options[:i], target) ||
					(target != original || !keepIllegal) &&
						((spec.ManaValueAtMostX && !targetManaValueAtMost(g, target, xValue)) ||
							(spec.ManaValueEqualsX && !targetManaValueEquals(g, target, xValue))) {
					continue
				}
				if len(filterTargetsDistinctFrom([]game.Target{target}, next[len(prefix):])) == 0 ||
					(spec.DistinctFromPriorTargets && len(filterTargetsDistinctFrom([]game.Target{target}, prefix)) == 0) {
					continue
				}
				appendMembers(member+1, append(slices.Clone(next), target))
			}
		}
		appendMembers(0, prefix)
	}
	appendSlots(0, 0, nil)
	result := targetChoiceResult{kind: targetLegalChoicesFound, choices: choices}
	// Every choice keeps the per-spec counts it was enumerated with, so a
	// stack object with several chosen modes (a fused split spell, or a
	// "choose two" spell) can still segment its new targets by mode when it
	// resolves.
	for range choices {
		result.targetCounts = append(result.targetCounts, slices.Clone(counts))
	}
	return result
}

func bindRetargetZoneVersions(g *game.Game, chosen, original []game.Target) ([]game.Target, bool) {
	bound := slices.Clone(chosen)
	for i, target := range bound {
		if i < len(original) && target == original[i] {
			continue
		}
		next, ok := bindCardTargetZoneVersions(g, []game.Target{target})
		if !ok {
			return nil, false
		}
		bound[i] = next[0]
	}
	return bound, true
}
