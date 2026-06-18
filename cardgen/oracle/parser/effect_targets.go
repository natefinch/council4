package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func parseTargets(tokens []shared.Token, atoms Atoms) []TargetSyntax {
	var targets []TargetSyntax
	for i, token := range tokens {
		plural := equalWord(token, "targets")
		if !equalWord(token, "target") && !plural {
			continue
		}
		start := i
		cardinality := TargetCardinalitySyntax{Min: 1, Max: 1}
		if enumStart, enumCard, ok := enumeratedTargetCardinality(tokens, i); ok {
			start = enumStart
			cardinality = enumCard
		} else {
			switch {
			case i >= 3 && effectWordsAt(tokens, i-3, "any", "number", "of"):
				start = i - 3
				cardinality = TargetCardinalitySyntax{Min: 0, Max: 99}
			case i >= 4 && effectWordsAt(tokens, i-4, "up", "to") &&
				(effectWordsAt(tokens, i-1, "another") || effectWordsAt(tokens, i-1, "other")):
				start = i - 4
				cardinality.Min = 0
				var ok bool
				cardinality.Max, ok = effectNumber(tokens[i-2], atoms)
				if !ok || cardinality.Max < 1 {
					cardinality = TargetCardinalitySyntax{}
				}
			case i >= 3 && effectWordsAt(tokens, i-3, "up", "to"):
				start = i - 3
				cardinality.Min = 0
				var ok bool
				cardinality.Max, ok = effectNumber(tokens[i-1], atoms)
				if !ok || cardinality.Max < 1 {
					cardinality = TargetCardinalitySyntax{}
				}
			case i >= 1:
				if count, ok := effectNumber(tokens[i-1], atoms); ok && count > 0 {
					start = i - 1
					cardinality = TargetCardinalitySyntax{Min: count, Max: count}
				} else if equalWord(tokens[i-1], "any") ||
					equalWord(tokens[i-1], "another") ||
					equalWord(tokens[i-1], "other") {
					start = i - 1
				}
			default:
			}
		}
		// A bare plural "targets" with no recognized preceding cardinality is not
		// a target production; only "<cardinality> targets" (e.g. "any number of
		// targets", "one or two targets") names targets directly.
		if plural && start == i {
			continue
		}
		end := targetSyntaxEnd(tokens, i+1)
		selectionTokens := append([]shared.Token(nil), tokens[start:i]...)
		selectionTokens = append(selectionTokens, tokens[i+1:end]...)
		selection := parseSelection(selectionTokens, atoms)
		if targetSelectionHasUnsupportedQualifier(selectionTokens, atoms) {
			selection = SelectionSyntax{Span: selection.Span, Text: selection.Text}
		}
		if plural {
			// "targets" with no following noun means "any target" — a permanent
			// or a player (CR 115.4).
			selection = SelectionSyntax{
				Span: shared.SpanOf(tokens[start:end]),
				Text: joinedEffectText(tokens[start:end]),
				Kind: SelectionAny,
			}
		}
		targets = append(targets, TargetSyntax{
			Span:        shared.SpanOf(tokens[start:end]),
			Text:        joinedEffectText(tokens[start:end]),
			Cardinality: cardinality,
			Selection:   selection,
			Exact:       exactRuntimeTargetSyntax(tokens[start:end], cardinality, selection),
		})
	}
	return targets
}

// enumeratedTargetCardinality recognizes the small fixed enumerations used by
// divided-damage wordings — "one or two" and "one, two, or three" — that precede
// the target word at index i. It returns the start index of the phrase and the
// inclusive count range, or ok=false when no enumeration is present.
func enumeratedTargetCardinality(tokens []shared.Token, i int) (int, TargetCardinalitySyntax, bool) {
	if i >= 3 &&
		equalWord(tokens[i-3], "one") &&
		equalWord(tokens[i-2], "or") &&
		equalWord(tokens[i-1], "two") {
		return i - 3, TargetCardinalitySyntax{Min: 1, Max: 2}, true
	}
	if i >= 6 &&
		equalWord(tokens[i-6], "one") &&
		tokens[i-5].Kind == shared.Comma &&
		equalWord(tokens[i-4], "two") &&
		tokens[i-3].Kind == shared.Comma &&
		equalWord(tokens[i-2], "or") &&
		equalWord(tokens[i-1], "three") {
		return i - 6, TargetCardinalitySyntax{Min: 1, Max: 3}, true
	}
	return 0, TargetCardinalitySyntax{}, false
}

