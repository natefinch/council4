package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type conditionContext struct {
	controller             game.PlayerID
	source                 *game.Permanent
	event                  *game.GameEvent
	obj                    *game.StackObject
	useBaseCharacteristics bool
}

func conditionSatisfied(g *game.Game, ctx conditionContext, condition opt.V[game.Condition]) bool {
	if !condition.Exists || condition.Val.Empty() {
		return true
	}
	cond := condition.Val
	matches := true
	if !cond.ControllerControls.Empty() {
		matches = matches && controllerControlsMatchingPermanent(g, ctx, cond.ControllerControls)
	}
	if cond.Object.Exists || len(cond.Types) > 0 {
		matches = matches && conditionObjectMatches(g, ctx, cond)
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		matches = matches && eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g, ctx)
	}
	if cond.SourceClassLevelAtLeast > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel >= cond.SourceClassLevelAtLeast
	}
	if cond.SourceClassLevelLessThan > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel < cond.SourceClassLevelLessThan
	}
	if cond.SourceNotMonstrous {
		matches = matches && ctx.source != nil && !ctx.source.Monstrous
	}
	if cond.ControllerHasMaxSpeed {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Speed >= 4
	}
	if cond.Negate {
		return !matches
	}
	return matches
}

func conditionObjectMatches(g *game.Game, ctx conditionContext, cond game.Condition) bool {
	obj := ctx.obj
	if obj == nil && ctx.event != nil {
		obj = &game.StackObject{HasTriggerEvent: true, TriggerEvent: *ctx.event, Controller: ctx.controller}
	}
	ref := game.ObjectReference{Kind: game.ObjectReferenceEventPermanent}
	if cond.Object.Exists {
		ref = cond.Object.Val
	}
	resolved, ok := resolveObjectReference(g, obj, ref)
	if !ok {
		return false
	}
	for _, cardType := range cond.Types {
		if !resolvedObjectHasType(g, resolved, cardType) {
			return false
		}
	}
	return true
}

func resolvedObjectHasType(g *game.Game, resolved resolvedObjectReference, cardType types.Card) bool {
	if resolved.permanent != nil {
		return permanentHasType(g, resolved.permanent, cardType)
	}
	return slices.Contains(resolved.snapshot.Types, cardType)
}

func controllerControlsMatchingPermanent(g *game.Game, ctx conditionContext, filter game.PermanentFilter) bool {
	want := filter.MinCount
	if want <= 0 {
		want = 1
	}
	count := 0
	totalPower := 0
	for _, permanent := range g.Battlefield {
		if ctx.useBaseCharacteristics {
			if permanent.Controller != ctx.controller {
				continue
			}
		} else if effectiveController(g, permanent) != ctx.controller {
			continue
		}
		if !permanentMatchesConditionFilter(g, permanent, filter, ctx.useBaseCharacteristics) {
			continue
		}
		count++
		if filter.TotalPower.Exists {
			values := effectivePermanentValues(g, permanent)
			if values.powerOK {
				totalPower += values.power
			}
		}
		if count >= want {
			if !filter.TotalPower.Exists || filter.TotalPower.Val.Matches(totalPower) {
				return true
			}
		}
	}
	if filter.TotalPower.Exists {
		return count >= want && filter.TotalPower.Val.Matches(totalPower)
	}
	return false
}

func eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g *game.Game, ctx conditionContext) bool {
	if ctx.event == nil || ctx.event.PermanentID == 0 {
		return false
	}
	resolved, ok := resolvePermanentOrLastKnown(g, ctx.event.PermanentID)
	if !ok {
		return false
	}
	name := resolvedObjectName(g, resolved)
	if name == "" {
		return false
	}
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == ctx.event.PermanentID || effectiveController(g, permanent) != ctx.controller || !permanentHasType(g, permanent, types.Creature) {
			continue
		}
		if def, ok := permanentCardDef(g, permanent); ok && def.Name == name {
			return false
		}
	}
	player, ok := playerByID(g, ctx.controller)
	if !ok {
		return false
	}
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if def.Name == name && def.HasType(types.Creature) {
			return false
		}
	}
	return true
}

func resolvedObjectName(g *game.Game, resolved resolvedObjectReference) string {
	if resolved.permanent != nil {
		if resolved.permanent.Token {
			return permanentTokenName(resolved.permanent)
		}
		return permanentEffectiveName(g, resolved.permanent)
	}
	return resolved.snapshot.Name
}

func permanentMatchesConditionFilter(g *game.Game, permanent *game.Permanent, filter game.PermanentFilter, useBase bool) bool {
	var values permanentEffectiveValues
	if useBase {
		values = basePermanentValues(g, permanent)
	} else {
		values = effectivePermanentValues(g, permanent)
	}
	for _, cardType := range filter.Types {
		if !slices.Contains(values.types, cardType) {
			return false
		}
	}
	for _, supertype := range filter.Supertypes {
		if !slices.Contains(values.supertypes, supertype) {
			return false
		}
	}
	if len(filter.SubtypesAny) > 0 && !slices.ContainsFunc(filter.SubtypesAny, func(subtype types.Sub) bool {
		return slices.Contains(values.subtypes, subtype)
	}) {
		return false
	}
	if filter.Power.Exists {
		if useBase {
			return false
		}
		if !values.powerOK || !filter.Power.Val.Matches(values.power) {
			return false
		}
	}
	if filter.Toughness.Exists {
		if useBase {
			return false
		}
		if !values.toughnessOK || !filter.Toughness.Val.Matches(values.toughness) {
			return false
		}
	}
	return true
}

func activationConditionSatisfied(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef) bool {
	if ability == nil {
		return false
	}
	return conditionSatisfied(g, conditionContext{
		controller: playerID,
		source:     permanent,
	}, ability.ActivationCondition)
}
