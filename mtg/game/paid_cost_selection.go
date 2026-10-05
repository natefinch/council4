package game

import (
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// PaidCostSelectionSupported limits predicates to captured characteristics
// projected by the shared Selection matcher, not later battlefield state.
func PaidCostSelectionSupported(selection Selection) bool {
	if len(selection.Validate()) != 0 {
		return false
	}
	for _, alternative := range selection.AnyOf {
		if !PaidCostSelectionSupported(alternative) {
			return false
		}
	}
	selection.AnyOf = nil
	selection.RequiredTypes = nil
	selection.RequiredTypesAny = nil
	selection.ExcludedTypes = nil
	selection.Supertypes = nil
	selection.ExcludedSupertype = ""
	selection.SubtypesAny = nil
	selection.ExcludedSubtype = ""
	selection.ColorsAny = nil
	selection.ExcludedColors = nil
	selection.Colorless = false
	selection.Multicolored = false
	selection.Colored = false
	selection.ManaValue = opt.V[compare.Int]{}
	selection.Power = opt.V[compare.Int]{}
	selection.Toughness = opt.V[compare.Int]{}
	return selection.Empty()
}
