package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// grammaticalNumber selects singular or plural rendering of a selection noun
// phrase. The "all" mass form renders plural ("all creatures"); the "each" mass
// form renders singular ("each creature").
type grammaticalNumber int

const (
	// numberSingular renders the noun phrase in the singular ("creature").
	numberSingular grammaticalNumber = iota
	// numberPlural renders the noun phrase in the plural ("creatures").
	numberPlural
)

// determinerKind selects the leading determiner a caller wants on the rendered
// phrase. The mass family supplies its own determiner affix ("all"/"each")
// around the bare noun phrase, so it requests determinerNone. New determiner
// forms must be appended at the end of this const block.
type determinerKind int

const (
	// determinerNone renders no leading determiner; the caller owns it.
	determinerNone determinerKind = iota
)

// selectionPhraseOptions configures how selectionPhrase renders a noun phrase
// for one effect-family context. Number controls pluralization; Determiner,
// ZoneNoun, and CardNoun select wordings used by families migrated in later
// stages (card-zone and search nouns). This stage renders permanent-group noun
// phrases only, so non-zero Determiner/ZoneNoun/CardNoun fail closed.
type selectionPhraseOptions struct {
	Number     grammaticalNumber
	Determiner determinerKind
	ZoneNoun   bool
	CardNoun   bool
}

// selectionPhrase renders the canonical Oracle noun phrase for a permanent-group
// SelectionSyntax, owning its type, supertype, subtype-free, color, token,
// controller, keyword-free, tapped/combat, and numeric qualifiers. It returns
// ok=false (fail closed) for any selection field the permanent-group context
// cannot represent, so a caller that gates on a true result can trust that the
// rendered text fully and faithfully describes the typed selection. Counter,
// chosen-type, and destination qualifiers are family affixes the caller strips
// before rendering, so they are not part of the noun phrase and are ignored.
//
// It is the one canonical renderer the mass/all/each group family uses to verify
// that the typed selection matches the source text, rather than only validating
// the text shape. Subtype-noun and keyword forms remain owned by their existing
// family validators for now and fail closed here.
func selectionPhrase(selection SelectionSyntax, opts selectionPhraseOptions) (string, bool) {
	if opts.Determiner != determinerNone || opts.ZoneNoun || opts.CardNoun {
		return "", false
	}
	if !selectionPhraseRepresentable(selection) {
		return "", false
	}
	prefix, ok := selectionPhrasePrefixWords(selection)
	if !ok {
		return "", false
	}
	noun, ok := selectionPhraseNoun(selection, opts.Number == numberPlural)
	if !ok {
		return "", false
	}
	words := make([]string, 0, len(prefix)+len(noun))
	words = append(words, prefix...)
	words = append(words, noun...)
	numericWords, ok := permanentNumericQualifierWords(selection)
	if !ok {
		return "", false
	}
	words = append(words, numericWords...)
	controllerSuffix, ok := massControllerSuffixWords(selection)
	if !ok {
		return "", false
	}
	words = append(words, controllerSuffix...)
	return strings.Join(words, " "), true
}

// selectionPhraseRepresentable reports whether every selection qualifier the
// permanent-group noun phrase carries is one selectionPhrase renders. It fails
// closed for subtype, keyword, token, color-attribute, supertype (other than the
// nonbasic-land exclusion), name, and relative-comparison qualifiers, leaving
// those forms to their existing family validators. Counter and chosen-type
// fields are deliberately not rejected: they are trailing affixes the caller
// strips from the phrase before rendering.
func selectionPhraseRepresentable(selection SelectionSyntax) bool {
	if selection.Another || selection.OtherThanSource ||
		selection.NonToken || selection.TokenOnly ||
		selection.Colorless || selection.Multicolored || selection.BasicLandType ||
		selection.Historic || selection.ConjunctiveTypes || selection.PlayerOrPlaneswalker ||
		selection.MatchTotalManaValue || selection.InclusiveOneOfEach ||
		selection.ManaValueX || selection.EnteredThisTurn ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.NameUniqueAmongControlled || selection.OpponentEach {
		return false
	}
	if selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown {
		return false
	}
	if selection.Zone != zone.None {
		return false
	}
	if selection.ManaValueDynamic != "" || selection.RequiredName != "" {
		return false
	}
	if len(selection.SubtypesAny) != 0 || len(selection.ExcludedSubtypes) != 0 ||
		len(selection.Supertypes) != 0 || len(selection.SourceTypes) != 0 ||
		len(selection.Alternatives) != 0 {
		return false
	}
	return true
}

