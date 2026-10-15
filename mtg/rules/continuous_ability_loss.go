package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func needsAbilityLayerEvaluation(g *game.Game, sources []staticAbilitySource) bool {
	for index := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[index]
		if effect.Layer == game.LayerAbility && abilityEffectMayDepend(effect) {
			return true
		}
	}
	for _, source := range sources {
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			for index := range body.ContinuousEffects {
				effect := &body.ContinuousEffects[index]
				if effect.Layer == game.LayerAbility &&
					(abilityEffectMayDepend(effect) || body.Condition.Exists) {
					return true
				}
			}
		}
	}
	return false
}

func abilityEffectMayDepend(effect *game.ContinuousEffect) bool {
	return effect.RemoveAllAbilities || len(effect.DependsOn) > 0 || selectionReadsKeywords(effect.Group.Selection())
}

func selectionReadsKeywords(selection game.Selection) bool {
	if selection.Keyword != game.KeywordNone || selection.ExcludedKeyword != game.KeywordNone {
		return true
	}
	return slices.ContainsFunc(selection.AnyOf, selectionReadsKeywords)
}

func needsColorLayerEvaluation(g *game.Game, sources []staticAbilitySource) bool {
	needed := false
	for index := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[index]
		if effect.Layer == game.LayerColor && (len(effect.DependsOn) > 0 || selectionReadsColors(effect.Group.Selection())) {
			needed = true
			break
		}
	}
	if !needed {
		for _, source := range sources {
			for index := range source.face.StaticAbilities {
				body := &source.face.StaticAbilities[index]
				if !staticAbilityFunctionsFromSource(body, source) {
					continue
				}
				for index := range body.ContinuousEffects {
					effect := &body.ContinuousEffects[index]
					if effect.Layer == game.LayerColor &&
						(body.Condition.Exists || len(effect.DependsOn) > 0 || selectionReadsColors(effect.Group.Selection())) {
						needed = true
						break
					}
				}
			}
		}
	}
	return needed
}

func applyDependentColorLayer(g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, sources []staticAbilitySource) bool {
	if !needsColorLayerEvaluation(g, sources) {
		return false
	}
	*values = cloneTypeLayerValues(continuousBoardLayer(g, recipient, sources, game.LayerColor).values[recipient.ObjectID])
	return true
}

func selectionReadsColors(selection game.Selection) bool {
	if len(selection.ColorsAny) > 0 || len(selection.ExcludedColors) > 0 ||
		selection.Colorless || selection.Colored || selection.Multicolored || selection.ColorChoice != game.ColorChoiceNone {
		return true
	}
	return slices.ContainsFunc(selection.AnyOf, selectionReadsColors)
}

func applyAbilityLossLayer(g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, sources []staticAbilitySource) bool {
	if !needsAbilityLayerEvaluation(g, sources) {
		return false
	}
	*values = cloneTypeLayerValues(abilityLossLayer(g, recipient, sources).values[recipient.ObjectID])
	return true
}

type abilityLayerEffect struct {
	landTypeEffect

	effectIndex int
}

// Ability removal can erase an unapplied generator anywhere on the battlefield.
// Resolve this layer globally, retaining begun effects' recipients for CR 613.6.
func abilityLossLayer(g *game.Game, recipient *game.Permanent, sources []staticAbilitySource) *landTypeLayerResult {
	return continuousBoardLayer(g, recipient, sources, game.LayerAbility)
}

