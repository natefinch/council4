package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// Retargeting preserves the number and position of targets and evaluates only
// the final set (CR 115.7e). Changing targets offers each slot only another
// legal target, so no combination exists unless every target changes (CR
// 115.7a). Choosing new targets also offers each slot its unchanged target,
// even an illegal one (CR 115.7d). The host stack object is never offered as
// one of its own new targets (CR 115.5).
func retargetChoices(g *game.Game, controller game.PlayerID, specs stackObjectSpecs, event game.Event,
	targets []game.Target, recorded []int, xValue int, rule game.RetargetRule) targetChoiceResult {
	keepUnchanged := rule == game.RetargetChooseNewTargets
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
			if keepUnchanged {
				options = append(options, original)
			} else {
				options = slices.DeleteFunc(options, func(target game.Target) bool { return sameRetargetTarget(g, target, original) })
			}
			for i, target := range options {
				if slices.Contains(options[:i], target) ||
					(target != original || !keepUnchanged) &&
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

// sameRetargetTarget reports whether candidate is the same object as an
// original target. Card candidates carry no zone version until bound, so a card
// that changed zones since it was targeted is a different object (CR 400.7).
func sameRetargetTarget(g *game.Game, candidate, original game.Target) bool {
	if candidate.Kind == game.TargetCard && !candidate.CardZoneVersionSet && original.CardZoneVersionSet {
		bound, ok := bindCardTargetZoneVersions(g, []game.Target{candidate})
		if !ok {
			return false
		}
		candidate = bound[0]
	}
	return candidate == original
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