func exactRuntimeTargetSyntax(tokens []shared.Token, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		return exactMultiPermanentTargetSyntax(joinedEffectText(tokens), cardinality, selection)
	}
	text := joinedEffectText(tokens)
	switch selection.Kind {
	case SelectionAny:
		return text == "any target"
	case SelectionPlayer:
		return strings.EqualFold(text, "target player")
	case SelectionOpponent:
		return strings.EqualFold(text, "target opponent")
	case SelectionActivatedAbility:
		return strings.EqualFold(text, "target activated ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionTriggeredAbility:
		return strings.EqualFold(text, "target triggered ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target activated or triggered ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionSpellActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target spell, activated ability, or triggered ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionTriggeredAbilityOrSpell:
		return selectionHasCounterAbilityQualifier(selection)
	case SelectionSpell:
		switch strings.ToLower(text) {
		case "target spell", "target instant spell", "target sorcery spell", "target creature spell",
			"target artifact spell", "target noncreature spell":
			return true
		}
		return exactSpellColorTargetSyntax(text, selection)
	case SelectionCreature:
		if strings.EqualFold(text, "target creature spell") {
			return true
		}
	case SelectionArtifact:
		if strings.EqualFold(text, "target artifact spell") {
			return true
		}
	default:
	}

	if len(selection.RequiredTypesAny) >= 2 {
		return exactTypeUnionTargetSyntax(text, selection)
	}
	if len(selection.ExcludedTypes) > 0 {
		return exactExcludedTypeTargetSyntax(text, selection)
	}
	if len(selection.ExcludedColors) > 0 {
		return exactExcludedColorTargetSyntax(text, selection)
	}

	expected, ok := exactPermanentTargetText(selection)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// exactMultiPermanentTargetSyntax reconstructs the canonical Oracle phrase for a
// multi-target or optional permanent target the executable backend lowers to a
// single multi-target spec: "up to one target <noun>" (Min 0, Max 1), the fixed
// "<N> target <noun>s" (Min N, Max N), and the optional "up to <N> target
// <noun>s" (Min 0, Max N) for a small cardinal N. It accepts only a plain
// permanent noun with an optional controller clause, failing closed for every
// other qualifier so unsupported plural wordings keep failing the byte-exact
// round-trip.
func exactMultiPermanentTargetSyntax(text string, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	prefix, plural, ok := multiTargetCardinalityPrefix(cardinality)
	if !ok {
		return false
	}
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.ExcludedTypes) != 0 || len(selection.Supertypes) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok || !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if plural {
		noun += "s"
	}
	expected, ok := targetControllerSuffix(prefix+"target "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// multiTargetCardinalityPrefix returns the canonical count words that precede
// "target" for a supported multi-target or optional cardinality, whether the
// target noun is plural, and whether the cardinality is one the round-trip
// represents. It fails closed for the unbounded "any number of" shape (Max 99),
// the divided-damage "one or two" ranges (Min neither 0 nor Max), and counts
// without a small-cardinal word.
func multiTargetCardinalityPrefix(c TargetCardinalitySyntax) (prefix string, plural, ok bool) {
	if c.Min == 0 && c.Max == 1 {
		return "up to one ", false, true
	}
	if c.Max < 2 {
		return "", false, false
	}
	word, found := cardinalWord(c.Max)
	if !found {
		return "", false, false
	}
	if c.Min == 0 {
		return "up to " + word + " ", true, true
	}
	if c.Min == c.Max {
		return word + " ", true, true
	}
	return "", false, false
}

// cardinalWord renders a small cardinal count (1..10) as its Oracle number word,
// the inverse of CardinalWordValue. It fails closed for counts outside that
// range so unbounded or unusual cardinalities cannot reconstruct exact wording.
func cardinalWord(n int) (string, bool) {
	switch n {
	case 1:
		return "one", true
	case 2:
		return "two", true
	case 3:
		return "three", true
	case 4:
		return "four", true
	case 5:
		return "five", true
	case 6:
		return "six", true
	case 7:
		return "seven", true
	case 8:
		return "eight", true
	case 9:
		return "nine", true
	case 10:
		return "ten", true
	default:
		return "", false
	}
}

// exactSpellColorTargetSyntax reconstructs the canonical Oracle phrase for a
// color-qualified spell target the executable backend can represent: a single
// color ("target blue spell"), a single excluded color ("target nonblue spell"),
// "target colorless spell", or "target multicolored spell". It fails closed for
// any combination of color shapes, monocolored spells, type/subtype/supertype
// filters, or controller and zone qualifiers, keeping unsupported wordings out of
// the byte-exact round-trip.
func exactSpellColorTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Controller != SelectionControllerAny ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		len(selection.RequiredTypesAny) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.Supertypes) != 0 || len(selection.SubtypesAny) != 0 {
		return false
	}
	colorShapes := len(selection.ColorsAny) + len(selection.ExcludedColors)
	if selection.Colorless {
		colorShapes++
	}
	if selection.Multicolored {
		colorShapes++
	}
	if colorShapes != 1 {
		return false
	}
	var qualifier string
	switch {
	case len(selection.ColorsAny) == 1:
		word, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return false
		}
		qualifier = word
	case len(selection.ExcludedColors) == 1:
		word, ok := colorWord(selection.ExcludedColors[0])
		if !ok {
			return false
		}
		qualifier = "non" + word
	case selection.Colorless:
		qualifier = "colorless"
	case selection.Multicolored:
		qualifier = "multicolored"
	default:
		return false
	}
	return strings.EqualFold(text, "target "+qualifier+" spell")
}

