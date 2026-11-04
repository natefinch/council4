package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// A static control effect's new controller is its source's controller ("You
// control enchanted creature." on Control Magic). When another layer-2 effect
// changes that source's controller (Word of Seizing gaining control of the
// Aura), the static effect depends on it (CR 613.8a): it waits until that
// effect has been applied and then gives the permanents it affects to the
// source's new controller (CR 613.8b). Dependency loops apply in timestamp
// order (CR 613.8b), so the whole layer is evaluated once for the battlefield
// rather than recursively per permanent. When no static control source is
// itself under a control effect, the per-permanent calculation is already
// exact and remains the fast path.

type controlBoardState uint8

const (
	controlBoardUnknown controlBoardState = iota
	controlBoardBuilding
	controlBoardIndependent
	controlBoardBuilt
)

type controlLayerCandidate struct {
	effect game.ContinuousEffect
	// source, when set, supplies the new controller: the static ability's
	// source permanent's controller at the time this effect is applied.
	source   *game.Permanent
	resolved bool
	affected []*game.Permanent
}

// controlLayerBoard returns every battlefield permanent's controller after
// layer 2, or false when no static control effect depends on another control
// effect (or the board is being built and a nested read must use the
// per-permanent calculation).
func controlLayerBoard(g *game.Game, fc *frameCache) (map[*game.Permanent]game.PlayerID, bool) {
	if fc == nil {
		return nil, false
	}
	switch fc.controlBoardState {
	case controlBoardBuilt:
		return fc.controlBoard, true
	case controlBoardUnknown:
	default:
		return nil, false
	}
	fc.controlBoardState = controlBoardBuilding
	sources := staticAbilitySources(g)
	if !controlSourceUnderControlEffect(g, sources) {
		fc.controlBoardState = controlBoardIndependent
		return nil, false
	}
	fc.controlBoard = buildControlLayerBoard(g, sources)
	fc.controlBoardState = controlBoardBuilt
	return fc.controlBoard, true
}

// applyControlLayerBoard applies layer 2 to values from the battlefield's
// control board and reports whether it did.
func applyControlLayerBoard(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) bool {
	board, ok := controlLayerBoard(g, frameCacheFor(g))
	if !ok {
		return false
	}
	controller, found := board[permanent]
	if !found {
		return false
	}
	if controller != values.controller {
		values.controller = controller
		recalculateDynamicPT(g, permanent, values)
	}
	return true
}

// controlSourceUnderControlEffect reports whether some battlefield permanent
// with a source-relative static control effect is itself affected by a
// layer-2 effect.
func controlSourceUnderControlEffect(g *game.Game, sources []staticAbilitySource) bool {
	for _, source := range sources {
		if source.permanent == nil || !sourceHasSourceRelativeControl(source) {
			continue
		}
		values := basePermanentValues(g, source.permanent)
		if len(continuousEffectsForLayer(g, source.permanent, &values, game.LayerControl, sources)) > 0 {
			return true
		}
	}
	return false
}

func sourceHasSourceRelativeControl(source staticAbilitySource) bool {
	if source.face == nil {
		return false
	}
	for i := range source.face.StaticAbilities {
		body := &source.face.StaticAbilities[i]
		if !staticAbilityFunctionsFromSource(body, source) {
			continue
		}
		for j := range body.ContinuousEffects {
			if body.ContinuousEffects[j].Layer == game.LayerControl && body.ContinuousEffects[j].NewController.Exists {
				return true
			}
		}
	}
	return false
}

// buildControlLayerBoard applies every layer-2 effect to the battlefield in
// dependency and timestamp order (CR 613.7, 613.8). The set each effect
// applies to is determined from the permanents' pre-layer-2 values, as the
// per-permanent calculation does; only a source-relative new controller is
// read as the effects are applied.
func buildControlLayerBoard(g *game.Game, sources []staticAbilitySource) map[*game.Permanent]game.PlayerID {
	permanents := g.Battlefield
	base := make([]permanentEffectiveValues, len(permanents))
	for i, permanent := range permanents {
		base[i] = basePermanentValues(g, permanent)
	}
	var candidates []controlLayerCandidate
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer != game.LayerControl {
			continue
		}
		candidate := controlLayerCandidate{effect: *effect, resolved: true}
		for j, permanent := range permanents {
			if continuousEffectApplies(g, permanent, &base[j], effect) {
				candidate.affected = append(candidate.affected, permanent)
			}
		}
		if len(candidate.affected) > 0 {
			candidates = append(candidates, candidate)
		}
	}
	type staticKey struct{ source, static, effect int }
	staticCandidates := make(map[staticKey]int)
	for s, source := range sources {
		if !sourceStaticAbilityHasLayer(source, game.LayerControl) {
			continue
		}
		for j, permanent := range permanents {
			visitStaticAbilitySourceContinuousEffects(g, source, permanent, &base[j], game.LayerControl, sources, func(effect game.ContinuousEffect, staticIndex, effectIndex int) {
				key := staticKey{source: s, static: staticIndex, effect: effectIndex}
				k, ok := staticCandidates[key]
				if !ok {
					k = len(candidates)
					staticCandidates[key] = k
					candidate := controlLayerCandidate{effect: effect}
					if effect.NewController.Exists {
						candidate.source = source.permanent
					}
					candidates = append(candidates, candidate)
				}
				candidates[k].affected = append(candidates[k].affected, permanent)
			})
		}
	}

	controllers := make(map[*game.Permanent]game.PlayerID, len(permanents))
	for i, permanent := range permanents {
		controllers[permanent] = base[i].controller
	}
	applyControlLayerCandidates(g, candidates, controllers)
	return controllers
}

