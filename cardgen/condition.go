package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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
	// conditionContextStaticRuleGuard gates the trailing guard clause on a
	// land-gated combat restriction ("... can't attack or block unless you
	// control seven or more lands."). Unlike a general static-ability condition,
	// the guard permits the "unless" form, which the runtime evaluates via the
	// condition's Negate flag.
	conditionContextStaticRuleGuard
	// conditionContextSpellCostReduction gates the "if ..." condition of a
	// source-spell flat cast cost reduction ("This spell costs {N} less to cast
	// if you control a Wizard."). The runtime evaluates the condition against the
	// caster's board and player state as the spell is cast, with no resolving
	// stack object or source permanent available, so only controller- and
	// board-scoped predicates are permitted.
	conditionContextSpellCostReduction
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
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerLifeAtMost:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerLifeAtLeastAboveStarting:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerLifeAboveStarting, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerHandSizeAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerHandSize, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerHandSizeExactly:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerHandSize, Op: compare.Equal, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerLibrarySizeAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerLibrarySize, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerLifeExactly:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.Equal, Value: condition.Threshold})
	case compiler.ConditionPredicateSpellXAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateEventSpellManaSpentToCastAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateEventSpellManaSpentToCast, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateEventSpellManaSpentToCastAtMost:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateEventSpellManaSpentToCast, Op: compare.LessOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateTriggeringPlayerHandSizeAtMost:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.LessOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateTriggeringPlayerHandSizeAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateAnyOpponentPoisonAtLeast:
		result.AnyOpponentPoisonAtLeast = condition.Threshold
	case compiler.ConditionPredicateAnyPlayerLifeAtMost:
		result.AnyPlayerLifeAtMost = condition.Threshold
	case compiler.ConditionPredicateOpponentCountAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
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
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerGraveyardCardTypeCountAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerGraveyardCardTypeCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerGraveyardPermanentCardCountAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerGraveyardPermanentCardCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerGraveyardManaValueCountAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerGraveyardManaValueCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateAnyOpponentGraveyardCardCountAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateAnyOpponentGraveyardCardCount, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast:
		if condition.GraveyardCountCardType == "" {
			return game.Condition{}, false
		}
		result.ControllerGraveyardCardOfTypeCountAtLeast = condition.Threshold
		result.ControllerGraveyardCountCardType = condition.GraveyardCountCardType
	case compiler.ConditionPredicateControllerCreaturePowerDiversityAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerCreaturePowerDiversity, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateAttackersAttackingControllerAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateAttackersAttackingController, Op: compare.GreaterOrEqual, Value: condition.Threshold})
	case compiler.ConditionPredicateControllerGainedLifeThisTurnAtLeast:
		result.Aggregates = append(result.Aggregates, game.AggregateComparison{Aggregate: game.AggregateControllerGainedLifeThisTurn, Op: compare.GreaterOrEqual, Value: condition.Threshold})
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
	case compiler.ConditionPredicateCastDuringControllerMainPhase:
		result.CastDuringControllerMainPhase = true
	case compiler.ConditionPredicateSpellWasKicked:
		result.SpellWasKicked = true
	case compiler.ConditionPredicateEventSubjectWasKicked:
		result.EventPermanentWasKicked = true
	case compiler.ConditionPredicateColoredManaSpentToCastAtLeast:
		if condition.ManaSpentColor == "" || condition.Threshold <= 0 {
			return game.Condition{}, false
		}
		result.SpellColorManaSpent = game.ColorManaSpendThreshold{
			Color: condition.ManaSpentColor,
			Count: condition.Threshold,
		}
	case compiler.ConditionPredicateSameColorManaSpentToCastAtLeast:
		if condition.Threshold <= 0 {
			return game.Condition{}, false
		}
		result.SpellSameColorManaSpentAtLeast = condition.Threshold
	case compiler.ConditionPredicateSpellWasCastFromGraveyard:
		result.CastFromZone = opt.Val(zone.Graveyard)
	case compiler.ConditionPredicateSourceSaddled:
		result.SourceSaddled = true
	case compiler.ConditionPredicateSourceNotSaddled:
		result.SourceSaddled = true
		result.Negate = !result.Negate
	case compiler.ConditionPredicateSourceTributeNotPaid:
		result.SourceTributeNotPaid = true
	case compiler.ConditionPredicateControllerControlsCommander:
		result.ControllerControlsCommander = true
	case compiler.ConditionPredicateLandEnteredThisTurnOrControlsBasic:
		result.LandEnteredThisTurnOrControlsBasicLand = true
	case compiler.ConditionPredicateControllerControlsNamed:
		if len(condition.ControlledNames) == 0 {
			return game.Condition{}, false
		}
		result.ControllerControlsNamed = append(result.ControllerControlsNamed, condition.ControlledNames...)
	case compiler.ConditionPredicateFirstCombatPhaseOfTurn:
		result.FirstCombatPhaseOfTurn = true
	case compiler.ConditionPredicateControllerTurn:
		result.SourceControllerTurn = true
	case compiler.ConditionPredicateControlsGreatestPowerCreature:
		result.ControllerControlsGreatestPowerCreature = true
	case compiler.ConditionPredicateControlsGreatestToughnessCreature:
		result.ControllerControlsGreatestToughnessCreature = true
	case compiler.ConditionPredicateControllerIsMonarch:
		result.ControllerIsMonarch = true
	case compiler.ConditionPredicateControllerHasInitiative:
		result.ControllerHasInitiative = true
	case compiler.ConditionPredicateControllerHasCityBlessing:
		result.ControllerHasCityBlessing = true
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
			Pattern:  pattern,
			Window:   window,
			MinCount: condition.EventHistoryMinCount,
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
	case conditionContextStaticRuleGuard:
		return (condition.Kind == compiler.ConditionAsLongAs ||
			condition.Kind == compiler.ConditionUnless) && !condition.Intervening
	case conditionContextActivation:
		return condition.Kind == compiler.ConditionOnlyIf && !condition.Intervening
	case conditionContextInterveningTrigger:
		return condition.Kind == compiler.ConditionIf && condition.Intervening
	case conditionContextReplacement:
		// The "unless" form ("This land enters tapped unless you control a
		// Plains.") gates the replacement on the negation of its condition; the
		// "if" form ("If you control two or more other lands, this land enters
		// tapped.") gates it on the condition holding. The runtime evaluates
		// both through the same Condition (the compiler records the "unless"
		// negation in Negated), so accept either non-intervening shape.
		return (condition.Kind == compiler.ConditionUnless ||
			condition.Kind == compiler.ConditionIf) && !condition.Intervening
	case conditionContextEntryCounters, conditionContextEffectGate, conditionContextSpellCostReduction:
		return condition.Kind == compiler.ConditionIf && !condition.Intervening
	default:
		return false
	}
}

