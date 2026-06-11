package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type conditionLoweringContext uint8

const (
	conditionContextStatic conditionLoweringContext = iota
	conditionContextActivation
	conditionContextInterveningTrigger
	conditionContextReplacement
)

// lowerCondition is the single semantic Condition to game.Condition adapter.
// The explicit context prevents a structurally valid predicate from being used
// in an ability shell whose runtime does not evaluate it.
func lowerCondition(condition oracle.CompiledCondition, ctx conditionLoweringContext) (game.Condition, bool) {
	if !conditionKindAllowedInContext(condition, ctx) ||
		!conditionPredicateAllowedInContext(condition.Predicate, ctx) {
		return game.Condition{}, false
	}
	result := game.Condition{
		Text:   condition.Text,
		Negate: condition.Negated,
	}
	switch condition.Predicate {
	case oracle.ConditionPredicateControllerLifeAtLeast:
		result.ControllerLifeAtLeast = condition.Threshold
	case oracle.ConditionPredicateControllerHandSizeAtLeast:
		result.ControllerHandSizeAtLeast = condition.Threshold
	case oracle.ConditionPredicateAnyPlayerLifeAtMost:
		result.AnyPlayerLifeAtMost = condition.Threshold
	case oracle.ConditionPredicateOpponentCountAtLeast:
		result.OpponentCountAtLeast = condition.Threshold
	case oracle.ConditionPredicateControllerControls:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.ControlsMatching = opt.Val(count)
	case oracle.ConditionPredicateAnyOpponentControls:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.AnyOpponentControls = opt.Val(count)
	case oracle.ConditionPredicateOpponentsControl:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.OpponentsControl = opt.Val(count)
	case oracle.ConditionPredicateControllerHandEmpty:
		result.ControllerHandEmpty = true
	case oracle.ConditionPredicateControllerGraveyardCardCountAtLeast:
		result.ControllerGraveyardCardCountAtLeast = condition.Threshold
	case oracle.ConditionPredicateControllerGraveyardCardTypeCountAtLeast:
		result.ControllerGraveyardCardTypeCountAtLeast = condition.Threshold
	case oracle.ConditionPredicateControllerCreaturePowerDiversityAtLeast:
		result.ControllerCreaturePowerDiversityAtLeast = condition.Threshold
	default:
		return game.Condition{}, false
	}
	return result, !result.Empty()
}

func conditionKindAllowedInContext(condition oracle.CompiledCondition, ctx conditionLoweringContext) bool {
	switch ctx {
	case conditionContextStatic:
		return condition.Kind == oracle.ConditionAsLongAs && !condition.Intervening
	case conditionContextActivation:
		return condition.Kind == oracle.ConditionOnlyIf && !condition.Intervening
	case conditionContextInterveningTrigger:
		return condition.Kind == oracle.ConditionIf && condition.Intervening
	case conditionContextReplacement:
		return condition.Kind == oracle.ConditionUnless && !condition.Intervening
	default:
		return false
	}
}

func conditionPredicateAllowedInContext(predicate oracle.ConditionPredicate, ctx conditionLoweringContext) bool {
	if ctx != conditionContextReplacement {
		switch predicate {
		case oracle.ConditionPredicateControllerLifeAtLeast,
			oracle.ConditionPredicateAnyPlayerLifeAtMost,
			oracle.ConditionPredicateOpponentCountAtLeast,
			oracle.ConditionPredicateControllerControls,
			oracle.ConditionPredicateAnyOpponentControls,
			oracle.ConditionPredicateOpponentsControl,
			oracle.ConditionPredicateControllerHandEmpty,
			oracle.ConditionPredicateControllerGraveyardCardCountAtLeast,
			oracle.ConditionPredicateControllerGraveyardCardTypeCountAtLeast,
			oracle.ConditionPredicateControllerCreaturePowerDiversityAtLeast:
			return true
		default:
			return ctx == conditionContextStatic &&
				predicate == oracle.ConditionPredicateControllerHandSizeAtLeast
		}
	}
	switch predicate {
	case oracle.ConditionPredicateControllerLifeAtLeast,
		oracle.ConditionPredicateAnyPlayerLifeAtMost,
		oracle.ConditionPredicateOpponentCountAtLeast,
		oracle.ConditionPredicateControllerControls,
		oracle.ConditionPredicateAnyOpponentControls,
		oracle.ConditionPredicateOpponentsControl:
		return true
	default:
		return false
	}
}

