package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
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
		if selection.RequiredTypes[len(selection.RequiredTypes)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.RequiredTypesAny {
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compileTriggerCardType(value))
		if selection.RequiredTypesAny[len(selection.RequiredTypesAny)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedTypes {
		selection.ExcludedTypes = append(selection.ExcludedTypes, compileTriggerCardType(value))
		if selection.ExcludedTypes[len(selection.ExcludedTypes)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.Supertypes {
		selection.Supertypes = append(selection.Supertypes, compileTriggerSupertype(value))
		if selection.Supertypes[len(selection.Supertypes)-1] == TriggerSupertypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ColorsAny {
		selection.ColorsAny = append(selection.ColorsAny, compileTriggerColor(value))
		if selection.ColorsAny[len(selection.ColorsAny)-1] == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedColors {
		selection.ExcludedColors = append(selection.ExcludedColors, compileTriggerColor(value))
		if selection.ExcludedColors[len(selection.ExcludedColors)-1] == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
	}
	if len(syntax.SubtypesAny) > 0 {
		selection.SubtypesAny = make([]TriggerSubtype, len(syntax.SubtypesAny))
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
	return selection, ok
}

func compileTriggerCardType(value parser.TriggerCardType) TriggerCardType {
	switch value {
	case parser.TriggerCardTypeArtifact:
		return TriggerCardTypeArtifact
	case parser.TriggerCardTypeBattle:
		return TriggerCardTypeBattle
	case parser.TriggerCardTypeCreature:
		return TriggerCardTypeCreature
	case parser.TriggerCardTypeEnchantment:
		return TriggerCardTypeEnchantment
	case parser.TriggerCardTypeInstant:
		return TriggerCardTypeInstant
	case parser.TriggerCardTypeLand:
		return TriggerCardTypeLand
	case parser.TriggerCardTypePlaneswalker:
		return TriggerCardTypePlaneswalker
	case parser.TriggerCardTypeSorcery:
		return TriggerCardTypeSorcery
	default:
		return TriggerCardTypeUnknown
	}
}

func compileTriggerSupertype(value parser.TriggerSupertype) TriggerSupertype {
	switch value {
	case parser.TriggerSupertypeLegendary:
		return TriggerSupertypeLegendary
	case parser.TriggerSupertypeSnow:
		return TriggerSupertypeSnow
	default:
		return TriggerSupertypeUnknown
	}
}

func compileTriggerColor(value parser.TriggerColor) TriggerColor {
	switch value {
	case parser.TriggerColorWhite:
		return TriggerColorWhite
	case parser.TriggerColorBlue:
		return TriggerColorBlue
	case parser.TriggerColorBlack:
		return TriggerColorBlack
	case parser.TriggerColorRed:
		return TriggerColorRed
	case parser.TriggerColorGreen:
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
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

func compileTriggerSelectionKeyword(value parser.KeywordKind) (TriggerKeyword, bool) {
	switch value {
	case parser.KeywordUnknown:
		return TriggerKeywordUnknown, true
	case parser.KeywordDefender:
		return TriggerKeywordDefender, true
	case parser.KeywordFlash:
		return TriggerKeywordFlash, true
	case parser.KeywordFlying:
		return TriggerKeywordFlying, true
	case parser.KeywordHaste:
		return TriggerKeywordHaste, true
	case parser.KeywordShadow:
		return TriggerKeywordShadow, true
	default:
		return TriggerKeywordUnknown, false
	}
}

func compileTriggerSelectionNumber(value parser.TriggerSelectionNumber) (TriggerNumberFilter, bool) {
	switch value.Comparison {
	case parser.TriggerSelectionComparisonUnknown:
		return TriggerNumberFilter{}, value.Value == 0
	case parser.TriggerSelectionComparisonEqual:
		return TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtMost:
		return TriggerNumberFilter{Comparison: TriggerComparisonAtMost, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtLeast:
		return TriggerNumberFilter{Comparison: TriggerComparisonAtLeast, Value: value.Value}, value.Value >= 0
	default:
		return TriggerNumberFilter{}, false
	}
}