func conditionPredicateAllowedInContext(predicate compiler.ConditionPredicate, ctx conditionLoweringContext) bool {
	if ctx == conditionContextSpellCostReduction {
		// A source-spell cost reduction is evaluated as the spell is cast, with
		// only the caster known: no resolving stack object, source permanent, or
		// triggering event is available. Permit only board- and player-state
		// predicates that resolve from the controller alone.
		switch predicate {
		case compiler.ConditionPredicateControllerControls,
			compiler.ConditionPredicateAnyOpponentControls,
			compiler.ConditionPredicateOpponentsControl,
			compiler.ConditionPredicateControlComparison,
			compiler.ConditionPredicateControllerLifeAtLeast,
			compiler.ConditionPredicateControllerLifeAtMost,
			compiler.ConditionPredicateControllerLifeExactly,
			compiler.ConditionPredicateControllerLifeAtLeastAboveStarting,
			compiler.ConditionPredicateAnyPlayerLifeAtMost,
			compiler.ConditionPredicateOpponentCountAtLeast,
			compiler.ConditionPredicateControllerHandSizeAtLeast,
			compiler.ConditionPredicateControllerHandSizeExactly,
			compiler.ConditionPredicateControllerHandEmpty,
			compiler.ConditionPredicateControllerLibrarySizeAtLeast,
			compiler.ConditionPredicateControllerGraveyardCardCountAtLeast,
			compiler.ConditionPredicateControllerGraveyardCardTypeCountAtLeast,
			compiler.ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast,
			compiler.ConditionPredicateControllerCreaturePowerDiversityAtLeast,
			compiler.ConditionPredicateAnyOpponentPoisonAtLeast:
			return true
		default:
			return false
		}
	}
	if ctx == conditionContextEntryCounters {
		switch predicate {
		case compiler.ConditionPredicateControllerLifeAtLeast,
			compiler.ConditionPredicateControllerLifeAtMost,
			compiler.ConditionPredicateControllerLifeAtLeastAboveStarting,
			compiler.ConditionPredicateAnyPlayerLifeAtMost,
			compiler.ConditionPredicateOpponentCountAtLeast,
			compiler.ConditionPredicateControllerControls,
			compiler.ConditionPredicateAnyOpponentControls,
			compiler.ConditionPredicateOpponentsControl,
			compiler.ConditionPredicateColoredManaSpentToCastAtLeast,
			compiler.ConditionPredicateSameColorManaSpentToCastAtLeast,
			compiler.ConditionPredicateEventSubjectWasKicked,
			compiler.ConditionPredicateEventHistory:
			return true
		default:
			return false
		}
	}
	if ctx != conditionContextReplacement {
		switch predicate {
		case compiler.ConditionPredicateControllerLifeAtLeast,
			compiler.ConditionPredicateControllerLifeAtMost,
			compiler.ConditionPredicateControllerLifeAtLeastAboveStarting,
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
			compiler.ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast,
			compiler.ConditionPredicateControllerGraveyardPermanentCardCountAtLeast,
			compiler.ConditionPredicateControllerGraveyardManaValueCountAtLeast,
			compiler.ConditionPredicateAnyOpponentGraveyardCardCountAtLeast,
			compiler.ConditionPredicateControllerCreaturePowerDiversityAtLeast,
			compiler.ConditionPredicateAnyOpponentPoisonAtLeast,
			compiler.ConditionPredicateControllerLibrarySizeAtLeast,
			compiler.ConditionPredicateControllerLifeExactly,
			compiler.ConditionPredicateControllerGainedLifeThisTurnAtLeast,
			compiler.ConditionPredicateObjectMatches,
			compiler.ConditionPredicateObjectExists:
			return true
		case compiler.ConditionPredicateEventHistory:
			return ctx == conditionContextInterveningTrigger ||
				ctx == conditionContextActivation ||
				ctx == conditionContextEffectGate
		case compiler.ConditionPredicateCastDuringControllerMainPhase,
			compiler.ConditionPredicateSpellWasKicked,
			compiler.ConditionPredicateSpellWasCastFromGraveyard,
			compiler.ConditionPredicateSpellXAtLeast,
			compiler.ConditionPredicateColoredManaSpentToCastAtLeast,
			compiler.ConditionPredicateSameColorManaSpentToCastAtLeast,
			compiler.ConditionPredicateSourceSaddled,
			compiler.ConditionPredicateSourceNotSaddled,
			compiler.ConditionPredicateControlsGreatestPowerCreature,
			compiler.ConditionPredicateControlsGreatestToughnessCreature,
			compiler.ConditionPredicateControllerControlsNamed:
			return ctx == conditionContextEffectGate
		case compiler.ConditionPredicateFirstCombatPhaseOfTurn:
			return ctx == conditionContextEffectGate ||
				ctx == conditionContextInterveningTrigger
		case compiler.ConditionPredicateControllerTurn:
			return ctx == conditionContextStatic ||
				ctx == conditionContextStaticRuleGuard
		case compiler.ConditionPredicateEventSubjectNameUnique,
			compiler.ConditionPredicateSourceTributeNotPaid,
			compiler.ConditionPredicateControllerIsMonarch,
			compiler.ConditionPredicateControllerHasInitiative,
			compiler.ConditionPredicateControllerHasCityBlessing,
			compiler.ConditionPredicateEventSpellManaSpentToCastAtLeast,
			compiler.ConditionPredicateEventSpellManaSpentToCastAtMost,
			compiler.ConditionPredicateTriggeringPlayerHandSizeAtMost,
			compiler.ConditionPredicateTriggeringPlayerHandSizeAtLeast,
			compiler.ConditionPredicateAttackersAttackingControllerAtLeast:
			return ctx == conditionContextInterveningTrigger
		case compiler.ConditionPredicateControllerControlsCommander:
			return ctx == conditionContextInterveningTrigger || ctx == conditionContextStatic
		case compiler.ConditionPredicateLandEnteredThisTurnOrControlsBasic:
			return ctx == conditionContextActivation
		case compiler.ConditionPredicateControllerHandSizeExactly:
			return ctx == conditionContextStatic || ctx == conditionContextActivation ||
				ctx == conditionContextInterveningTrigger
		default:
			return ctx == conditionContextStatic &&
				predicate == compiler.ConditionPredicateControllerHandSizeAtLeast
		}
	}
	switch predicate {
	case compiler.ConditionPredicateControllerLifeAtLeast,
		compiler.ConditionPredicateControllerLifeAtMost,
		compiler.ConditionPredicateControllerLifeAtLeastAboveStarting,
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
	if condition.Selection.MatchDistinctNamesAtLeast {
		result.DistinctNames = opt.Val(compare.Int{
			Op:    compare.GreaterOrEqual,
			Value: condition.Selection.DistinctNamesAtLeast,
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
	case compiler.ConditionComparisonScopeTriggeringPlayer:
		return game.ControlPlayerTriggeringPlayer, true
	default:
		return game.ControlPlayerController, false
	}
}

// lowerConditionSelection projects a condition-filter onto the canonical
// game.Selection. It is a thin adapter over the shared SelectionForSelector
// projector: conditionSelectionSelector translates the parallel
// ConditionSelection clone enums into a compiler.CompiledSelector,
// SelectionForSelectorMasked maps that shared dimension cluster
// (types/supertypes/subtypes/colors/colorless/multicolored/tapped/combat/keyword)
// onto the runtime Selection, and the genuine condition-specific extras are
// applied as a documented rider afterward. Routing the shared cluster through
// the canonical projector keeps condition filters in lockstep with every other
// selector context instead of maintaining a second hand-written projector.
func lowerConditionSelection(selection compiler.ConditionSelection) (game.Selection, bool) {
	selector, ok := conditionSelectionSelector(selection)
	if !ok {
		return game.Selection{}, false
	}
	result, ok := SelectionForSelectorMasked(selector, SelectionMask{}.Rejecting(DimRequiredName))
	if !ok {
		return game.Selection{}, false
	}
	// Per-context extras kept on the projector result (umbrella #1414):
	// AnyCounter (MatchAnyCounter), the named-counter count threshold
	// (RequiredCounter + RequiredCounterCount), ExcludeSource, the power-at-least
	// bound (Power), and TokenOnly. None of these round-trip byte-identically
	// through CompiledSelector here (the counter-count threshold has no selector
	// field at all), so they ride directly on the shared-core result.
	result.MatchAnyCounter = selection.AnyCounter
	result.ExcludeSource = selection.ExcludeSource
	result.TokenOnly = selection.TokenOnly
	switch selection.Attachment {
	case compiler.ConditionAttachmentEnchanted:
		result.MatchEnchanted = true
	case compiler.ConditionAttachmentEquipped:
		result.MatchEquipped = true
	default:
	}
	if selection.CounterKindKnown {
		result.RequiredCounter = selection.CounterKind
		result.RequiredCounterCount = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.CounterCountAtLeast})
	}
	if selection.MatchPowerAtLeast {
		result.Power = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.PowerAtLeast})
	} else if selection.PowerAtLeast != 0 {
		return game.Selection{}, false
	}
	return result, len(result.Validate()) == 0
}

