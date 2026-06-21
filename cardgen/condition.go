package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
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
	conditionContextEffectGate
	// conditionContextEntryCounters gates the "if ..." conditions on
	// enters-with-counters replacements ("This creature enters with a +1/+1
	// counter on it if you attacked this turn."). The runtime evaluates these
	// conditions as the permanent enters, with the entering permanent supplied as
	// the condition's source, so source-relative EventHistory predicates ("you
	// attacked this turn") and controller-scoped control predicates resolve.
	conditionContextEntryCounters
)

// lowerCondition is the single semantic Condition to game.Condition adapter.
// The explicit context prevents a structurally valid predicate from being used
// in an ability shell whose runtime does not evaluate it.
func lowerCondition(condition compiler.CompiledCondition, ctx conditionLoweringContext) (game.Condition, bool) {
	if !conditionKindAllowedInContext(condition, ctx) ||
		!conditionPredicateAllowedInContext(condition.Predicate, ctx) {
		return game.Condition{}, false
	}
	result := game.Condition{
		Text:   condition.Text,
		Negate: condition.Negated,
	}
	switch condition.Predicate {
	case compiler.ConditionPredicateControllerLifeAtLeast:
		result.ControllerLifeAtLeast = condition.Threshold
	case compiler.ConditionPredicateControllerHandSizeAtLeast:
		result.ControllerHandSizeAtLeast = condition.Threshold
	case compiler.ConditionPredicateControllerHandSizeExactly:
		result.ControllerHandSizeExactly = opt.Val(condition.Threshold)
	case compiler.ConditionPredicateAnyOpponentPoisonAtLeast:
		result.AnyOpponentPoisonAtLeast = condition.Threshold
	case compiler.ConditionPredicateAnyPlayerLifeAtMost:
		result.AnyPlayerLifeAtMost = condition.Threshold
	case compiler.ConditionPredicateOpponentCountAtLeast:
		result.OpponentCountAtLeast = condition.Threshold
	case compiler.ConditionPredicateControllerControls:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.ControlsMatching = opt.Val(count)
	case compiler.ConditionPredicateAnyOpponentControls:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.AnyOpponentControls = opt.Val(count)
	case compiler.ConditionPredicateOpponentsControl:
		count, ok := lowerConditionSelectionCount(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.OpponentsControl = opt.Val(count)
	case compiler.ConditionPredicateControlComparison:
		comparison, ok := lowerControlCountComparison(condition)
		if !ok {
			return game.Condition{}, false
		}
		result.ControlComparison = opt.Val(comparison)
	case compiler.ConditionPredicateControllerHandEmpty:
		result.ControllerHandEmpty = true
	case compiler.ConditionPredicateEventSubjectNameUnique:
		result.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures = true
	case compiler.ConditionPredicateControllerCreatedTokenThisTurn:
		result.ControllerCreatedTokenThisTurn = true
	case compiler.ConditionPredicateControllerGraveyardCardCountAtLeast:
		result.ControllerGraveyardCardCountAtLeast = condition.Threshold
	case compiler.ConditionPredicateControllerGraveyardCardTypeCountAtLeast:
		result.ControllerGraveyardCardTypeCountAtLeast = condition.Threshold
	case compiler.ConditionPredicateControllerCreaturePowerDiversityAtLeast:
		result.ControllerCreaturePowerDiversityAtLeast = condition.Threshold
	case compiler.ConditionPredicateObjectMatches:
		object, ok := lowerConditionObjectReference(condition.ObjectBinding)
		if !ok {
			return game.Condition{}, false
		}
		selection, ok := lowerConditionSelection(condition.Selection)
		if !ok || selection.Empty() {
			return game.Condition{}, false
		}
		result.Object = opt.Val(object)
		result.ObjectMatches = opt.Val(selection)
	case compiler.ConditionPredicateObjectExists:
		if condition.ObjectBinding != compiler.ReferenceBindingSource ||
			!conditionSelectionEmpty(condition.Selection) {
			return game.Condition{}, false
		}
		object, ok := lowerConditionObjectReference(condition.ObjectBinding)
		if !ok {
			return game.Condition{}, false
		}
		result.Object = opt.Val(object)
	case compiler.ConditionPredicateEventHistory:
		if condition.EventHistoryPattern == nil {
			return game.Condition{}, false
		}
		pattern, ok := lowerTriggerPattern(condition.EventHistoryPattern)
		if !ok {
			return game.Condition{}, false
		}
		window, ok := lowerEventHistoryWindow(condition.EventHistoryWindow)
		if !ok {
			return game.Condition{}, false
		}
		result.EventHistory = opt.Val(game.EventHistoryCondition{
			Pattern: pattern,
			Window:  window,
		})
	default:
		return game.Condition{}, false
	}
	return result, !result.Empty()
}

func conditionKindAllowedInContext(condition compiler.CompiledCondition, ctx conditionLoweringContext) bool {
	switch ctx {
	case conditionContextStatic:
		return condition.Kind == compiler.ConditionAsLongAs && !condition.Intervening
	case conditionContextActivation:
		return condition.Kind == compiler.ConditionOnlyIf && !condition.Intervening
	case conditionContextInterveningTrigger:
		return condition.Kind == compiler.ConditionIf && condition.Intervening
	case conditionContextReplacement:
		return condition.Kind == compiler.ConditionUnless && !condition.Intervening
	case conditionContextEntryCounters, conditionContextEffectGate:
		return condition.Kind == compiler.ConditionIf && !condition.Intervening
	default:
		return false
	}
}

func conditionPredicateAllowedInContext(predicate compiler.ConditionPredicate, ctx conditionLoweringContext) bool {
	if ctx == conditionContextEntryCounters {
		switch predicate {
		case compiler.ConditionPredicateControllerLifeAtLeast,
			compiler.ConditionPredicateAnyPlayerLifeAtMost,
			compiler.ConditionPredicateOpponentCountAtLeast,
			compiler.ConditionPredicateControllerControls,
			compiler.ConditionPredicateAnyOpponentControls,
			compiler.ConditionPredicateOpponentsControl,
			compiler.ConditionPredicateEventHistory:
			return true
		default:
			return false
		}
	}
	if ctx != conditionContextReplacement {
		switch predicate {
		case compiler.ConditionPredicateControllerLifeAtLeast,
			compiler.ConditionPredicateAnyPlayerLifeAtMost,
			compiler.ConditionPredicateOpponentCountAtLeast,
			compiler.ConditionPredicateControllerControls,
			compiler.ConditionPredicateAnyOpponentControls,
			compiler.ConditionPredicateOpponentsControl,
			compiler.ConditionPredicateControlComparison,
			compiler.ConditionPredicateControllerHandEmpty,
			compiler.ConditionPredicateControllerCreatedTokenThisTurn,
			compiler.ConditionPredicateControllerGraveyardCardCountAtLeast,
			compiler.ConditionPredicateControllerGraveyardCardTypeCountAtLeast,
			compiler.ConditionPredicateControllerCreaturePowerDiversityAtLeast,
			compiler.ConditionPredicateAnyOpponentPoisonAtLeast,
			compiler.ConditionPredicateObjectMatches,
			compiler.ConditionPredicateObjectExists:
			return true
		case compiler.ConditionPredicateEventHistory:
			return ctx == conditionContextInterveningTrigger || ctx == conditionContextActivation
		case compiler.ConditionPredicateEventSubjectNameUnique:
			return ctx == conditionContextInterveningTrigger
		case compiler.ConditionPredicateControllerHandSizeExactly:
			return ctx == conditionContextStatic || ctx == conditionContextActivation
		default:
			return ctx == conditionContextStatic &&
				predicate == compiler.ConditionPredicateControllerHandSizeAtLeast
		}
	}
	switch predicate {
	case compiler.ConditionPredicateControllerLifeAtLeast,
		compiler.ConditionPredicateAnyPlayerLifeAtMost,
		compiler.ConditionPredicateOpponentCountAtLeast,
		compiler.ConditionPredicateControllerControls,
		compiler.ConditionPredicateAnyOpponentControls,
		compiler.ConditionPredicateOpponentsControl,
		compiler.ConditionPredicateControlComparison:
		return true
	default:
		return false
	}
}

func lowerConditionSelectionCount(condition compiler.CompiledCondition) (game.SelectionCount, bool) {
	selection, ok := lowerConditionSelection(condition.Selection)
	if !ok {
		return game.SelectionCount{}, false
	}
	result := game.SelectionCount{
		Selection: selection,
		MinCount:  condition.Threshold,
	}
	if condition.Selection.MatchTotalPowerAtLeast {
		result.TotalPower = opt.Val(compare.Int{
			Op:    compare.GreaterOrEqual,
			Value: condition.Selection.TotalPowerAtLeast,
		})
	}
	return result, !selection.Empty()
}

// lowerControlCountComparison maps a typed cross-player control-count comparison
// onto the runtime form, failing closed unless exactly one side counts the
// controller and the counted Selection is non-empty.
func lowerControlCountComparison(condition compiler.CompiledCondition) (game.ControlCountComparison, bool) {
	selection, ok := lowerConditionSelection(condition.Selection)
	if !ok || selection.Empty() {
		return game.ControlCountComparison{}, false
	}
	left, ok := lowerComparisonScope(condition.ControlComparisonLeft)
	if !ok {
		return game.ControlCountComparison{}, false
	}
	right, ok := lowerComparisonScope(condition.ControlComparisonRight)
	if !ok {
		return game.ControlCountComparison{}, false
	}
	if (left == game.ControlPlayerController) == (right == game.ControlPlayerController) {
		return game.ControlCountComparison{}, false
	}
	op := compare.GreaterThan
	if !condition.ControlComparisonGreater {
		op = compare.LessThan
	}
	return game.ControlCountComparison{
		Selection: selection,
		Left:      left,
		Right:     right,
		Op:        op,
	}, true
}

func lowerComparisonScope(scope compiler.ConditionComparisonScope) (game.ControlPlayerScope, bool) {
	switch scope {
	case compiler.ConditionComparisonScopeController:
		return game.ControlPlayerController, true
	case compiler.ConditionComparisonScopeAnyOpponent:
		return game.ControlPlayerAnyOpponent, true
	case compiler.ConditionComparisonScopeEachOpponent:
		return game.ControlPlayerEachOpponent, true
	default:
		return game.ControlPlayerController, false
	}
}

func lowerConditionSelection(selection compiler.ConditionSelection) (game.Selection, bool) {
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
	tapped, ok := lowerConditionTriState(selection.Tapped)
	if !ok {
		return game.Selection{}, false
	}
	combatState, ok := lowerConditionCombatState(selection.CombatState)
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
		Multicolored:  selection.Multicolored,
		TokenOnly:     selection.TokenOnly,
		ExcludeSource: selection.ExcludeSource,
		Tapped:        tapped,
		CombatState:   combatState,
	}
	if selection.MatchPowerAtLeast {
		result.Power = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.PowerAtLeast})
	} else if selection.PowerAtLeast != 0 {
		return game.Selection{}, false
	}
	if selection.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selection.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		result.Keyword = keyword
	}
	return result, len(result.Validate()) == 0
}

