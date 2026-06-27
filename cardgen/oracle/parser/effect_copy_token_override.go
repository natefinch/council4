package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// copyTokenOverride holds the recognized characteristic overrides of a copy-token
// "except <override>" exception clause: a fixed power/toughness, colors, card
// types, subtypes, and granted keywords, with replacement-versus-additive
// semantics for the colors and types ("except it's a 1/1 green Frog" replaces;
// "except it's a 2/2 black Zombie in addition to its other colors and types"
// adds). The legendary drop ("it's not legendary") is recorded separately on the
// effect's TokenCopyDropLegendary.
type copyTokenOverride struct {
	dropLegendary  bool
	ptKnown        bool
	power          int
	toughness      int
	colors         []Color
	subtypes       []types.Sub
	cardTypes      []types.Card
	keywords       []KeywordKind
	additiveTypes  bool
	additiveColors bool
	// colorMode and subtypeMode track the additive choice already committed for
	// colors and subtypes (-1 unset, 0 replace, 1 additive) so a later clause that
	// disagrees fails closed rather than silently mixing semantics.
	colorMode   int
	subtypeMode int
}

// hasCharacteristic reports whether the override sets a characteristic beyond the
// legendary drop and granted keywords, which the bare not-legendary and
// keyword-rider paths already handle on their own.
func (o copyTokenOverride) hasCharacteristic() bool {
	return o.ptKnown || len(o.colors) > 0 || len(o.subtypes) > 0 || len(o.cardTypes) > 0
}

// copyTokenExceptOverride recognizes a copy-token characteristic-overriding
// "except <override>" exception from the effect's tokens. It returns the parsed
// override and ok=true only when every rider clause is recognized and at least
// one characteristic is set; any unrecognized wording, quoted granted ability, or
// parenthetical fails closed so the copy stays unsupported.
func copyTokenExceptOverride(effect *EffectSyntax) (copyTokenOverride, bool) {
	rider := copyTokenExceptRiderRun(effect.Tokens)
	if len(rider) == 0 {
		return copyTokenOverride{}, false
	}
	for _, token := range rider {
		if token.Kind == shared.Quote || token.Kind == shared.LeftParen {
			return copyTokenOverride{}, false
		}
	}
	clauses := splitCopyTokenOverrideClauses(rider)
	if len(clauses) == 0 {
		return copyTokenOverride{}, false
	}
	override := copyTokenOverride{colorMode: -1, subtypeMode: -1}
	for _, clause := range clauses {
		if !recognizeCopyTokenOverrideClause(clause, &override) {
			return copyTokenOverride{}, false
		}
	}
	if !override.hasCharacteristic() {
		return copyTokenOverride{}, false
	}
	return override, true
}

// copyTokenExceptRiderRun returns the tokens after the final "except" word with
// any trailing period stripped, or nil when there is no "except" word.
func copyTokenExceptRiderRun(tokens []shared.Token) []shared.Token {
	exceptIndex := -1
	for i := range tokens {
		if equalWord(tokens[i], "except") {
			exceptIndex = i
		}
	}
	if exceptIndex < 0 {
		return nil
	}
	rider := tokens[exceptIndex+1:]
	for len(rider) > 0 && rider[len(rider)-1].Kind == shared.Period {
		rider = rider[:len(rider)-1]
	}
	return rider
}

// splitCopyTokenOverrideClauses splits an override rider into individual clauses
// on top-level commas and on an "and" conjunction that introduces a new clause
// subject ("... and it's a ...", "... and it has ..."). An "and" inside a
// characteristic ("colors and types", "flying and haste") joins a single clause
// and is preserved.
func splitCopyTokenOverrideClauses(rider []shared.Token) [][]shared.Token {
	var clauses [][]shared.Token
	var current []shared.Token
	flush := func() {
		if len(current) > 0 {
			clauses = append(clauses, current)
			current = nil
		}
	}
	for i := range len(rider) {
		token := rider[i]
		if token.Kind == shared.Comma {
			flush()
			continue
		}
		if equalWord(token, "and") && i+1 < len(rider) && copyTokenOverrideClauseSubjectStart(rider[i+1]) {
			flush()
			continue
		}
		current = append(current, token)
	}
	flush()
	return clauses
}

// copyTokenOverrideClauseSubjectStart reports whether a token begins a new
// override clause subject ("it", "it's", or "the" of "the token").
func copyTokenOverrideClauseSubjectStart(token shared.Token) bool {
	return equalWord(token, "it") || equalWord(token, "it's") || equalWord(token, "the")
}

