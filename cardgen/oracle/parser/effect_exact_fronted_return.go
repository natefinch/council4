package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// exactFrontedHandGraveyardReturnEffectSyntax reports whether effect is a
// graveyard-card return to hand whose destination is fronted ahead of its
// target: "Return to your hand target <noun> in your graveyard[<qualifier>]."
// (Scrap Trawler's "return to your hand target artifact card in your graveyard
// with lesser mana value"). The canonical target-trailing form ("Return target
// <noun> from your graveyard to your hand.") is reconstructed by
// exactGraveyardReturnEffectSyntax; this fronted form spells its graveyard source
// inline with an "in your graveyard" phrase and places the destination first, so
// it needs its own byte-exact reconstruction. It accepts only a single controller
// ("your") graveyard target with one of the shared graveyard-card noun filters
// and an optional mana-value, power, or toughness qualifier, and fails closed for
// every other shape so an unrepresentable fronted return keeps failing rather
// than lowering to a wrong predicate.
func exactFrontedHandGraveyardReturnEffectSyntax(effect *EffectSyntax) bool {
	if effect.ToZone != zone.Hand || effect.FromZone != zone.Graveyard {
		return false
	}
	if len(effect.Targets) != 1 || effect.Amount.VariableX {
		return false
	}
	targetText, ok := exactFrontedGraveyardCardTargetText(&effect.Targets[0])
	if !ok {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Return to your hand "+targetText+".")
}

// exactFrontedGraveyardCardTargetText reconstructs the canonical target noun
// phrase of a fronted-destination graveyard-card return ("target artifact card in
// your graveyard with lesser mana value") from the target's typed selection and
// reports whether the reconstruction is representable. It accepts only a single
// "your" graveyard target ("target <noun> in your graveyard") using the shared
// graveyard-card noun and numeric-qualifier reconstructors, so it recognizes the
// same card-type, color, subtype, supertype, permanent, and plain "card" nouns
// with an optional mana-value, power, or toughness qualifier that the
// target-trailing return does. It fails closed for a self-exclusion, any
// multi-target or optional cardinality, a non-"your" graveyard owner, and every
// combat, tap, keyword, or source-type rider those shared reconstructors do not
// render, so an unsupported shape never produces a wrong noun phrase.
func exactFrontedGraveyardCardTargetText(target *TargetSyntax) (string, bool) {
	sel := target.Selection
	if sel.Zone != zone.Graveyard || sel.Controller != SelectionControllerYou {
		return "", false
	}
	if target.Cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		return "", false
	}
	if sel.Another || sel.Other || sel.All || sel.Attacking || sel.Blocking ||
		sel.Tapped || sel.Untapped || sel.SingleGraveyard ||
		sel.Keyword != KeywordUnknown ||
		len(sel.SourceTypes) != 0 || len(sel.ExcludedColors) != 0 {
		return "", false
	}
	noun, ok := graveyardCardNoun(sel, false)
	if !ok {
		return "", false
	}
	manaClause, ok := graveyardNumericQualifier(sel)
	if !ok {
		return "", false
	}
	return "target " + noun + " in your graveyard" + manaClause, true
}
