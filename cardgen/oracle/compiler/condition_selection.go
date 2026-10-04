package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileConditionPermanentSelection(syntax parser.ConditionSelection) (ConditionSelection, bool) {
	selection, ok := compileConditionSelection(syntax)
	if !ok || !conditionSelectionPermanentTypes(selection) {
		return ConditionSelection{}, false
	}
	return selection, true
}

func conditionSelectionPermanentTypes(selection ConditionSelection) bool {
	nonpermanent := func(cardType types.Card) bool {
		return cardType == types.Instant || cardType == types.Sorcery
	}
	if slices.ContainsFunc(selection.RequiredTypes, nonpermanent) ||
		slices.ContainsFunc(selection.RequiredTypesAny, nonpermanent) {
		return false
	}
	for _, alternative := range selection.AnyOf {
		if !conditionSelectionPermanentTypes(alternative) {
			return false
		}
	}
	return true
}
