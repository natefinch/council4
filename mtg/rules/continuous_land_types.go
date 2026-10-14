package rules

import (
	"maps"
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type staticBodyKey struct {
	object id.ID
	origin id.ID
	face   game.FaceIndex
	index  int
}

func sourceBodyKey(source staticAbilitySource, index int) staticBodyKey {
	return staticBodyKey{sourceObjectID(source), source.originID, source.faceIndex, index}
}

type landTypeEffect struct {
	effect game.ContinuousEffect
	source staticAbilitySource
	body   *game.StaticAbility
	key    staticBodyKey
}

type landTypeLayerResult struct {
	values  map[id.ID]permanentEffectiveValues
	started map[staticBodyKey]map[id.ID]bool
}

func hasBasicLandTypeSet(g *game.Game, sources []staticAbilitySource) bool {
	for i := range g.ContinuousEffects {
		if len(g.ContinuousEffects[i].SetBasicLandTypes) > 0 {
			return true
		}
	}
	for _, source := range sources {
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			for j := range body.ContinuousEffects {
				effect := &body.ContinuousEffects[j]
				if (len(effect.SetBasicLandTypes) > 0 ||
					effect.Layer == game.LayerType && firstStaticLayer(body) == game.LayerType && len(body.ContinuousEffects) > 1) &&
					staticAbilityFunctionsFromSource(body, source) {
					return true
				}
			}
		}
	}
	return false
}

// A type SET can erase another type effect's source (CR 305.7, 613.8a).
// Multi-layer type effects also need their complete recipient set (CR 613.6).
// Evaluate this bounded layer across the board, not separately per recipient:
// otherwise an unaffected basic land would incorrectly see Urborg's effect.
// Other boards retain the existing direct layer path.
func applyBasicLandTypeLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, sources []staticAbilitySource) bool {
	if !hasBasicLandTypeSet(g, sources) {
		return false
	}
	result := basicLandTypeLayer(g, permanent, sources)
	*values = cloneTypeLayerValues(result.values[permanent.ObjectID])
	return true
}

