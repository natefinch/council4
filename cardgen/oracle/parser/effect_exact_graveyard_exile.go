package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

func exactChosenGraveyardExileEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController || effect.FromZone != zone.Graveyard ||
		len(effect.Targets) != 0 || len(effect.References) != 0 || !effect.Amount.Known ||
		effect.Amount.Value < 1 || effect.Amount.DynamicKind != EffectDynamicAmountNone {
		return false
	}
	selection := effect.Selection
	if selection.Zone != zone.Graveyard || selection.Controller != SelectionControllerYou ||
		selection.All || selection.Another || selection.Other || selection.Attacking ||
		selection.Blocking || selection.Tapped || selection.Untapped {
		return false
	}
	noun, ok := graveyardCardNoun(selection, false)
	if !ok {
		return false
	}
	if effect.Amount.Value > 1 {
		noun += "s"
	}
	qualifier, ok := graveyardNumericQualifier(selection)
	if !ok {
		return false
	}
	count := effectAmountSourceText(effect)
	expected := "exile " + count + " " + noun + qualifier + " from your graveyard."
	return strings.EqualFold(exactEffectClauseText(effect), expected) ||
		strings.EqualFold(exactEffectClauseText(effect), "you "+expected)
}