// controlCandidateController returns the player candidate gives control to
// when applied to controllers, or false when it currently gives control to no
// one (an ended grant, or no living monarch).
func controlCandidateController(g *game.Game, candidate *controlLayerCandidate, controllers map[*game.Permanent]game.PlayerID) (game.PlayerID, bool) {
	if candidate.effect.NewControllerIsMonarch {
		if monarch := livingMonarch(g); monarch.Exists {
			return monarch.Val, true
		}
	}
	newController := candidate.effect.NewController
	if candidate.source != nil {
		newController.Val = controllers[candidate.source]
	}
	if !newController.Exists || g.ControlGrantsEnded[newController.Val] {
		return 0, false
	}
	return newController.Val, true
}

// applyControlLayerCandidates applies candidates to controllers in timestamp
// order, overridden by dependencies (CR 613.8b), with each dependency loop
// applied in timestamp order once its outside dependencies have been applied.
// Dependencies are determined anew from the current layer-2 state after each
// effect is applied (CR 613.8c).
func applyControlLayerCandidates(g *game.Game, candidates []controlLayerCandidate, controllers map[*game.Permanent]game.PlayerID) {
	remaining := make([]int, len(candidates))
	for i := range remaining {
		remaining[i] = i
	}
	slices.SortStableFunc(remaining, func(a, b int) int {
		return compareContinuousEffects(&candidates[a].effect, &candidates[b].effect)
	})
	for len(remaining) > 0 {
		adjacency := make([][]int, len(remaining))
		for i, a := range remaining {
			for j, b := range remaining {
				if i != j && controlCandidateDependsOn(g, candidates, controllers, a, b) {
					adjacency[i] = append(adjacency[i], j)
				}
			}
		}
		reachable := reachabilityMatrix(adjacency)
		next := 0
		for i := range remaining {
			if effectComponentApplicable(i, adjacency, reachable) {
				next = i
				break
			}
		}
		candidate := &candidates[remaining[next]]
		if controller, ok := controlCandidateController(g, candidate, controllers); ok {
			for _, permanent := range candidate.affected {
				controllers[permanent] = controller
			}
		}
		remaining = slices.Delete(remaining, next, next+1)
	}
}

// controlCandidateDependsOn reports whether candidate a depends on candidate b
// in the current layer-2 state (CR 613.8a): applying b now would change the
// controller of a's source, so it would change what a does. An effect that
// would give the source to the player already controlling it changes nothing
// and isn't a dependency. A resolved effect's members share its timestamp (one
// effect locked to several objects, CR 611.2c), so a also waits for b when
// another member of b's effect would change a's source. Declared dependencies
// also count.
func controlCandidateDependsOn(g *game.Game, candidates []controlLayerCandidate, controllers map[*game.Permanent]game.PlayerID, a, b int) bool {
	dependent, dependency := &candidates[a], &candidates[b]
	if dependency.effect.ID != 0 && slices.Contains(dependent.effect.DependsOn, dependency.effect.ID) {
		return true
	}
	if dependent.source == nil {
		return false
	}
	changesSource := func(member *controlLayerCandidate) bool {
		if !slices.Contains(member.affected, dependent.source) {
			return false
		}
		controller, ok := controlCandidateController(g, member, controllers)
		return ok && controller != controllers[dependent.source]
	}
	if changesSource(dependency) {
		return true
	}
	if !dependency.resolved {
		return false
	}
	for i := range candidates {
		member := &candidates[i]
		if i != a && i != b && member.resolved && member.effect.Timestamp == dependency.effect.Timestamp && changesSource(member) {
			return true
		}
	}
	return false
}