// exactPermanentTargetText reconstructs the canonical Oracle phrase for a single
// permanent target restricted only to qualifiers the executable backend can
// represent exactly: an "another"/"other" self-exclusion, a combat or tapped
// state, a single supertype, a single color, a single subtype that either
// qualifies an explicit type noun ("Beast creature") or stands in for it
// ("Soldier"), a "with power"/"with toughness" comparison, and a controller
// relation. It fails closed for every other qualifier so unsupported wordings
// keep failing the text-blind round-trip.
func exactPermanentTargetText(selection SelectionSyntax) (string, bool) {
	if selection.All || selection.Zone != zone.None ||
		selection.Keyword != KeywordUnknown ||
		selection.Colorless || selection.Multicolored ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.RequiredTypesAny) > 1 ||
		len(selection.ColorsAny) > 1 ||
		len(selection.SubtypesAny) > 1 ||
		len(selection.Supertypes) > 1 {
		return "", false
	}
	if (selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return "", false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if !hasNoun && selection.Kind != SelectionUnknown {
		return "", false
	}
	// The parser records a permanent noun both as the selection Kind and as a
	// redundant single-element RequiredTypesAny. Accept only that redundant form
	// (a type inconsistent with the noun is not representable here).
	if len(selection.RequiredTypesAny) == 1 {
		requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
		if !ok || !hasNoun || requiredNoun != noun {
			return "", false
		}
	}
	var words []string
	switch {
	case selection.Another:
		words = append(words, "another", "target")
	case selection.Other:
		words = append(words, "other", "target")
	default:
		words = append(words, "target")
	}
	switch {
	case selection.Attacking && selection.Blocking:
		words = append(words, "attacking", "or", "blocking")
	case selection.Attacking:
		words = append(words, "attacking")
	case selection.Blocking:
		words = append(words, "blocking")
	case selection.Tapped:
		words = append(words, "tapped")
	case selection.Untapped:
		words = append(words, "untapped")
	default:
	}
	if len(selection.Supertypes) == 1 {
		supertypeText, ok := supertypeWord(selection.Supertypes[0])
		if !ok {
			return "", false
		}
		words = append(words, supertypeText)
	}
	if len(selection.ColorsAny) == 1 {
		colorText, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return "", false
		}
		words = append(words, colorText)
	}
	if len(selection.SubtypesAny) == 1 {
		words = append(words, string(selection.SubtypesAny[0]))
	}
	switch {
	case hasNoun:
		words = append(words, noun)
	case len(selection.SubtypesAny) == 1:
	default:
		return "", false
	}
	numericWords, ok := permanentNumericQualifierWords(selection)
	if !ok {
		return "", false
	}
	words = append(words, numericWords...)
	return targetControllerSuffix(strings.Join(words, " "), selection.Controller)
}