// conditionSelectionSelector translates a ConditionSelection's filter
// dimensions into a compiler.CompiledSelector. Its shared-typed required-type,
// supertype, and color fields are consumed directly, failing closed on any
// value outside the permanent-selection vocabulary. The condition extras
// (counters, ExcludeSource, the power-at-least bound, and TokenOnly) are applied
// by lowerConditionSelection directly because they do not round-trip through
// CompiledSelector.
func conditionSelectionSelector(selection compiler.ConditionSelection) (compiler.CompiledSelector, bool) {
	required, ok := conditionCardTypes(selection.RequiredTypes)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	supertypes, ok := conditionSupertypes(selection.Supertypes)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	colors, ok := conditionColors(selection.ColorsAny)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	tapped, ok := lowerConditionTriState(selection.Tapped)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	combatState, ok := lowerConditionCombatState(selection.CombatState)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	subtypes := make([]types.Sub, 0, len(selection.SubtypesAny))
	for _, subtype := range selection.SubtypesAny {
		if subtype == "" {
			return compiler.CompiledSelector{}, false
		}
		subtypes = append(subtypes, types.Sub(subtype))
	}

	selector := compiler.CompiledSelector{
		Kind:         compiler.SelectorPermanent,
		Colorless:    selection.Colorless,
		Multicolored: selection.Multicolored,
		// A condition's required-type nouns are conjunctive (the matched
		// permanent must carry every named type at once), matching the legacy
		// projector that put them in Selection.RequiredTypes. ConjunctiveTypes
		// makes SelectionForSelectorMasked fold RequiredTypesAny into the
		// conjunctive RequiredTypes field.
		ConjunctiveTypes: true,
		Keyword:          selection.Keyword,
	}

	switch tapped {
	case game.TriTrue:
		selector.Tapped = true
	case game.TriFalse:
		selector.Untapped = true
	default:
	}

	switch combatState {
	case game.CombatStateAttacking:
		selector.Attacking = true
	case game.CombatStateBlocking:
		selector.Blocking = true
	case game.CombatStateAttackingOrBlocking:
		selector.Attacking = true
		selector.Blocking = true
	default:
	}

	selector = selector.WithAtoms(compiler.CompiledSelectorAtoms{
		RequiredTypesAny: required,
		Supertypes:       supertypes,
		SubtypesAny:      subtypes,
		ColorsAny:        colors,
	})

	return selector, true
}