func lowerConditionSelectionCount(condition oracle.CompiledCondition) (game.SelectionCount, bool) {
	selection, ok := lowerConditionSelection(condition.Selection)
	if !ok {
		return game.SelectionCount{}, false
	}
	return game.SelectionCount{
		Selection: selection,
		MinCount:  condition.Threshold,
	}, !selection.Empty()
}

func lowerConditionSelection(selection oracle.ConditionSelection) (game.Selection, bool) {
	required, ok := lowerConditionCardTypes(selection.RequiredTypes)
	if !ok {
		return game.Selection{}, false
	}
	supertypes, ok := lowerConditionSupertypes(selection.Supertypes)
	if !ok {
		return game.Selection{}, false
	}
	colors, ok := lowerConditionColors(selection.ColorsAny)
	if !ok {
		return game.Selection{}, false
	}
	subtypes := make([]types.Sub, 0, len(selection.SubtypesAny))
	for _, subtype := range selection.SubtypesAny {
		if subtype == "" {
			return game.Selection{}, false
		}
		subtypes = append(subtypes, types.Sub(subtype))
	}
	result := game.Selection{
		RequiredTypes: required,
		Supertypes:    supertypes,
		SubtypesAny:   subtypes,
		ColorsAny:     colors,
		Colorless:     selection.Colorless,
		ExcludeSource: selection.ExcludeSource,
	}
	if selection.MatchPowerAtLeast {
		result.Power = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.PowerAtLeast})
	} else if selection.PowerAtLeast != 0 {
		return game.Selection{}, false
	}
	return result, len(result.Validate()) == 0
}

func lowerConditionCardTypes(values []oracle.ConditionCardType) ([]types.Card, bool) {
	result := make([]types.Card, 0, len(values))
	for _, value := range values {
		switch value {
		case oracle.ConditionCardTypeArtifact:
			result = append(result, types.Artifact)
		case oracle.ConditionCardTypeBattle:
			result = append(result, types.Battle)
		case oracle.ConditionCardTypeCreature:
			result = append(result, types.Creature)
		case oracle.ConditionCardTypeEnchantment:
			result = append(result, types.Enchantment)
		case oracle.ConditionCardTypeLand:
			result = append(result, types.Land)
		case oracle.ConditionCardTypePlaneswalker:
			result = append(result, types.Planeswalker)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerConditionSupertypes(values []oracle.ConditionSupertype) ([]types.Super, bool) {
	result := make([]types.Super, 0, len(values))
	for _, value := range values {
		switch value {
		case oracle.ConditionSupertypeBasic:
			result = append(result, types.Basic)
		case oracle.ConditionSupertypeSnow:
			result = append(result, types.Snow)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerConditionColors(values []oracle.ConditionColor) ([]color.Color, bool) {
	result := make([]color.Color, 0, len(values))
	for _, value := range values {
		switch value {
		case oracle.ConditionColorWhite:
			result = append(result, color.White)
		case oracle.ConditionColorBlue:
			result = append(result, color.Blue)
		case oracle.ConditionColorBlack:
			result = append(result, color.Black)
		case oracle.ConditionColorRed:
			result = append(result, color.Red)
		case oracle.ConditionColorGreen:
			result = append(result, color.Green)
		default:
			return nil, false
		}
	}
	return result, true
}