// permanentNumericQualifierWords reconstructs the "with mana value"/"with
// power"/"with toughness" clause of a permanent target. It returns no words when
// the selection carries no mana value, power, or toughness comparison, and fails
// closed for any comparison shape the canonical phrasing cannot reproduce,
// keeping the text-blind round-trip honest.
func permanentNumericQualifierWords(selection SelectionSyntax) ([]string, bool) {
	var clauses [][]string
	if selection.MatchManaValue {
		clause, ok := comparisonClauseWords("mana value", selection.ManaValue)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if selection.MatchPower {
		clause, ok := comparisonClauseWords("power", selection.Power)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if selection.MatchToughness {
		clause, ok := comparisonClauseWords("toughness", selection.Toughness)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if len(clauses) == 0 {
		return nil, true
	}
	words := []string{"with"}
	for i, clause := range clauses {
		if i > 0 {
			words = append(words, "and")
		}
		words = append(words, clause...)
	}
	return words, true
}

// comparisonClauseWords renders a single "<qualifier> N", "<qualifier> N or less",
// or "<qualifier> N or greater" clause. It fails closed for comparison operators
// without a canonical Oracle phrasing the round-trip can reproduce.
func comparisonClauseWords(qualifier string, comparison compare.Int) ([]string, bool) {
	value := strconv.Itoa(comparison.Value)
	switch comparison.Op {
	case compare.Equal:
		return []string{qualifier, value}, true
	case compare.LessOrEqual:
		return []string{qualifier, value, "or", "less"}, true
	case compare.GreaterOrEqual:
		return []string{qualifier, value, "or", "greater"}, true
	default:
		return nil, false
	}
}

// exactTypeUnionTargetSyntax recognizes a permanent target whose only restriction
// is a union of card types, e.g. "target creature or planeswalker" or "target
// artifact or enchantment you control". It fails closed when any other qualifier
// (color, supertype, subtype, power, toughness, keyword, zone, combat or
// tapped state, "another"/"other", or excluded types) is present, or when any
// member is not a permanent card type.
func exactTypeUnionTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		len(selection.ExcludedTypes) != 0 || len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	nouns := make([]string, 0, len(selection.RequiredTypesAny))
	for _, cardType := range selection.RequiredTypesAny {
		noun, ok := permanentCardTypeNoun(cardType)
		if !ok {
			return false
		}
		nouns = append(nouns, noun)
	}
	expected := "target " + strings.Join(nouns, " or ")
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		expected += " you control"
	case SelectionControllerOpponent:
		expected += " an opponent controls"
	case SelectionControllerNotYou:
		expected += " you don't control"
	default:
		return false
	}
	return strings.EqualFold(text, expected)
}

// permanentCardTypeNoun returns the lowercase Oracle noun for a permanent card
// type. It fails closed for the non-permanent spell types (instant, sorcery).
func permanentCardTypeNoun(cardType CardType) (string, bool) {
	switch cardType {
	case CardTypeArtifact:
		return "artifact", true
	case CardTypeBattle:
		return "battle", true
	case CardTypeCreature:
		return "creature", true
	case CardTypeEnchantment:
		return "enchantment", true
	case CardTypeLand:
		return "land", true
	case CardTypePlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

// permanentSelectionNoun returns the lowercase Oracle noun for a permanent
// selection kind. It fails closed for non-permanent selection kinds.
func permanentSelectionNoun(kind SelectionKind) (string, bool) {
	switch kind {
	case SelectionArtifact:
		return "artifact", true
	case SelectionBattle:
		return "battle", true
	case SelectionCreature:
		return "creature", true
	case SelectionEnchantment:
		return "enchantment", true
	case SelectionLand:
		return "land", true
	case SelectionPermanent:
		return "permanent", true
	case SelectionPlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

// targetControllerSuffix appends the canonical controller clause for a target's
// controller relation, returning false for an unrecognized relation.
func targetControllerSuffix(expected string, controller SelectionController) (string, bool) {
	switch controller {
	case SelectionControllerAny:
		return expected, true
	case SelectionControllerYou:
		return expected + " you control", true
	case SelectionControllerOpponent:
		return expected + " an opponent controls", true
	case SelectionControllerNotYou:
		return expected + " you don't control", true
	default:
		return "", false
	}
}

// exactExcludedTypeTargetSyntax recognizes a permanent target whose only
// restriction is a single excluded card type ("target nonland permanent",
// "target noncreature artifact"). It fails closed when any other qualifier is
// present or when more than one type is excluded.
// selectionRedundantRequiredNoun reports whether selection's RequiredTypesAny is
// either empty or the single redundant card-type the parser records alongside a
// permanent noun Kind (e.g. "creature" recorded both as Kind and RequiredTypesAny).
// Excluded-color/type target reconstruction renders from Kind, so it accepts only
// that redundant form.
func selectionRedundantRequiredNoun(selection SelectionSyntax) bool {
	if len(selection.RequiredTypesAny) == 0 {
		return true
	}
	if len(selection.RequiredTypesAny) != 1 {
		return false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if !hasNoun {
		return false
	}
	requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
	return ok && requiredNoun == noun
}

func exactExcludedColorTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedColors) != 1 {
		return false
	}
	excludedColor, ok := colorWord(selection.ExcludedColors[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non"+excludedColor+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

func exactExcludedTypeTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedTypes) != 1 {
		return false
	}
	excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non"+excludedNoun+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

func targetSelectionHasUnsupportedQualifier(tokens []shared.Token, atoms Atoms) bool {
	for _, token := range tokens {
		if token.Kind == shared.Integer || token.Kind == shared.Comma ||
			selectionGrammarWord(token) || selectionAtomCoversToken(atoms, token) {
			continue
		}
		return true
	}
	return false
}

func selectionGrammarWord(token shared.Token) bool {
	for _, word := range []string{
		"a", "an", "all", "any", "number", "of", "up", "to", "or", "and",
		"with", "without", "from", "in", "your", "you", "control", "controls", "don't",
		"opponent", "opponent's", "opponents", "activated", "triggered", "source",
		"mana", "value", "power", "toughness", "equal", "less", "greater",
		"battlefield", "graveyard", "hand", "library", "exile", "command",
	} {
		if equalWord(token, word) {
			return true
		}
	}
	return false
}

func selectionAtomCoversToken(atoms Atoms, token shared.Token) bool {
	covered := func(span shared.Span) bool {
		return spanCovers(span, token.Span)
	}
	for _, atom := range atoms.Colors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedColors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ColorQualifiers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.CardTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Supertypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Subtypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ObjectNouns() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Zones() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Cardinals() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.SelectionFlags() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Controllers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.KeywordSelectors() {
		if covered(atom.Span) {
			return true
		}
	}
	return false
}

func targetSyntaxEnd(tokens []shared.Token, start int) int {
	if end, ok := counterAbilityListEnd(tokens, start); ok {
		return end
	}
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == shared.Comma || token.Kind == shared.Period || token.Kind == shared.Semicolon ||
			targetDestinationStartsAt(tokens, end) ||
			equalWord(token, "unless") ||
			(equalWord(token, "equal") && end+1 < len(tokens) && equalWord(tokens[end+1], "to")) ||
			(equalWord(token, "and") && end+2 < len(tokens) && equalWord(tokens[end+1], "you") && effectWordKind(tokens[end+2]) != EffectUnknown) ||
			(equalWord(token, "and") && end+1 < len(tokens) && effectWordKind(tokens[end+1]) != EffectUnknown) ||
			(end > start && effectWordKind(token) != EffectUnknown) ||
			(equalWord(token, "until") && end+1 < len(tokens)) ||
			(equalWord(token, "for") && effectWordsAt(tokens, end, "for", "as", "long", "as")) ||
			(equalWord(token, "as") && effectWordsAt(tokens, end, "as", "long", "as", "this")) {
			break
		}

		end++
	}

	return end
}

func targetDestinationStartsAt(tokens []shared.Token, index int) bool {
	if index < 0 || index > len(tokens) {
		return false
	}
	for _, phrase := range [][]string{
		{"to", "its", "owner's", "hand"},
		{"to", "their", "owners", "'", "hands"},
		{"to", "your", "hand"},
		{"to", "their", "hand"},
		{"to", "their", "hands"},
		{"to", "the", "battlefield"},
		{"onto", "the", "battlefield"},
		{"into", "your", "graveyard"},
		{"into", "your", "library"},
		{"on", "top", "of", "your", "library"},
		{"on", "the", "top", "of", "your", "library"},
		{"on", "bottom", "of", "your", "library"},
		{"on", "the", "bottom", "of", "your", "library"},
	} {
		if _, ok := cutTokenPrefix(tokens[index:], phrase...); ok {
			return true
		}
	}
	return false
}

func ambiguousZoneChoice(tokens []shared.Token, atoms Atoms, span shared.Span) bool {
	zones := atoms.Zones()
	for i, first := range zones {
		if !spanCovers(span, first.Span) {
			continue
		}
		for _, second := range zones[i+1:] {
			if first.Zone == second.Zone || !spanCovers(span, second.Span) {
				continue
			}
			for _, token := range tokens {
				if token.Span.Start.Offset >= first.Span.End.Offset &&
					token.Span.End.Offset <= second.Span.Start.Offset &&
					equalWord(token, "or") {
					return true
				}
			}
		}
	}
	return false
}

func parseSelection(tokens []shared.Token, atoms Atoms) SelectionSyntax {
	if recognized, ok := counterAbilitySelectionSyntax(tokens, shared.SpanOf(tokens), joinedEffectText(tokens)); ok {
		return recognized
	}
	selection := SelectionSyntax{Span: shared.SpanOf(tokens), Text: joinedEffectText(tokens)}
	words := normalizedWords(tokens)
	switch {
	case slices.Equal(words, []string{"activated", "ability"}):
		selection.Kind = SelectionActivatedAbility
	case slices.Equal(words, []string{"triggered", "ability"}):
		selection.Kind = SelectionTriggeredAbility
	case slices.Equal(words, []string{"activated", "or", "triggered", "ability"}):
		selection.Kind = SelectionActivatedOrTriggeredAbility
	case slices.Equal(words, []string{"spell", "activated", "ability", "or", "triggered", "ability"}):
		selection.Kind = SelectionSpellActivatedOrTriggeredAbility
	default:
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && selection.Kind == SelectionUnknown {
			selection.Kind = selectionKindForNoun(noun)
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && !slices.Contains(selection.RequiredTypesAny, cardType) {
			selection.RequiredTypesAny = append(selection.RequiredTypesAny, cardType)
		}
		if cardType, ok := atoms.ExcludedCardTypeAt(token.Span); ok && !slices.Contains(selection.ExcludedTypes, cardType) {
			selection.ExcludedTypes = append(selection.ExcludedTypes, cardType)
		}
		if colorValue, ok := atoms.ColorAt(token.Span); ok && !slices.Contains(selection.ColorsAny, colorValue) {
			selection.ColorsAny = append(selection.ColorsAny, colorValue)
		}
		if colorValue, ok := atoms.ExcludedColorAt(token.Span); ok && !slices.Contains(selection.ExcludedColors, colorValue) {
			selection.ExcludedColors = append(selection.ExcludedColors, colorValue)
		}
		if supertype, ok := atoms.SupertypeAt(token.Span); ok && !slices.Contains(selection.Supertypes, supertype) {
			selection.Supertypes = append(selection.Supertypes, supertype)
		}
		if qualifier, ok := atoms.ColorQualifierAt(token.Span); ok {
			switch qualifier {
			case ColorQualifierColorless:
				selection.Colorless = true
			case ColorQualifierMulticolored:
				selection.Multicolored = true
			default:
			}
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounSpell &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionSpell
			break
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounAbility &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionUnknown
			break
		}
	}
	span := shared.SpanOf(tokens)
	selection.SubtypesAny = atoms.SubtypesIn(span)
	if relation, ok := atoms.ControllerIn(span); ok {
		switch relation {
		case ControllerRelationYouControl:
			selection.Controller = SelectionControllerYou
		case ControllerRelationOpponentControls:
			selection.Controller = SelectionControllerOpponent
		case ControllerRelationYouDontControl:
			selection.Controller = SelectionControllerNotYou
		default:
		}
	}
	selection.Zone = firstZone(atoms, span, ZoneRoleFrom)
	if selection.Zone == zone.None {
		selection.Zone = firstZone(atoms, span, ZoneRolePlain)
	}
	switch {
	case effectContainsWords(words, "your", "graveyard"):
		selection.Controller = SelectionControllerYou
	case effectContainsWords(words, "opponent's", "graveyard"):
		selection.Controller = SelectionControllerOpponent
	default:
	}
	selection.All = slices.Contains(words, "all")
	selection.Another = atoms.SelectionFlagIn(span, SelectionFlagAnother)
	selection.Other = atoms.SelectionFlagIn(span, SelectionFlagOther)
	selection.Attacking = atoms.SelectionFlagIn(span, SelectionFlagAttacking)
	selection.Blocking = atoms.SelectionFlagIn(span, SelectionFlagBlocking)
	selection.Tapped = atoms.SelectionFlagIn(span, SelectionFlagTapped)
	selection.Untapped = atoms.SelectionFlagIn(span, SelectionFlagUntapped)
	if slices.Contains(words, "any") && selection.Kind == SelectionUnknown {
		selection.Kind = SelectionAny
	}
	if keyword, ok := atoms.KeywordSelectorIn(span, false); ok {
		selection.Keyword = keyword.Keyword
	}
	if !parseSelectionNumbers(tokens, atoms, &selection) {
		return SelectionSyntax{Span: span, Text: joinedEffectText(tokens)}
	}
	return selection
}

func parseSelectionNumbers(tokens []shared.Token, atoms Atoms, selection *SelectionSyntax) bool {
	for i := range tokens {
		if i+2 < len(tokens) && effectWordsAt(tokens, i, "mana", "value") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+2:], atoms)
			if !ok {
				return false
			}
			selection.ManaValue = comparison
			selection.MatchManaValue = true
			continue
		}
		if equalWord(tokens[i], "power") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Power = comparison
			selection.MatchPower = true
			continue
		}
		if equalWord(tokens[i], "toughness") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Toughness = comparison
			selection.MatchToughness = true
		}
	}
	return true
}

func parseSelectionNumberComparison(tokens []shared.Token, atoms Atoms) (compare.Int, bool) {
	if len(tokens) == 0 {
		return compare.Int{}, false
	}
	if value, ok := effectNumber(tokens[0], atoms); ok {
		if len(tokens) >= 3 && equalWord(tokens[1], "or") {
			switch {
			case equalWord(tokens[2], "less"):
				return compare.Int{Op: compare.LessOrEqual, Value: value}, true
			case equalWord(tokens[2], "greater"):
				return compare.Int{Op: compare.GreaterOrEqual, Value: value}, true
			}
		}
		return compare.Int{Op: compare.Equal, Value: value}, true
	}
	if len(tokens) >= 3 && effectWordsAt(tokens, 0, "equal", "to") {
		if value, ok := effectNumber(tokens[2], atoms); ok {
			return compare.Int{Op: compare.Equal, Value: value}, true
		}
	}
	return compare.Int{}, false
}

// staticGroupVerb reports whether token introduces a resolving plural creature
// or permanent group effect clause: "get"/"have" for a power/toughness or
// characteristic change, or "gain" for a keyword grant ("Creatures you control
// gain trample until end of turn."). The keyword-grant form lowers as a one-shot
// continuous effect over the affected group, mirroring the "get" pump form.
func staticGroupVerb(token shared.Token) bool {
	return equalWord(token, "get") || equalWord(token, "have") || equalWord(token, "gain")
}

func parseEffectStaticSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	subtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	subtypeKnown := func(index int) bool {
		_, ok := subtype(index)
		return ok
	}
	if subject, ok := parseColoredControlledCreatureGroup(tokens); ok {
		return subject
	}
	if subject, ok := parseBattlefieldCreatureGroupSubject(tokens, atoms); ok {
		return subject
	}
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "all", "other", "creatures") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllOtherCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "all", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "attacking", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttackingCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "blocking", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectBlockingCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "other", "creatures", "you", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "creatures", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "creatures", "your", "opponents", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "wall", "you", "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:4]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "walls", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:3]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "artifacts", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledArtifacts, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "tokens", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledTokens, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 6 && equalWord(tokens[0], "other") && equalWord(tokens[2], "creatures") &&
		effectWordsAt(tokens, 3, "you", "control") &&
		(equalWord(tokens[5], "have") || equalWord(tokens[5], "get")) &&
		subtypeKnown(1):
		value, _ := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:5]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}
	case len(tokens) >= 5 && equalWord(tokens[1], "creatures") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "have") || equalWord(tokens[4], "get")) &&
		subtypeKnown(0):
		value, _ := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "have") || equalWord(tokens[4], "get")):
		value, ok := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: ok}
	case len(tokens) >= 4 && effectWordsAt(tokens, 1, "you", "control") &&
		(equalWord(tokens[3], "have") || equalWord(tokens[3], "get")):
		value, ok := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: ok}
	default:
		return EffectStaticSubjectSyntax{}
	}
}