func continuousBoardLayer(g *game.Game, recipient *game.Permanent, sources []staticAbilitySource, layer game.ContinuousLayer) *landTypeLayerResult {
	fc := frameCacheFor(g)
	if fc != nil && fc.layerBoards[layer] != nil {
		if _, ok := fc.layerBoards[layer].values[recipient.ObjectID]; ok {
			return fc.layerBoards[layer]
		}
	}
	var permanents []*game.Permanent
	found := false
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) {
			permanents = append(permanents, permanent)
			found = found || permanent.ObjectID == recipient.ObjectID
		}
	}
	if !found {
		permanents = append(permanents, recipient)
	}
	result := &landTypeLayerResult{
		values:  make(map[id.ID]permanentEffectiveValues, len(permanents)),
		started: make(map[staticBodyKey]map[id.ID]bool),
	}
	for _, permanent := range permanents {
		values := cloneTypeLayerValues(permanentValuesBeforeLayer(g, permanent, layer))
		if layer == game.LayerType && values.keywords.has(game.Changeling) {
			values.subtypes = appendUniqueSubtypes(values.subtypes, types.SubtypesForType(types.Creature)...)
		}
		result.values[permanent.ObjectID] = values
	}
	remaining := continuousLayerCandidates(g, sources, layer)
	if layer == game.LayerAbility {
		for _, permanent := range permanents {
			for _, effect := range keywordCounterEffects(permanent) {
				remaining = append(remaining, abilityLayerEffect{landTypeEffect: landTypeEffect{effect: effect}})
			}
		}
	}
	slices.SortStableFunc(remaining, func(a, b abilityLayerEffect) int {
		return compareContinuousEffects(&a.effect, &b.effect)
	})
	for len(remaining) > 0 {
		adjacency := make([][]int, len(remaining))
		for j, other := range remaining {
			after := make(map[id.ID]permanentEffectiveValues, len(result.values))
			for _, permanent := range permanents {
				values := cloneTypeLayerValues(result.values[permanent.ObjectID])
				if abilityLossEffectApplies(g, other, permanent, result.values, sources) {
					applyAbilityLayerCandidate(g, permanent, &values, other)
				}
				after[permanent.ObjectID] = values
			}
			for i, candidate := range remaining {
				if i == j {
					continue
				}
				depends := other.effect.ID != 0 && slices.Contains(candidate.effect.DependsOn, other.effect.ID)
				for _, permanent := range permanents {
					if abilityLossEffectApplies(g, candidate, permanent, result.values, sources) !=
						abilityLossEffectApplies(g, candidate, permanent, after, sources) {
						depends = true
						break
					}
				}
				if depends {
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
		chosen := remaining[next]
		var affected []*game.Permanent
		for _, permanent := range permanents {
			if abilityLossEffectApplies(g, chosen, permanent, result.values, sources) {
				affected = append(affected, permanent)
			}
		}
		for _, permanent := range affected {
			values := result.values[permanent.ObjectID]
			applyAbilityLayerCandidate(g, permanent, &values, chosen)
			result.values[permanent.ObjectID] = values
			if chosen.body != nil {
				if result.started[chosen.key] == nil {
					result.started[chosen.key] = make(map[id.ID]bool)
				}
				result.started[chosen.key][permanent.ObjectID] = true
			}
		}
		remaining = slices.Delete(remaining, next, next+1)
	}
	if fc != nil && found {
		if fc.layerBoards == nil {
			fc.layerBoards = make(map[game.ContinuousLayer]*landTypeLayerResult)
		}
		fc.layerBoards[layer] = result
	}
	return result
}

func continuousLayerCandidates(g *game.Game, sources []staticAbilitySource, layer game.ContinuousLayer) []abilityLayerEffect {
	var effects []abilityLayerEffect
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == layer {
			effects = append(effects, abilityLayerEffect{landTypeEffect: landTypeEffect{effect: effect}})
		}
	}
	for _, source := range sources {
		if source.permanent != nil && layer != game.LayerControl {
			source.controller = effectiveController(g, source.permanent)
		}
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			for j, template := range body.ContinuousEffects {
				if template.Layer != layer || template.AffectedSource && !template.Group.Empty() {
					continue
				}
				effect := template
				effect.SourceObjectID, effect.SourceCardID = sourceObjectID(source), source.cardID
				effect.Controller, effect.Timestamp = source.controller, source.timestamp
				if effect.AffectedSource {
					if source.permanent == nil {
						continue
					}
					effect.AffectedObjectID = source.permanent.ObjectID
				}
				rewriteChosenColorProtectionEffect(&effect, source.permanent)
				rewriteCommanderIdentityProtectionEffect(g, &effect, source.controller)
				if layer == game.LayerControl && effect.NewController.Exists {
					effect.NewController.Val = source.controller
				}
				effects = append(effects, abilityLayerEffect{
					landTypeEffect: landTypeEffect{effect: effect, source: source, body: body, key: sourceBodyKey(source, i)},
					effectIndex:    j,
				})
			}
		}
	}
	return effects
}

func abilityLossEffectApplies(g *game.Game, candidate abilityLayerEffect, recipient *game.Permanent, board map[id.ID]permanentEffectiveValues, sources []staticAbilitySource) bool {
	layer := candidate.effect.Layer
	if candidate.body != nil {
		first := firstStaticLayer(candidate.body)
		if first == layer && candidate.source.permanent != nil &&
			!valuesHaveStaticBody(board[sourceObjectID(candidate.source)], candidate.body) {
			return false
		}
		if !landTypeStaticBodyApplies(g, candidate.source, candidate.key.index, candidate.body, recipient, layer, sources) {
			return false
		}
		if landTypeBodyContinues(g, candidate.source, candidate.key.index, candidate.body, recipient, layer, sources) {
			return true
		}
		if !conditionSatisfied(g, conditionContext{
			controller: candidate.source.controller, source: candidate.source.permanent,
			characteristicsBefore: layer, layerValues: board,
		}, candidate.body.Condition) {
			return false
		}
	}
	values := board[recipient.ObjectID]
	return continuousEffectApplies(g, recipient, &values, &candidate.effect)
}

func applyAbilityLayerCandidate(g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, candidate abilityLayerEffect) {
	effect := candidate.effect
	if candidate.body != nil && len(effect.AddAbilities) > 0 {
		effect.AddAbilities = slices.Clone(effect.AddAbilities)
		if values.triggerGrantOrigins == nil {
			values.triggerGrantOrigins = make(map[*game.Ability]game.TriggerOrigin)
		}
		values.triggerGrantOrigins[&effect.AddAbilities[0]] = game.TriggerOrigin{
			SourceID: sourceObjectID(candidate.source), OriginID: candidate.source.originID,
			Face: candidate.source.faceIndex, StaticIndex: candidate.key.index,
			EffectIndex: candidate.effectIndex, Granted: true,
		}
	}
	applyContinuousEffect(g, recipient, values, &effect)
}
