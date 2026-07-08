package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileTriggerSelection(syntax parser.TriggerSelection) (TriggerSelection, bool) {
	selection := TriggerSelection{
		Colorless:    syntax.Colorless,
		Multicolored: syntax.Multicolored,
		NonToken:     syntax.NonToken,
		TokenOnly:    syntax.TokenOnly,
	}
	var ok bool
	for _, value := range syntax.RequiredTypes {
		selection.RequiredTypes = append(selection.RequiredTypes, compileTriggerCardType(value))
		if selection.RequiredTypes[len(selection.RequiredTypes)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.RequiredTypesAny {
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compileTriggerCardType(value))
		if selection.RequiredTypesAny[len(selection.RequiredTypesAny)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedTypes {
		selection.ExcludedTypes = append(selection.ExcludedTypes, compileTriggerCardType(value))
		if selection.ExcludedTypes[len(selection.ExcludedTypes)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.Supertypes {
		selection.Supertypes = append(selection.Supertypes, compileTriggerSupertype(value))
		if selection.Supertypes[len(selection.Supertypes)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ColorsAny {
		selection.ColorsAny = append(selection.ColorsAny, compileTriggerColor(value))
		if selection.ColorsAny[len(selection.ColorsAny)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedColors {
		selection.ExcludedColors = append(selection.ExcludedColors, compileTriggerColor(value))
		if selection.ExcludedColors[len(selection.ExcludedColors)-1] == "" {
			return TriggerSelection{}, false
		}
	}
	if len(syntax.SubtypesAny) > 0 {
		selection.SubtypesAny = make([]types.Sub, len(syntax.SubtypesAny))
		copy(selection.SubtypesAny, syntax.SubtypesAny)
	}
	selection.Controller, ok = compileTriggerController(syntax.Controller)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Tapped, ok = compileTriggerSelectionTapped(syntax.Tapped)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.CombatState, ok = compileTriggerSelectionCombatState(syntax.CombatState)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Keyword, ok = compileTriggerSelectionKeyword(syntax.Keyword)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.ExcludedKeyword, ok = compileTriggerSelectionKeyword(syntax.ExcludedKeyword)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.ManaValue, ok = compileTriggerSelectionNumber(syntax.ManaValue)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Power, ok = compileTriggerSelectionNumber(syntax.Power)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Toughness, ok = compileTriggerSelectionNumber(syntax.Toughness)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.SubtypeFromEntryChoice = syntax.SubtypeFromEntryChoice
	selection.MatchAnyCounter = syntax.MatchAnyCounter
	selection.MatchCounter = syntax.MatchCounter
	selection.RequiredCounter = syntax.RequiredCounter
	selection.Modified = syntax.Modified
	selection.Commander = syntax.Commander
	for i := range syntax.AnyOf {
		alternative, ok := compileTriggerSelection(syntax.AnyOf[i])
		if !ok {
			return TriggerSelection{}, false
		}
		selection.AnyOf = append(selection.AnyOf, alternative)
	}
	return selection, true
}

func compileTriggerCardType(value parser.TriggerCardType) types.Card {
	switch value {
	case parser.TriggerCardTypeArtifact:
		return types.Artifact
	case parser.TriggerCardTypeBattle:
		return types.Battle
	case parser.TriggerCardTypeCreature:
		return types.Creature
	case parser.TriggerCardTypeEnchantment:
		return types.Enchantment
	case parser.TriggerCardTypeInstant:
		return types.Instant
	case parser.TriggerCardTypeLand:
		return types.Land
	case parser.TriggerCardTypePlaneswalker:
		return types.Planeswalker
	case parser.TriggerCardTypeSorcery:
		return types.Sorcery
	default:
		return ""
	}
}

func compileTriggerSupertype(value parser.TriggerSupertype) types.Super {
	switch value {
	case parser.TriggerSupertypeLegendary:
		return types.Legendary
	case parser.TriggerSupertypeSnow:
		return types.Snow
	default:
		return ""
	}
}

func compileTriggerColor(value parser.TriggerColor) color.Color {
	switch value {
	case parser.TriggerColorWhite:
		return color.White
	case parser.TriggerColorBlue:
		return color.Blue
	case parser.TriggerColorBlack:
		return color.Black
	case parser.TriggerColorRed:
		return color.Red
	case parser.TriggerColorGreen:
		return color.Green
	default:
		return ""
	}
}

func compileTriggerController(value parser.TriggerController) (ControllerKind, bool) {
	switch value {
	case parser.ControllerAny:
		return ControllerAny, true
	case parser.ControllerYou:
		return ControllerYou, true
	case parser.ControllerOpponent:
		return ControllerOpponent, true
	default:
		return ControllerAny, false
	}
}

func compileTriggerSelectionTapped(value parser.TriggerSelectionTappedState) (TriggerTriState, bool) {
	switch value {
	case parser.TriggerSelectionTappedAny:
		return TriggerTriAny, true
	case parser.TriggerSelectionTapped:
		return TriggerTriTrue, true
	case parser.TriggerSelectionUntapped:
		return TriggerTriFalse, true
	default:
		return TriggerTriAny, false
	}
}

func compileTriggerSelectionCombatState(value parser.TriggerSelectionCombatState) (TriggerCombatState, bool) {
	switch value {
	case parser.TriggerSelectionCombatAny:
		return TriggerCombatStateAny, true
	case parser.TriggerSelectionAttacking:
		return TriggerCombatStateAttacking, true
	case parser.TriggerSelectionBlocking:
		return TriggerCombatStateBlocking, true
	default:
		return TriggerCombatStateAny, false
	}
}

func compileTriggerSelectionKeyword(value parser.KeywordKind) (parser.KeywordKind, bool) {
	switch value {
	case parser.KeywordUnknown:
		return parser.KeywordUnknown, true
	case parser.KeywordDeathtouch:
		return parser.KeywordDeathtouch, true
	case parser.KeywordDefender:
		return parser.KeywordDefender, true
	case parser.KeywordFlash:
		return parser.KeywordFlash, true
	case parser.KeywordFlying:
		return parser.KeywordFlying, true
	case parser.KeywordHaste:
		return parser.KeywordHaste, true
	case parser.KeywordShadow:
		return parser.KeywordShadow, true
	default:
		return parser.KeywordUnknown, false
	}
}

func compileTriggerSelectionNumber(value parser.TriggerSelectionNumber) (compare.Int, bool) {
	switch value.Comparison {
	case parser.TriggerSelectionComparisonUnknown:
		return compare.Int{}, value.Value == 0
	case parser.TriggerSelectionComparisonEqual:
		return compare.Int{Op: compare.Equal, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtMost:
		return compare.Int{Op: compare.LessOrEqual, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtLeast:
		return compare.Int{Op: compare.GreaterOrEqual, Value: value.Value}, value.Value >= 0
	default:
		return compare.Int{}, false
	}
}