// recognizeCopyTokenOverrideClause recognizes a single override rider clause: the
// "it's not legendary" drop, an "it has <keyword>[ and <keyword>]" grant, or an
// "it's [a/an] <characteristics>[ in addition to its other [colors and] types]"
// characteristic override. It fails closed on any other wording.
func recognizeCopyTokenOverrideClause(clause []shared.Token, override *copyTokenOverride) bool {
	if copyTokenNotLegendaryClause(normalizedWords(clause)) {
		override.dropLegendary = true
		return true
	}
	if rest, ok := copyTokenHasSubject(clause); ok {
		keywords, kwOK := copyTokenOverrideKeywordList(rest)
		if !kwOK {
			return false
		}
		override.keywords = append(override.keywords, keywords...)
		return true
	}
	return recognizeCopyTokenCharacteristicClause(clause, override)
}

// recognizeCopyTokenCharacteristicClause recognizes the "it's [a/an]
// <characteristics>[ in addition to its other [colors and] types]" form and
// records its power/toughness, colors, card types, subtypes, and keyword rider on
// the override. Every word must be consumed by a recognized characteristic; any
// leftover word fails closed.
func recognizeCopyTokenCharacteristicClause(clause []shared.Token, override *copyTokenOverride) bool {
	rest, ok := copyTokenCharacteristicSubject(clause)
	if !ok {
		return false
	}
	if len(rest) > 0 && (equalWord(rest[0], "a") || equalWord(rest[0], "an")) {
		rest = rest[1:]
	}
	rest, additiveTypes, additiveColors := stripCopyTokenInAdditionSuffix(rest)
	spec := rest
	if withIndex := copyTokenWordIndex(rest, "with"); withIndex >= 0 {
		keywords, kwOK := copyTokenOverrideKeywordList(rest[withIndex+1:])
		if !kwOK {
			return false
		}
		override.keywords = append(override.keywords, keywords...)
		spec = rest[:withIndex]
	}
	var colors []Color
	var subtypes []types.Sub
	var cardTypes []types.Card
	index := 0
	if power, toughness, ptOK := copyTokenLeadingPowerToughness(spec); ptOK {
		override.ptKnown = true
		override.power = power
		override.toughness = toughness
		index = 3
	}
	for _, token := range spec[index:] {
		if token.Kind != shared.Word {
			return false
		}
		if color, colorOK := recognizeColorWord(token.Text); colorOK {
			colors = append(colors, color)
			continue
		}
		if cardType, typeOK := entersAsCopyAddTypeWord(token.Text); typeOK {
			cardTypes = append(cardTypes, cardType)
			continue
		}
		if subtype, subOK := recognizeSubtypePhrase(token.Text); subOK {
			subtypes = append(subtypes, subtype)
			continue
		}
		return false
	}
	if !override.ptKnown && len(colors) == 0 && len(subtypes) == 0 && len(cardTypes) == 0 {
		return false
	}
	// "in addition to its other types" keeps the copied colors, so a color word
	// in that form is contradictory and fails closed.
	if additiveTypes && !additiveColors && len(colors) > 0 {
		return false
	}
	if !applyCopyTokenColorMode(override, len(colors) > 0, additiveColors) {
		return false
	}
	if !applyCopyTokenSubtypeMode(override, len(subtypes) > 0, additiveTypes) {
		return false
	}
	override.colors = append(override.colors, colors...)
	override.subtypes = append(override.subtypes, subtypes...)
	override.cardTypes = append(override.cardTypes, cardTypes...)
	if additiveTypes {
		override.additiveTypes = true
	}
	if additiveColors {
		override.additiveColors = true
	}
	return true
}

// applyCopyTokenColorMode commits the additive-versus-replace choice for the
// override's colors and fails closed when a later clause disagrees with an
// already-committed choice.
func applyCopyTokenColorMode(override *copyTokenOverride, hasColors, additive bool) bool {
	if !hasColors {
		return true
	}
	mode := 0
	if additive {
		mode = 1
	}
	if override.colorMode >= 0 && override.colorMode != mode {
		return false
	}
	override.colorMode = mode
	return true
}

// applyCopyTokenSubtypeMode commits the additive-versus-replace choice for the
// override's subtypes and fails closed when a later clause disagrees.
func applyCopyTokenSubtypeMode(override *copyTokenOverride, hasSubtypes, additive bool) bool {
	if !hasSubtypes {
		return true
	}
	mode := 0
	if additive {
		mode = 1
	}
	if override.subtypeMode >= 0 && override.subtypeMode != mode {
		return false
	}
	override.subtypeMode = mode
	return true
}