// selectionPhrasePrefixWords reconstructs the canonical adjective prefix that
// precedes a permanent-group noun: the "other" determiner-adjective, the
// combat/tapped state, a single color or excluded color, a single excluded card
// type, and the nonbasic-land supertype exclusion. It fails closed when more
// than one color or excluded color is named, or when an excluded supertype other
// than the basic-land exclusion is present.
func selectionPhrasePrefixWords(selection SelectionSyntax) ([]string, bool) {
	var words []string
	if selection.Other {
		words = append(words, "other")
	}
	combatWords, ok := selectionCombatStateWords(selection)
	if !ok {
		return nil, false
	}
	words = append(words, combatWords...)
	if len(selection.ColorsAny) > 1 || len(selection.ExcludedColors) > 1 {
		return nil, false
	}
	if len(selection.ColorsAny) == 1 {
		color, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return nil, false
		}
		words = append(words, color)
	}
	if len(selection.ExcludedColors) == 1 {
		color, ok := colorWord(selection.ExcludedColors[0])
		if !ok {
			return nil, false
		}
		words = append(words, "non"+color)
	}
	if len(selection.ExcludedTypes) > 1 {
		return nil, false
	}
	if len(selection.ExcludedTypes) == 1 {
		excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
		if !ok {
			return nil, false
		}
		words = append(words, "non"+excludedNoun)
	}
	if len(selection.ExcludedSupertypes) != 0 {
		if len(selection.ExcludedSupertypes) != 1 || selection.ExcludedSupertypes[0] != SupertypeBasic {
			return nil, false
		}
		words = append(words, "nonbasic")
	}
	return words, true
}

// selectionPhraseNoun reconstructs the permanent-group base noun from the typed
// selection: a multi-member type union renders as the canonical "and"-joined
// plural list ("creatures and lands"), and a single kind renders its lowercase
// permanent noun, pluralized when plural is set. A single redundant
// RequiredTypesAny entry must match the kind noun; any other type/kind
// combination fails closed.
func selectionPhraseNoun(selection SelectionSyntax, plural bool) ([]string, bool) {
	if len(selection.RequiredTypesAny) >= 2 {
		nouns := make([]string, 0, len(selection.RequiredTypesAny))
		seen := make(map[CardType]bool, len(selection.RequiredTypesAny))
		for _, cardType := range selection.RequiredTypesAny {
			if seen[cardType] {
				return nil, false
			}
			seen[cardType] = true
			noun, ok := permanentCardTypeNoun(cardType)
			if !ok {
				return nil, false
			}
			nouns = append(nouns, pluralizeNoun(noun, plural))
		}
		return []string{joinUnionNounsSep(nouns, "and")}, true
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return nil, false
	}
	if len(selection.RequiredTypesAny) == 1 {
		requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
		if !ok || requiredNoun != noun {
			return nil, false
		}
	}
	return []string{pluralizeNoun(noun, plural)}, true
}

// pluralizeNoun appends the regular "s" plural to a permanent noun when plural is
// set. Every permanent card-type noun (creature, artifact, enchantment, land,
// planeswalker, permanent, battle) pluralizes regularly, so the simple rule is
// exact for this context.
func pluralizeNoun(noun string, plural bool) string {
	if plural {
		return noun + "s"
	}
	return noun
}

// massControllerSuffixWords reconstructs the trailing controller clause of a mass
// group: the plural "your opponents control" for an opponent-controlled group,
// "you control"/"you don't control" for the controller relations, and no clause
// for an uncontrolled group. It fails closed for any other controller relation.
func massControllerSuffixWords(selection SelectionSyntax) ([]string, bool) {
	switch selection.Controller {
	case SelectionControllerAny:
		return nil, true
	case SelectionControllerYou:
		return []string{"you", "control"}, true
	case SelectionControllerNotYou:
		return []string{"you", "don't", "control"}, true
	case SelectionControllerOpponent:
		return []string{"your", "opponents", "control"}, true
	default:
		return nil, false
	}
}
