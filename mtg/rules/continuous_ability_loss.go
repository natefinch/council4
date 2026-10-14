package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func hasAbilityLoss(g *game.Game, sources []staticAbilitySource) bool {
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == game.LayerAbility && effect.RemoveAllAbilities {
			return true
		}
	}
	for _, source := range sources {
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			for _, effect := range body.ContinuousEffects {
				if effect.Layer == game.LayerAbility && effect.RemoveAllAbilities {
					return true
				}
			}
		}
	}
	return false
}

func applyAbilityLossLayer(g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, sources []staticAbilitySource) bool {
	if !hasAbilityLoss(g, sources) {
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
	fc := frameCacheFor(g)
	if fc != nil && fc.abilityLoss != nil {
		if _, ok := fc.abilityLoss.values[recipient.ObjectID]; ok {
			return fc.abilityLoss
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
		result.values[permanent.ObjectID] = cloneTypeLayerValues(permanentValuesBeforeLayer(g, permanent, game.LayerAbility))
	}
	remaining := abilityLossEffects(g, sources)
	for _, permanent := range permanents {
		for _, effect := range keywordCounterEffects(permanent) {
			remaining = append(remaining, abilityLayerEffect{landTypeEffect: landTypeEffect{effect: effect}})
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
		fc.abilityLoss = result
	}
	return result
}

func abilityLossEffects(g *game.Game, sources []staticAbilitySource) []abilityLayerEffect {
	var effects []abilityLayerEffect
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == game.LayerAbility {
			effects = append(effects, abilityLayerEffect{landTypeEffect: landTypeEffect{effect: effect}})
		}
	}
	for _, source := range sources {
		if source.permanent != nil {
			source.controller = effectiveController(g, source.permanent)
		}
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			for j, template := range body.ContinuousEffects {
				if template.Layer != game.LayerAbility || template.AffectedSource && !template.Group.Empty() {
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
	if candidate.body != nil {
		first := firstStaticLayer(candidate.body)
		if first == game.LayerAbility && candidate.source.permanent != nil &&
			!valuesHaveStaticBody(board[sourceObjectID(candidate.source)], candidate.body) {
			return false
		}
		if !landTypeStaticBodyApplies(g, candidate.source, candidate.key.index, candidate.body, recipient, game.LayerAbility, sources) {
			return false
		}
		if landTypeBodyContinues(g, candidate.source, candidate.key.index, candidate.body, recipient, game.LayerAbility, sources) {
			return true
		}
		if !conditionSatisfied(g, conditionContext{
			controller: candidate.source.controller, source: candidate.source.permanent,
			characteristicsBefore: game.LayerAbility,
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
