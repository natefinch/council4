package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
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
	if cond.Negate {
		return !matches
	}
	return matches
}

func controllerControlsMatchingPermanent(g *game.Game, ctx conditionContext, filter game.PermanentFilter) bool {
	want := filter.MinCount
	if want <= 0 {
		want = 1
	}
	count := 0
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
		if count >= want {
			return true
		}
	}
	return false
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
	if len(filter.SubtypesAny) > 0 && !slices.ContainsFunc(filter.SubtypesAny, func(subtype string) bool {
		return slices.Contains(values.subtypes, subtype)
	}) {
		return false
	}
	if filter.Power.Exists {
		if useBase {
			return false
		}
		if !values.powerOK || !intComparisonMatches(values.power, filter.Power.Val) {
			return false
		}
	}
	if filter.Toughness.Exists {
		if useBase {
			return false
		}
		if !values.toughnessOK || !intComparisonMatches(values.toughness, filter.Toughness.Val) {
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