func lowerConditionObjectReference(binding compiler.ReferenceBinding) (game.ObjectReference, bool) {
	if binding == compiler.ReferenceBindingCreatedToken {
		return game.LinkedObjectReference(createdTokenLinkKey), true
	}
	return lowerObjectReference(compiler.CompiledReference{Binding: binding}, referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
		AllowTarget: true,
	})
}

// targetColorGateSelection returns the color filter of a single resolving
// "if it's <color>" target rider (Pyroblast, Red Elemental Blast). It reports
// false unless the conditions are exactly one ConditionPredicateTargetColor
// clause whose color selection lowers to at least one color. The caller binds
// the filter to the effect's own target object reference.
func targetColorGateSelection(conditions []compiler.CompiledCondition) (game.Selection, bool) {
	if len(conditions) != 1 || conditions[0].Predicate != compiler.ConditionPredicateTargetColor {
		return game.Selection{}, false
	}
	selection, ok := lowerConditionSelection(conditions[0].Selection)
	if !ok || len(selection.ColorsAny) == 0 {
		return game.Selection{}, false
	}
	return selection, true
}

// targetColorEffectCondition builds an instruction gate that resolves the effect
// only if the target named by ref currently has one of the colors in selection.
func targetColorEffectCondition(ref game.ObjectReference, selection game.Selection, text string) game.EffectCondition {
	return game.EffectCondition{
		Text:   text,
		Object: ref,
		Condition: opt.Val(game.Condition{
			Text:          text,
			Object:        opt.Val(ref),
			ObjectMatches: opt.Val(selection),
		}),
	}
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

// conditionCardTypes validates a condition selection's required-type filter,
// failing closed on any value outside the permanent card types a condition may
// select (CR 300.1) or an unset entry.
func conditionCardTypes(values []types.Card) ([]types.Card, bool) {
	result := make([]types.Card, 0, len(values))
	for _, value := range values {
		switch value {
		case types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker:
			result = append(result, value)
		default:
			return nil, false
		}
	}
	return result, true
}

// conditionSupertypes validates a condition selection's supertype filter,
// failing closed on any value outside the supertypes a condition may select or
// an unset entry.
func conditionSupertypes(values []types.Super) ([]types.Super, bool) {
	result := make([]types.Super, 0, len(values))
	for _, value := range values {
		switch value {
		case types.Basic, types.Snow, types.Legendary:
			result = append(result, value)
		default:
			return nil, false
		}
	}
	return result, true
}

// conditionColors validates a condition selection's color filter, failing closed
// on any value outside Magic's five colors or an unset entry.
func conditionColors(values []color.Color) ([]color.Color, bool) {
	result := make([]color.Color, 0, len(values))
	for _, value := range values {
		switch value {
		case color.White, color.Blue, color.Black, color.Red, color.Green:
			result = append(result, value)
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
