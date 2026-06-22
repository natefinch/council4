package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerStaticSelection projects a static affected-group filter onto the
// canonical game.Selection. Its shared dimension cluster (card types,
// supertypes, excluded supertypes, subtypes, colors, combat state, tap state,
// controller, keywords, token/non-token) is routed through the single
// SelectionForSelector projector via staticSelectionSelector, exactly like every
// other effect context, instead of a second hand-written projector. The genuine
// static-affected-group extras the canonical CompiledSelector does not model are
// kept as a documented rider applied to the projector result below.
func lowerStaticSelection(selection compiler.StaticSelection) (game.Selection, bool) {
	selector, ok := staticSelectionSelector(selection)
	if !ok {
		return game.Selection{}, false
	}
	result, ok := SelectionForSelectorMasked(selector, SelectionMask{})
	if !ok {
		return game.Selection{}, false
	}

	// Per-context extras kept on the projector result (umbrella #1414):
	// SubtypeFromEntryChoice, ColorFromEntryChoice, MatchCounter/RequiredCounter,
	// MatchAnyCounter, Modified, Commander, the power-or-toughness disjunction,
	// Power/Toughness, PowerLessThanSource, PowerGreaterThanSource. The shared
	// CompiledSelector does not carry these static-affected-group dimensions, so
	// they are applied here rather than added to the canonical selector.
	if selection.SubtypeFromEntryChoice {
		result.SubtypeChoice = game.SubtypeChoiceSourceEntry
	}
	if selection.ColorFromEntryChoice {
		result.ColorChoice = game.ColorChoiceSourceEntry
	}
	if selection.MatchCounter {
		result.MatchCounter = true
		result.RequiredCounter = selection.RequiredCounter
	}
	if selection.MatchAnyCounter {
		result.MatchAnyCounter = true
	}
	if selection.Modified {
		result.MatchModified = true
	}
	if selection.Commander {
		result.MatchCommander = true
	}
	switch {
	case selection.PowerOrToughness:
		result.AnyOf = []game.Selection{
			{Power: opt.Val(selection.Power)},
			{Toughness: opt.Val(selection.Toughness)},
		}
	default:
		if selection.MatchPower {
			result.Power = opt.Val(selection.Power)
		}
		if selection.MatchToughness {
			result.Toughness = opt.Val(selection.Toughness)
		}
	}
	if selection.PowerLessThanSource {
		result.PowerLessThanSource = true
	}
	if selection.PowerGreaterThanSource {
		result.PowerGreaterThanSource = true
	}
	return result, len(result.Validate()) == 0
}

// staticSelectionSelector translates a StaticSelection's shared filter dimensions
// into a compiler.CompiledSelector that SelectionForSelectorMasked projects onto
// the runtime Selection. The clone enums are mapped with the existing per-enum
// helpers (lowerStaticCombatState, lowerStaticTapState, lowerStaticCardType),
// which fail closed on any value they cannot translate. The controller and
// keyword filters share the canonical selector enums, so they pass through
// directly and the projector reproduces their mapping (and the controller
// fail-closed guard the prior hand-written projector enforced). The static
// required card types are a conjunctive "all of" set, so they ride RequiredTypesAny
// with ConjunctiveTypes set, which the projector lowers to Selection.RequiredTypes.
func staticSelectionSelector(selection compiler.StaticSelection) (compiler.CompiledSelector, bool) {
	combatState, ok := lowerStaticCombatState(selection.CombatState)
	if !ok {
		return compiler.CompiledSelector{}, false
	}
	tapState, ok := lowerStaticTapState(selection.TapState)
	if !ok {
		return compiler.CompiledSelector{}, false
	}

	var requiredTypes []types.Card
	for _, cardType := range selection.RequiredTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return compiler.CompiledSelector{}, false
		}
		requiredTypes = append(requiredTypes, value)
	}
	var excludedTypes []types.Card
	for _, cardType := range selection.ExcludedTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return compiler.CompiledSelector{}, false
		}
		excludedTypes = append(excludedTypes, value)
	}

	selector := compiler.CompiledSelector{
		Kind:             compiler.SelectorPermanent,
		Controller:       selection.Controller,
		Colorless:        selection.Colorless,
		Multicolored:     selection.Multicolored,
		TokenOnly:        selection.TokenOnly,
		NonToken:         selection.NonToken,
		Keyword:          selection.Keyword,
		ExcludedKeyword:  selection.ExcludedKeyword,
		ConjunctiveTypes: len(requiredTypes) > 0,
	}

	switch combatState {
	case game.CombatStateAttacking:
		selector.Attacking = true
	case game.CombatStateBlocking:
		selector.Blocking = true
	default:
	}

	switch tapState {
	case game.TriTrue:
		selector.Tapped = true
	case game.TriFalse:
		selector.Untapped = true
	default:
	}

	var excludedSupertypes []types.Super
	if len(selection.ExcludedSupertypes) > 0 {
		excludedSupertypes = []types.Super{selection.ExcludedSupertypes[0]}
	}

	selector = selector.WithAtoms(compiler.CompiledSelectorAtoms{
		RequiredTypesAny:   requiredTypes,
		ExcludedTypes:      excludedTypes,
		Supertypes:         slices.Clone(selection.Supertypes),
		ExcludedSupertypes: excludedSupertypes,
		SubtypesAny:        slices.Clone(selection.SubtypesAny),
		ColorsAny:          slices.Clone(selection.ColorsAny),
	})

	return selector, true
}

func lowerStaticCombatState(state compiler.StaticCombatState) (game.CombatStateFilter, bool) {
	switch state {
	case compiler.StaticCombatStateAny:
		return game.CombatStateAny, true
	case compiler.StaticCombatStateAttacking:
		return game.CombatStateAttacking, true
	case compiler.StaticCombatStateBlocking:
		return game.CombatStateBlocking, true
	default:
		return game.CombatStateAny, false
	}
}

func lowerStaticTapState(state compiler.StaticTapState) (game.TriState, bool) {
	switch state {
	case compiler.StaticTapStateAny:
		return game.TriAny, true
	case compiler.StaticTapStateTapped:
		return game.TriTrue, true
	case compiler.StaticTapStateUntapped:
		return game.TriFalse, true
	default:
		return game.TriAny, false
	}
}

func lowerStaticCardType(cardType compiler.StaticCardType) (types.Card, bool) {
	switch cardType {
	case compiler.StaticCardTypeArtifact:
		return types.Artifact, true
	case compiler.StaticCardTypeCreature:
		return types.Creature, true
	case compiler.StaticCardTypeLand:
		return types.Land, true
	case compiler.StaticCardTypeEnchantment:
		return types.Enchantment, true
	case compiler.StaticCardTypeInstant:
		return types.Instant, true
	case compiler.StaticCardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}
