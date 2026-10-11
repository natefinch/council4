package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func conditionObjectCardTypes(values []types.Card) ([]types.Card, bool) {
	result := make([]types.Card, 0, len(values))
	for _, value := range values {
		switch value {
		case types.Artifact, types.Battle, types.Creature, types.Enchantment,
			types.Instant, types.Land, types.Planeswalker, types.Sorcery:
			result = append(result, value)
		default:
			return nil, false
		}
	}
	return result, true
}

// Expanded selections use the same exact reference/domain check as their gate.
func conditionTypeSelectionBindingsSupported(content compiler.AbilityContent) bool {
	if _, ok := lowerShuffleRevealPermanentSequence(contentCtx{content: content}); ok {
		return true
	}
	for _, condition := range content.Conditions {
		if condition.Predicate != compiler.ConditionPredicateObjectMatches ||
			!expandedConditionTypeSelection(condition.Selection) {
			continue
		}
		if condition.ObjectBinding != compiler.ReferenceBindingTarget && condition.ObjectBinding != compiler.ReferenceBindingEventPermanent {
			continue
		}
		if condition.ObjectBinding == compiler.ReferenceBindingTarget && !condition.HasSubjectReference {
			return false
		}
		if _, ok := lowerObjectMatchReference(condition, conditionContextEffectGate); !ok {
			return false
		}
	}
	return true
}

func expandedConditionTypeSelection(selection compiler.ConditionSelection) bool {
	return len(selection.RequiredTypesAny) > 0 || len(selection.ExcludedTypes) > 0 ||
		len(selection.AnyOf) > 0 || slices.Contains(selection.RequiredTypes, types.Instant) ||
		slices.Contains(selection.RequiredTypes, types.Sorcery)
}

// The linked-reveal adapter already binds and checks its card independently of
// the generic condition reference. Recognition must not disable that adapter.
func shuffleRevealPermanentCondition(condition compiler.CompiledCondition) bool {
	if condition.Predicate != compiler.ConditionPredicateObjectMatches || condition.Negated {
		return false
	}
	selection, ok := lowerConditionSelection(condition.Selection)
	if !ok || !slices.Equal(selection.RequiredTypesAny,
		[]types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}) {
		return false
	}
	selection.RequiredTypesAny = nil
	return selection.Empty()
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
	result.RequiredTypesAny, ok = conditionObjectCardTypes(selection.RequiredTypesAny)
	if !ok {
		return game.Selection{}, false
	}
	for _, alternative := range selection.AnyOf {
		lowered, ok := lowerConditionSelection(alternative)
		if !ok {
			return game.Selection{}, false
		}
		result.AnyOf = append(result.AnyOf, lowered)
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
	result.NonToken = selection.NonToken
	switch selection.Attachment {
	case compiler.ConditionAttachmentEnchanted:
		result.MatchEnchanted = true
	case compiler.ConditionAttachmentEquipped:
		result.MatchEquipped = true
	default:
	}
	if selection.CounterKindKnown {
		result.RequiredCounter = selection.CounterKind
		if selection.CounterCountLessThan > 0 {
			result.RequiredCounterCount = opt.Val(compare.Int{Op: compare.LessThan, Value: selection.CounterCountLessThan})
		} else {
			result.RequiredCounterCount = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.CounterCountAtLeast})
		}
	}
	if selection.MatchPowerAtLeast {
		result.Power = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: selection.PowerAtLeast})
	} else if selection.PowerAtLeast != 0 {
		return game.Selection{}, false
	}
	switch selection.AttributeCompare.Attribute {
	case compiler.ConditionAttributeNone:
	case compiler.ConditionAttributeManaValue:
		result.ManaValue = opt.Val(compare.Int{Op: selection.AttributeCompare.Op, Value: selection.AttributeCompare.Value})
	case compiler.ConditionAttributePower:
		// PowerAtLeast/MatchPowerAtLeast (the narrower "with power N or greater"
		// trailing selection qualifier) and AttributeCompare are populated by
		// disjoint recognizers and should never both name Power; fail closed
		// rather than silently letting one bound overwrite the other.
		if result.Power.Exists {
			return game.Selection{}, false
		}
		result.Power = opt.Val(compare.Int{Op: selection.AttributeCompare.Op, Value: selection.AttributeCompare.Value})
	case compiler.ConditionAttributeToughness:
		result.Toughness = opt.Val(compare.Int{Op: selection.AttributeCompare.Op, Value: selection.AttributeCompare.Value})
	default:
		return game.Selection{}, false
	}
	return result, len(result.Validate()) == 0
}

// conditionSelectionSelector translates a ConditionSelection's filter
// dimensions into a compiler.CompiledSelector. Its shared-typed required-type,
// supertype, and color fields are consumed directly, failing closed on any
// value outside the object-selection vocabulary. The condition extras
// (counters, ExcludeSource, the power-at-least bound, and TokenOnly) are applied
// by lowerConditionSelection directly because they do not round-trip through
// CompiledSelector.
func conditionSelectionSelector(selection compiler.ConditionSelection) (compiler.CompiledSelector, bool) {
	required, ok := conditionObjectCardTypes(selection.RequiredTypes)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	excluded, ok := conditionObjectCardTypes(selection.ExcludedTypes)
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
		ExcludedTypes:    excluded,
		Supertypes:       supertypes,
		SubtypesAny:      subtypes,
		ColorsAny:        colors,
	})

	return selector, true
}