// parseBattlefieldCreatureGroupSubject recognizes battlefield-wide creature group
// subjects whose affected group spans every matching permanent regardless of
// controller: "Attacking creatures you control get/have ...", "All <Subtype>
// creatures get/have ...", and "Other <Subtype> creatures get/have ...". It
// returns the typed subject, or false so callers fall through to the bare
// grammar. The subtype forms require a known creature/kindred subtype so color
// and other qualifiers fail closed.
func parseBattlefieldCreatureGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	subtypeAt := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	switch {
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "attacking", "creatures") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledAttackingCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && equalWord(tokens[0], "all") && equalWord(tokens[2], "creatures") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		if value, ok := subtypeAt(1); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}, true
		}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && equalWord(tokens[2], "creatures") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		if value, ok := subtypeAt(1); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}, true
		}
	default:
	}
	return EffectStaticSubjectSyntax{}, false
}

// staticGroupColorFilter is a recognized color constraint on an affected creature
// group, holding the disjunctive single colors and the colorless/multicolored
// color-family qualifiers.
type staticGroupColorFilter struct {
	colors       []Color
	colorless    bool
	multicolored bool
}

// parseColoredControlledCreatureGroup recognizes a controller-permanent creature
// group carrying a color filter: "[Other] <color> creatures you control
// get/have ...". It returns the typed subject, mirroring the bare controlled and
// other-controlled creature group forms with the color predicate attached. It
// fails closed for any non-color qualifier so callers fall through to the bare
// grammar.
func parseColoredControlledCreatureGroup(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	colorIndex, kind, spanEnd := 0, EffectStaticSubjectControlledCreatures, 4
	if len(tokens) >= 1 && equalWord(tokens[0], "other") {
		colorIndex, kind, spanEnd = 1, EffectStaticSubjectOtherControlledCreatures, 5
	}
	filter, width, ok := staticColorFilterAt(tokens, colorIndex)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	creature := colorIndex + width
	if len(tokens) < creature+4 ||
		!equalWord(tokens[creature], "creatures") ||
		!effectWordsAt(tokens, creature+1, "you", "control") ||
		!staticGroupVerb(tokens[creature+3]) {
		return EffectStaticSubjectSyntax{}, false
	}
	return EffectStaticSubjectSyntax{
		Kind:         kind,
		Span:         shared.SpanOf(tokens[:spanEnd]),
		Colors:       filter.colors,
		Colorless:    filter.colorless,
		Multicolored: filter.multicolored,
	}, true
}