// copyTokenCharacteristicSubject strips a leading "it's" / "it is" / "the token
// is" subject from a characteristic clause, returning the remaining tokens.
func copyTokenCharacteristicSubject(clause []shared.Token) ([]shared.Token, bool) {
	switch {
	case len(clause) >= 1 && equalWord(clause[0], "it's"):
		return clause[1:], true
	case len(clause) >= 2 && equalWord(clause[0], "it") && equalWord(clause[1], "is"):
		return clause[2:], true
	case len(clause) >= 3 && equalWord(clause[0], "the") && equalWord(clause[1], "token") && equalWord(clause[2], "is"):
		return clause[3:], true
	default:
		return nil, false
	}
}

// copyTokenLeadingPowerToughness reads a leading "<integer>/<integer>"
// power/toughness from a characteristic spec, returning the values when the spec
// begins with the symbol triple.
func copyTokenLeadingPowerToughness(spec []shared.Token) (power, toughness int, ok bool) {
	if len(spec) < 3 || spec[0].Kind != shared.Integer ||
		spec[1].Kind != shared.Slash || spec[2].Kind != shared.Integer {
		return 0, 0, false
	}
	power, err := strconv.Atoi(spec[0].Text)
	if err != nil {
		return 0, 0, false
	}
	toughness, err = strconv.Atoi(spec[2].Text)
	if err != nil {
		return 0, 0, false
	}
	return power, toughness, true
}

// copyTokenOverrideKeywordList parses a keyword list ("flying", "flying and
// haste", "flying, vigilance, and trample") into its keyword kinds, requiring
// every token to belong to a recognized keyword name or a separating
// comma/"and". It fails closed on any other token or an empty list.
func copyTokenOverrideKeywordList(tokens []shared.Token) ([]KeywordKind, bool) {
	var keywords []KeywordKind
	index := 0
	for index < len(tokens) {
		if tokens[index].Kind == shared.Comma || equalWord(tokens[index], "and") {
			index++
			continue
		}
		kind, width, ok := recognizeKeywordNameAt(tokens, index)
		if !ok {
			return nil, false
		}
		keywords = append(keywords, kind)
		index += width
	}
	if len(keywords) == 0 {
		return nil, false
	}
	return keywords, true
}

// stripCopyTokenInAdditionSuffix removes a trailing "in addition to its other
// types" or "in addition to its other colors and types" suffix from a
// characteristic spec, reporting whether the types and colors are additive.
func stripCopyTokenInAdditionSuffix(spec []shared.Token) (rest []shared.Token, additiveTypes, additiveColors bool) {
	colorsSuffix := []string{"in", "addition", "to", "its", "other", "colors", "and", "types"}
	typesSuffix := []string{"in", "addition", "to", "its", "other", "types"}
	if n := copyTokenTrailingWords(spec, colorsSuffix); n > 0 {
		return spec[:len(spec)-n], true, true
	}
	if n := copyTokenTrailingWords(spec, typesSuffix); n > 0 {
		return spec[:len(spec)-n], true, false
	}
	return spec, false, false
}

// copyTokenTrailingWords reports the length of words when the spec ends with
// exactly those words in order, or 0 otherwise.
func copyTokenTrailingWords(spec []shared.Token, words []string) int {
	if len(spec) < len(words) {
		return 0
	}
	offset := len(spec) - len(words)
	for i, word := range words {
		if !equalWord(spec[offset+i], word) {
			return 0
		}
	}
	return len(words)
}

// copyTokenWordIndex returns the index of the first token equal to word, or -1.
func copyTokenWordIndex(tokens []shared.Token, word string) int {
	for i := range tokens {
		if equalWord(tokens[i], word) {
			return i
		}
	}
	return -1
}

// applyCopyTokenOverride records a recognized characteristic override on the
// effect so the compiler and lowering can map it onto the runtime copy spec.
func applyCopyTokenOverride(effect *EffectSyntax, override copyTokenOverride) {
	effect.TokenCopyOverride = true
	effect.TokenCopyOverridePTKnown = override.ptKnown
	effect.TokenCopyOverridePower = override.power
	effect.TokenCopyOverrideToughness = override.toughness
	effect.TokenCopyOverrideColors = override.colors
	effect.TokenCopyOverrideSubtypes = override.subtypes
	effect.TokenCopyOverrideTypes = override.cardTypes
	effect.TokenCopyOverrideKeywords = override.keywords
	effect.TokenCopyOverrideAdditiveTypes = override.additiveTypes
	effect.TokenCopyOverrideAdditiveColors = override.additiveColors
}