func basicLandTypeLayer(g *game.Game, recipient *game.Permanent, sources []staticAbilitySource) *landTypeLayerResult {
	fc := frameCacheFor(g)
	if fc != nil && fc.landTypes != nil {
		if _, ok := fc.landTypes.values[recipient.ObjectID]; ok {
			return fc.landTypes
		}
	}
	permanents := make([]*game.Permanent, 0, len(g.Battlefield)+1)
	found := false
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) {
			permanents = append(permanents, permanent)
			found = found || permanent.ObjectID == recipient.ObjectID
		}
	}
	if !found {
		// CR 614.12: include a prospective entrant without publishing it into
		// the battlefield or allowing its own statics to affect other objects.
		permanents = append(permanents, recipient)
	}
	result := &landTypeLayerResult{
		values:  make(map[id.ID]permanentEffectiveValues, len(permanents)),
		started: make(map[staticBodyKey]map[id.ID]bool),
	}
	for _, permanent := range permanents {
		values := cloneTypeLayerValues(permanentValuesBeforeLayer(g, permanent, game.LayerType))
		if values.keywords.has(game.Changeling) {
			values.subtypes = appendUniqueSubtypes(values.subtypes, types.SubtypesForType(types.Creature)...)
		}
		result.values[permanent.ObjectID] = values
	}
	remaining := landTypeEffects(g, sources)
	slices.SortStableFunc(remaining, func(a, b landTypeEffect) int {
		return compareContinuousEffects(&a.effect, &b.effect)
	})
	for len(remaining) > 0 {
		adjacency := make([][]int, len(remaining))
		for j, other := range remaining {
			after := simulateLandTypeEffect(g, permanents, result.values, other)
			for i, candidate := range remaining {
				if i == j {
					continue
				}
				depends := other.effect.ID != 0 && slices.Contains(candidate.effect.DependsOn, other.effect.ID)
				if !depends {
					for _, permanent := range permanents {
						if landTypeEffectApplies(g, candidate, permanent, result.values) !=
							landTypeEffectApplies(g, candidate, permanent, after) {
							depends = true
							break
						}
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
		// Capture the whole affected set before applying to any member. Removing
		// the generating ability from its source cannot truncate a begun effect.
		var affected []*game.Permanent
		for _, permanent := range permanents {
			if landTypeEffectApplies(g, chosen, permanent, result.values) {
				affected = append(affected, permanent)
			}
		}
		for _, permanent := range affected {
			values := result.values[permanent.ObjectID]
			applyTypeLayer(g, &values, &chosen.effect)
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
		fc.landTypes = result
	}
	return result
}

func landTypeEffects(g *game.Game, sources []staticAbilitySource) []landTypeEffect {
	var effects []landTypeEffect
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == game.LayerType {
			effects = append(effects, landTypeEffect{effect: effect})
		}
	}
	for _, source := range sources {
		if source.permanent != nil {
			source.controller = effectiveController(g, source.permanent)
		}
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityFunctionsFromSource(body, source) ||
				!conditionSatisfied(g, conditionContext{
					controller: source.controller, source: source.permanent, characteristicsBefore: game.LayerType,
				}, body.Condition) {
				continue
			}
			for _, template := range body.ContinuousEffects {
				if template.Layer != game.LayerType || template.AffectedSource && !template.Group.Empty() {
					continue
				}
				effect := template
				effect.SourceObjectID = sourceObjectID(source)
				effect.SourceCardID = source.cardID
				effect.Controller = source.controller
				effect.Timestamp = source.timestamp
				if effect.AffectedSource {
					if source.permanent == nil {
						continue
					}
					effect.AffectedObjectID = source.permanent.ObjectID
				}
				effects = append(effects, landTypeEffect{effect: effect, source: source, body: body, key: sourceBodyKey(source, i)})
			}
		}
	}
	return effects
}

func landTypeEffectApplies(g *game.Game, candidate landTypeEffect, permanent *game.Permanent, board map[id.ID]permanentEffectiveValues) bool {
	if candidate.body != nil && candidate.source.permanent != nil &&
		firstStaticLayer(candidate.body) == game.LayerType &&
		!valuesHaveStaticBody(board[sourceObjectID(candidate.source)], candidate.body) {
		return false
	}
	values := board[permanent.ObjectID]
	return continuousEffectApplies(g, permanent, &values, &candidate.effect)
}

func simulateLandTypeEffect(g *game.Game, permanents []*game.Permanent, board map[id.ID]permanentEffectiveValues, candidate landTypeEffect) map[id.ID]permanentEffectiveValues {
	after := make(map[id.ID]permanentEffectiveValues, len(board))
	for _, permanent := range permanents {
		values := cloneTypeLayerValues(board[permanent.ObjectID])
		if landTypeEffectApplies(g, candidate, permanent, board) {
			applyTypeLayer(g, &values, &candidate.effect)
		}
		after[permanent.ObjectID] = values
	}
	return after
}

func cloneTypeLayerValues(values permanentEffectiveValues) permanentEffectiveValues {
	values.colors = slices.Clone(values.colors)
	values.types = slices.Clone(values.types)
	values.supertypes = slices.Clone(values.supertypes)
	values.subtypes = slices.Clone(values.subtypes)
	values.abilities = slices.Clone(values.abilities)
	values.resolutionUses = maps.Clone(values.resolutionUses)
	values.triggerOrigins = maps.Clone(values.triggerOrigins)
	values.triggerGrantOrigins = maps.Clone(values.triggerGrantOrigins)
	return values
}

func valuesHaveStaticBody(values permanentEffectiveValues, body *game.StaticAbility) bool {
	return slices.ContainsFunc(values.abilities, func(ability game.Ability) bool {
		static, ok := ability.(*game.StaticAbility)
		return ok && (static == body || reflect.DeepEqual(static, body))
	})
}

func firstStaticLayer(body *game.StaticAbility) game.ContinuousLayer {
	first := game.LayerPowerToughnessSwitch + 1
	for _, effect := range body.ContinuousEffects {
		first = min(first, effect.Layer)
	}
	return first
}

func landTypeStaticBodyApplies(g *game.Game, source staticAbilitySource, index int, body *game.StaticAbility, recipient *game.Permanent, layer game.ContinuousLayer, sources []staticAbilitySource) bool {
	if source.permanent == nil || layer <= game.LayerType {
		return true
	}

	first := firstStaticLayer(body)
	if first == game.LayerType && hasBasicLandTypeSet(g, sources) {
		return basicLandTypeLayer(g, recipient, sources).started[sourceBodyKey(source, index)][recipient.ObjectID]
	}
	if first == game.LayerAbility && layer > first && hasAbilityLoss(g, sources) {
		return abilityLossLayer(g, recipient, sources).started[sourceBodyKey(source, index)][recipient.ObjectID]
	}
	// CR 613.6: an effect begun before type-layer ability loss continues.
	// Effects beginning later require an ability still present at that boundary.
	values := permanentValuesBeforeLayer(g, source.permanent, first)
	return valuesHaveStaticBody(values, body)
}

func landTypeBodyContinues(g *game.Game, source staticAbilitySource, index int, body *game.StaticAbility, recipient *game.Permanent, layer game.ContinuousLayer, sources []staticAbilitySource) bool {
	if source.permanent == nil {
		return false
	}
	switch firstStaticLayer(body) {
	case game.LayerType:
		return layer > game.LayerType && hasBasicLandTypeSet(g, sources) &&
			basicLandTypeLayer(g, recipient, sources).started[sourceBodyKey(source, index)][recipient.ObjectID]
	case game.LayerAbility:
		return layer > game.LayerAbility && hasAbilityLoss(g, sources) &&
			abilityLossLayer(g, recipient, sources).started[sourceBodyKey(source, index)][recipient.ObjectID]
	default:
		return false
	}
}