// staticColorFilterAt recognizes a single color word or color-family qualifier
// at index, returning the typed color filter and its token width. A bare color
// word ("red") yields a one-element colors slice; "colorless" and "multicolored"
// yield the matching qualifier flag. It fails closed for any other word,
// including "monocolored", which no Selection color filter can represent.
func staticColorFilterAt(tokens []shared.Token, index int) (staticGroupColorFilter, int, bool) {
	if index < 0 || index >= len(tokens) {
		return staticGroupColorFilter{}, 0, false
	}
	if value, ok := recognizeColorWord(tokens[index].Text); ok {
		return staticGroupColorFilter{colors: []Color{value}}, 1, true
	}
	switch qualifier, ok := recognizeColorQualifierWord(tokens[index].Text); {
	case ok && qualifier == ColorQualifierColorless:
		return staticGroupColorFilter{colorless: true}, 1, true
	case ok && qualifier == ColorQualifierMulticolored:
		return staticGroupColorFilter{multicolored: true}, 1, true
	}
	return staticGroupColorFilter{}, 0, false
}

func selectionKindForNoun(noun ObjectNoun) SelectionKind {
	switch noun {
	case ObjectNounArtifact:
		return SelectionArtifact
	case ObjectNounCard:
		return SelectionCard
	case ObjectNounCreature:
		return SelectionCreature
	case ObjectNounEnchantment:
		return SelectionEnchantment
	case ObjectNounLand:
		return SelectionLand
	case ObjectNounOpponent:
		return SelectionOpponent
	case ObjectNounPermanent:
		return SelectionPermanent
	case ObjectNounPlaneswalker:
		return SelectionPlaneswalker
	case ObjectNounPlayer:
		return SelectionPlayer
	case ObjectNounSpell:
		return SelectionSpell
	default:
		return SelectionUnknown
	}
}

func effectWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func effectContainsWords(words []string, sequence ...string) bool {
	for i := 0; i+len(sequence) <= len(words); i++ {
		if slices.Equal(words[i:i+len(sequence)], sequence) {
			return true
		}
	}
	return false
}

func joinedEffectText(tokens []shared.Token) string {
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && token.Span.Start.Offset > tokens[i-1].Span.End.Offset {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}
