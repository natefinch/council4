package rules

import (
	"maps"
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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
		effect := &g.ContinuousEffects[i]
		if effect.Layer == game.LayerType &&
			(len(effect.SetBasicLandTypes) > 0 || len(effect.DependsOn) > 0 || selectionReadsTypes(effect.Group.Selection())) {
			return true
		}
	}
	for _, source := range sources {
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			for j := range body.ContinuousEffects {
				effect := &body.ContinuousEffects[j]
				if effect.Layer == game.LayerType &&
					(len(effect.SetBasicLandTypes) > 0 || len(effect.DependsOn) > 0 ||
						selectionReadsTypes(effect.Group.Selection()) || body.Condition.Exists || len(body.ContinuousEffects) > 1) &&
					staticAbilityFunctionsFromSource(body, source) {
					return true
				}
			}
		}
	}
	return false
}

func selectionReadsTypes(selection game.Selection) bool {
	if len(selection.RequiredTypes) > 0 || len(selection.RequiredTypesAny) > 0 || len(selection.ExcludedTypes) > 0 ||
		len(selection.Supertypes) > 0 || selection.ExcludedSupertype != "" ||
		len(selection.SubtypesAny) > 0 || selection.ExcludedSubtype != "" ||
		selection.SubtypeChoice != game.SubtypeChoiceNone {
		return true
	}
	return slices.ContainsFunc(selection.AnyOf, selectionReadsTypes)
}

// Type-changing effects can alter recipients or erase another effect's source
// (CR 305.7, 613.8a). Independent fixed changes retain the direct layer path.
func applyBasicLandTypeLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, sources []staticAbilitySource) bool {
	if !hasBasicLandTypeSet(g, sources) {
		return false
	}
	result := basicLandTypeLayer(g, permanent, sources)
	*values = cloneTypeLayerValues(result.values[permanent.ObjectID])
	return true
}

func basicLandTypeLayer(g *game.Game, recipient *game.Permanent, sources []staticAbilitySource) *landTypeLayerResult {
	return continuousBoardLayer(g, recipient, sources, game.LayerType)
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
	if first == game.LayerColor && layer > first && needsColorLayerEvaluation(g, sources) {
		return continuousBoardLayer(g, recipient, sources, first).started[sourceBodyKey(source, index)][recipient.ObjectID]
	}
	if first == game.LayerAbility && layer > first && needsAbilityLayerEvaluation(g, sources) {
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
		return layer > game.LayerAbility && needsAbilityLayerEvaluation(g, sources) &&
			abilityLossLayer(g, recipient, sources).started[sourceBodyKey(source, index)][recipient.ObjectID]
	case game.LayerColor:
		return layer > game.LayerColor && needsColorLayerEvaluation(g, sources) &&
			continuousBoardLayer(g, recipient, sources, game.LayerColor).started[sourceBodyKey(source, index)][recipient.ObjectID]
	default:
		return false
	}
}