func lowerConditionObjectReference(binding compiler.ReferenceBinding) (game.ObjectReference, bool) {
	return lowerObjectReference(compiler.CompiledReference{Binding: binding}, referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
}

func conditionSelectionEmpty(selection compiler.ConditionSelection) bool {
	lowered, ok := lowerConditionSelection(selection)
	return ok && lowered.Empty()
}

func lowerConditionTriState(value compiler.ConditionTriState) (game.TriState, bool) {
	switch value {
	case compiler.ConditionTriAny:
		return game.TriAny, true
	case compiler.ConditionTriTrue:
		return game.TriTrue, true
	case compiler.ConditionTriFalse:
		return game.TriFalse, true
	default:
		return game.TriAny, false
	}
}

func lowerConditionCombatState(value compiler.ConditionCombatState) (game.CombatStateFilter, bool) {
	switch value {
	case compiler.ConditionCombatStateAny:
		return game.CombatStateAny, true
	case compiler.ConditionCombatStateAttacking:
		return game.CombatStateAttacking, true
	case compiler.ConditionCombatStateBlocking:
		return game.CombatStateBlocking, true
	case compiler.ConditionCombatStateAttackingOrBlocking:
		return game.CombatStateAttackingOrBlocking, true
	default:
		return game.CombatStateAny, false
	}
}

func lowerConditionCardTypes(values []compiler.ConditionCardType) ([]types.Card, bool) {
	result := make([]types.Card, 0, len(values))
	for _, value := range values {
		switch value {
		case compiler.ConditionCardTypeArtifact:
			result = append(result, types.Artifact)
		case compiler.ConditionCardTypeBattle:
			result = append(result, types.Battle)
		case compiler.ConditionCardTypeCreature:
			result = append(result, types.Creature)
		case compiler.ConditionCardTypeEnchantment:
			result = append(result, types.Enchantment)
		case compiler.ConditionCardTypeLand:
			result = append(result, types.Land)
		case compiler.ConditionCardTypePlaneswalker:
			result = append(result, types.Planeswalker)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerConditionSupertypes(values []compiler.ConditionSupertype) ([]types.Super, bool) {
	result := make([]types.Super, 0, len(values))
	for _, value := range values {
		switch value {
		case compiler.ConditionSupertypeBasic:
			result = append(result, types.Basic)
		case compiler.ConditionSupertypeSnow:
			result = append(result, types.Snow)
		case compiler.ConditionSupertypeLegendary:
			result = append(result, types.Legendary)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerConditionColors(values []compiler.ConditionColor) ([]color.Color, bool) {
	result := make([]color.Color, 0, len(values))
	for _, value := range values {
		switch value {
		case compiler.ConditionColorWhite:
			result = append(result, color.White)
		case compiler.ConditionColorBlue:
			result = append(result, color.Blue)
		case compiler.ConditionColorBlack:
			result = append(result, color.Black)
		case compiler.ConditionColorRed:
			result = append(result, color.Red)
		case compiler.ConditionColorGreen:
			result = append(result, color.Green)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerEventHistoryWindow(window compiler.ConditionEventHistoryWindow) (game.EventHistoryWindow, bool) {
	switch window {
	case compiler.ConditionEventHistoryWindowCurrentTurn:
		return game.EventHistoryCurrentTurn, true
	case compiler.ConditionEventHistoryWindowPreviousTurn:
		return game.EventHistoryPreviousTurn, true
	default:
		return game.EventHistoryCurrentTurn, false
	}
}
