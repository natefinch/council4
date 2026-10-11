package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// Stack selectors reconstruct only whole-predicate qualifiers. A color or
// excluded type before a type union can qualify just one branch; flattened
// selection atoms cannot faithfully express that reading.
func exactStackSpellTargetSyntax(text string, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	if selection.Kind != SelectionSpell ||
		selection.All || selection.Another || selection.Other || selection.OtherThanSource ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.NonToken || selection.TokenOnly || selection.Colored || selection.Historic ||
		selection.BasicLandType || selection.ConjunctiveTypes || selection.PlayerOrPlaneswalker ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.MatchTotalManaValue || selection.ManaValueX ||
		selection.ManaValueDynamic != EffectDynamicAmountNone || selection.ManaValueDynamicCount != nil ||
		selection.ManaValueSacrificedCostAddend != nil ||
		selection.ManaValueLessThanEventPermanent || selection.ManaValueLessOrEqualEventPermanent ||
		selectionHasCounterQualifier(selection) || selectionHasAttachmentQualifier(selection) ||
		selection.SubtypeFromEntryChoice || selection.SubtypeFromChosenType || selection.SubtypeFromChosenTypeExcluded ||
		selection.RequiredName != "" || selection.EnteredThisTurn || selection.DealtDamageThisTurn ||
		selection.NameUniqueAmongControlled || selection.SingleGraveyard || selection.SameNameGroup != nil ||
		len(selection.Supertypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.SubtypesAny) != 0 || len(selection.ExcludedSubtypes) != 0 ||
		len(selection.SourceTypes) != 0 || len(selection.Alternatives) != 0 {
		return false
	}
	prefix, plural := "", false
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		var ok bool
		prefix, plural, ok = multiTargetCardinalityPrefix(cardinality)
		if !ok {
			return false
		}
	}
	colorPrefix, ok := stackSpellColorPrefix(selection)
	if !ok || len(selection.ExcludedTypes) > 1 ||
		len(selection.RequiredTypesAny) > 1 && (colorPrefix != "" || len(selection.ExcludedTypes) != 0) {
		return false
	}
	excludedPrefix := ""
	if len(selection.ExcludedTypes) == 1 {
		word, ok := cardTypeWord(selection.ExcludedTypes[0])
		if !ok {
			return false
		}
		excludedPrefix = "non" + word + " "
	}
	nouns := make([]string, 0, len(selection.RequiredTypesAny))
	for _, cardType := range selection.RequiredTypesAny {
		word, ok := cardTypeWord(cardType)
		if !ok || slices.Contains(nouns, word) {
			return false
		}
		nouns = append(nouns, word)
	}
	expected := prefix + "target " + colorPrefix + excludedPrefix
	if len(nouns) > 0 {
		expected += joinUnionNouns(nouns) + " "
	}
	expected += "spell"
	if plural {
		expected += "s"
	}
	expected, ok = targetControllerSuffix(expected, selection.Controller)
	if !ok {
		return false
	}
	if selection.MatchManaValue {
		rider, ok := targetManaValueRider(selection.ManaValue)
		if !ok {
			return false
		}
		expected += rider
	}
	if len(selection.SpellTargetRestrictions) > 0 {
		rider, ok := spellTargetRestrictionsClause(selection.SpellTargetRestrictions)
		if !ok {
			return false
		}
		expected += rider
	}
	return strings.EqualFold(text, expected)
}

func stackSpellColorPrefix(selection SelectionSyntax) (string, bool) {
	shapes := len(selection.ColorsAny) + len(selection.ExcludedColors)
	if selection.Colorless {
		shapes++
	}
	if selection.Multicolored {
		shapes++
	}
	if shapes > 1 {
		return "", false
	}
	switch {
	case len(selection.ColorsAny) == 1:
		word, ok := colorWord(selection.ColorsAny[0])
		return word + " ", ok
	case len(selection.ExcludedColors) == 1:
		word, ok := colorWord(selection.ExcludedColors[0])
		return "non" + word + " ", ok
	case selection.Colorless:
		return "colorless ", true
	case selection.Multicolored:
		return "multicolored ", true
	default:
		return "", true
	}
}
