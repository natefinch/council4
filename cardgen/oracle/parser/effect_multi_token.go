package parser

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// multiTokenCreateSpecs splits a multi-token create clause into one parsed
// token spec per created token ("Create a 1/1 green Snake creature token, a 2/2
// green Wolf creature token, and a 3/3 green Elephant creature token." -> three
// specs; "create X 1/1 white Halfling creature tokens and X Food tokens." -> two
// specs). The clause is the run of tokens after the "Create" verb. It returns
// the specs in source order with ok=true only when the clause names two or more
// tokens that share the same representable count — either a single token each
// (the "a"/"an" article form) or the spell's variable X count applied to every
// spec — and every spec is either a fixed power/toughness creature token the
// single-token path already synthesizes or a predefined artifact token (Food,
// Treasure, ...) the runtime already models. Every other shape — a single token,
// any quoted granted ability, mixed or unrepresentable counts, a variable "X/X"
// token, or a trailing count phrase — returns ok=false so the clause stays on
// its existing single-token path and fails closed when unsupported.
func multiTokenCreateSpecs(kind EffectKind, clause []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if kind != EffectCreate {
		return nil, false
	}
	segments := splitMultiTokenClause(clause)
	if len(segments) < 2 {
		return nil, false
	}
	specs := make([]EffectSyntax, 0, len(segments))
	var sharedAmount EffectAmountSyntax
	for i, segment := range segments {
		amount, ok := multiTokenSegmentAmount(segment)
		if !ok {
			return nil, false
		}
		if i == 0 {
			sharedAmount = amount
		} else if !sameMultiTokenAmount(sharedAmount, amount) {
			// Every spec must create the same number of tokens; a mixed clause
			// ("a Snake and X Food tokens") cannot share one Quantity and fails
			// closed rather than dropping or mis-counting a token type.
			return nil, false
		}
		spec, ok := multiTokenSegmentSpec(segment, amount, atoms)
		if !ok {
			return nil, false
		}
		specs = append(specs, spec)
	}
	return specs, true
}

// multiTokenSegmentSpec parses one split token-spec run into an EffectSyntax
// carrying the shared amount. The run is either a fixed power/toughness creature
// token (at least one subtype) or a predefined artifact token whose single
// subtype the runtime models (Food, Treasure, ...) with no printed
// power/toughness, color, keyword, or explicit name. Any other run returns
// ok=false so the whole multi-token clause fails closed.
func multiTokenSegmentSpec(segment []shared.Token, amount EffectAmountSyntax, atoms Atoms) (EffectSyntax, bool) {
	power, toughness, ptKnown := parseTokenPowerToughness(EffectCreate, segment)
	selection := parseSelection(segment, atoms)
	// parseSelection folds a token's single "with <keyword>" rider into the
	// selection's Keyword slot, but a created token's complete keyword list is
	// captured separately by parseTokenKeywords. Clear the selection keyword so
	// the token's keywords flow only through TokenKeywords; this keeps the
	// lowering from counting a single keyword twice (once from the selector and
	// once from the keyword list).
	selection.Keyword = KeywordUnknown
	keywords := parseTokenKeywords(EffectCreate, segment, atoms)
	tokenName := parseTokenName(EffectCreate, segment)
	predefinedName := parsePredefinedTokenName(EffectCreate, segment)
	switch {
	case ptKnown:
		if len(selection.SubtypesAny) < 1 {
			return EffectSyntax{}, false
		}
	default:
		// A predefined artifact token (Food, Treasure, ...) carries no printed
		// power/toughness; it is named by exactly one modeled subtype with no
		// color, keyword, explicit name, or predefined-name qualifier.
		if len(selection.SubtypesAny) != 1 ||
			!namedArtifactTokenSubtype(selection.SubtypesAny[0]) ||
			len(selection.ColorsAny) != 0 ||
			len(keywords) != 0 ||
			tokenName != "" ||
			predefinedName != "" {
			return EffectSyntax{}, false
		}
	}
	return EffectSyntax{
		Kind:                EffectCreate,
		Context:             EffectContextController,
		Selection:           selection,
		TokenPower:          power,
		TokenToughness:      toughness,
		TokenPTKnown:        ptKnown,
		TokenKeywords:       keywords,
		TokenToxic:          parseTokenKeywordToxic(EffectCreate, segment, atoms),
		TokenName:           tokenName,
		TokenPredefinedName: predefinedName,
		Amount:              amount,
		Tokens:              append([]shared.Token(nil), segment...),
		ClauseSpan:          shared.SpanOf(segment),
	}, true
}

// multiTokenSegmentAmount reads a split token-spec run's leading count. A run
// begins with the article "a"/"an" (one token) or the spell's variable "X"
// (the shared X count); every split run starts with one of these signals by
// construction of splitMultiTokenClause. Any other lead returns ok=false.
func multiTokenSegmentAmount(segment []shared.Token) (EffectAmountSyntax, bool) {
	if len(segment) == 0 {
		return EffectAmountSyntax{}, false
	}
	first := segment[0]
	switch {
	case equalWord(first, "X"):
		return EffectAmountSyntax{Span: first.Span, VariableX: true}, true
	case equalWord(first, "a") || equalWord(first, "an"):
		return EffectAmountSyntax{Known: true, Value: 1}, true
	default:
		return EffectAmountSyntax{}, false
	}
}

// sameMultiTokenAmount reports whether two multi-token segment amounts create the
// same number of tokens: both the spell's variable X, or both the same fixed
// count. It gates the shared-count requirement so every spec in a multi-token
// create emits with one common Quantity.
func sameMultiTokenAmount(a, b EffectAmountSyntax) bool {
	if a.VariableX || b.VariableX {
		return a.VariableX && b.VariableX
	}
	return a.Known && b.Known && a.Value == b.Value
}

// splitMultiTokenClause divides a create clause's post-verb tokens into one
// token-spec run per created token. Each spec begins with the article "a"/"an"
// or the spell's variable "X" at the start of the clause or immediately after a
// top-level comma or "and" connector, and must itself contain a "token"/"tokens"
// noun. The clause must hold two or more such specs and no quoted text;
// otherwise it returns nil so the caller leaves the clause on its single-token
// path. Trailing connector tokens (the comma and/or "and" that separate one spec
// from the next) are trimmed from each returned run.
func splitMultiTokenClause(clause []shared.Token) [][]shared.Token {
	for _, token := range clause {
		if token.Kind == shared.Quote {
			return nil
		}
	}
	var starts []int
	for i, token := range clause {
		if !multiTokenSpecStart(token) {
			continue
		}
		if i == 0 {
			starts = append(starts, i)
			continue
		}
		prev := clause[i-1]
		if prev.Kind == shared.Comma || equalWord(prev, "and") {
			starts = append(starts, i)
		}
	}
	if len(starts) < 2 {
		return nil
	}
	segments := make([][]shared.Token, 0, len(starts))
	for k, start := range starts {
		end := len(clause)
		if k+1 < len(starts) {
			end = starts[k+1]
		}
		segment := trimTrailingSeparators(clause[start:end])
		if !segmentHasTokenNoun(segment) {
			return nil
		}
		segments = append(segments, segment)
	}
	return segments
}

// trimTrailingSeparators drops trailing comma and "and" connector tokens from a
// split token-spec run so the run ends at the spec's own last word.
func trimTrailingSeparators(segment []shared.Token) []shared.Token {
	for len(segment) > 0 {
		last := segment[len(segment)-1]
		if last.Kind == shared.Comma || equalWord(last, "and") {
			segment = segment[:len(segment)-1]
			continue
		}
		break
	}
	return segment
}

// multiTokenSpecStart reports whether a token begins a token spec inside a
// multi-token create clause: the article "a"/"an" (one token) or the spell's
// variable "X" count. Restricting the signal to these leads keeps an internal
// color conjunction ("white and blue") from being mistaken for a spec boundary,
// since the word after such an "and" is a color rather than a count.
func multiTokenSpecStart(token shared.Token) bool {
	return equalWord(token, "a") || equalWord(token, "an") || equalWord(token, "X")
}

// segmentHasTokenNoun reports whether a split run contains a "token"/"tokens"
// noun, the minimum signal that the run names a created token rather than a
// stray "a"/"an" phrase (such as "a +1/+1 counter").
func segmentHasTokenNoun(segment []shared.Token) bool {
	for _, token := range segment {
		if equalWord(token, "token") || equalWord(token, "tokens") {
			return true
		}
	}
	return false
}

// exactCreateMultiTokenEffectSyntax recognizes a multi-token create clause by
// reconstructing each token's canonical spec from its parsed fields and
// byte-comparing the conjoined list against the clause text. It accepts only the
// controller-creates form ("Create <spec>, <spec>, and <spec>.") whose specs
// share one representable count — a single token each (the "a"/"an" article
// form) or the spell's variable X applied to every spec — and whose specs are
// each a fixed power/toughness creature token or a predefined artifact token
// (Food, Treasure, ...). Any other recipient, count, negation, or
// unreconstructable spec fails closed. It returns false when the effect carries
// no additional token specs, so single-token creates are untouched.
func exactCreateMultiTokenEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.AdditionalTokens) == 0 {
		return false
	}
	if effect.Context != EffectContextController || effect.Negated {
		return false
	}
	variableX := effect.Amount.VariableX
	if !variableX && (!effect.Amount.Known || effect.Amount.Value != 1) {
		return false
	}
	specs := make([]*EffectSyntax, 0, 1+len(effect.AdditionalTokens))
	specs = append(specs, effect)
	for i := range effect.AdditionalTokens {
		specs = append(specs, &effect.AdditionalTokens[i])
	}
	bodies := make([]string, 0, len(specs))
	for _, spec := range specs {
		if variableX {
			if !spec.Amount.VariableX {
				return false
			}
		} else if !spec.Amount.Known || spec.Amount.Value != 1 {
			return false
		}
		body, ok := multiTokenSpecBody(spec, variableX)
		if !ok {
			return false
		}
		bodies = append(bodies, body)
	}
	var joined string
	if len(bodies) == 2 {
		joined = bodies[0] + " and " + bodies[1]
	} else {
		joined = strings.Join(bodies[:len(bodies)-1], ", ") + ", and " + bodies[len(bodies)-1]
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Create "+joined+".")
}

// multiTokenSpecBody renders one multi-token spec's canonical clause body under
// the shared count (variable X or a single token each). A fixed power/toughness
// creature token reuses the single-token creature reconstruction; a predefined
// artifact token (Food, Treasure, ...) renders "<count> [tapped ]<Subtype>
// token(s)". It fails closed for a variable "X/X" token, a granted ability, or a
// predefined spec carrying any color, keyword, name, or extra selector.
func multiTokenSpecBody(spec *EffectSyntax, variableX bool) (string, bool) {
	noun := "token"
	if variableX {
		noun = "tokens"
	}
	if spec.TokenPTKnown {
		if spec.TokenPTVariableX || spec.TokenGrantedAbility != nil {
			return "", false
		}
		body, ok := creatureTokenSpecBody(spec)
		if !ok {
			return "", false
		}
		countWord := createTokenArticle(spec)
		if variableX {
			countWord = "X"
		}
		return body(countWord, noun), true
	}
	sel := spec.Selection
	if sel.Kind != SelectionUnknown ||
		len(sel.SubtypesAny) != 1 ||
		!namedArtifactTokenSubtype(sel.SubtypesAny[0]) ||
		sel.Keyword != KeywordUnknown ||
		len(sel.ColorsAny) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.RequiredTypesAny) != 0 || len(sel.ExcludedTypes) != 0 ||
		len(sel.SourceTypes) != 0 || len(sel.Supertypes) != 0 ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other ||
		sel.Colorless || sel.Multicolored ||
		spec.TokenName != "" || spec.TokenPredefinedName != "" ||
		len(spec.TokenKeywords) != 0 {
		return "", false
	}
	tappedPart := ""
	if sel.Tapped {
		tappedPart = "tapped "
	}
	countWord := "a"
	if variableX {
		countWord = "X"
	}
	return fmt.Sprintf("%s %s%s %s", countWord, tappedPart, string(sel.SubtypesAny[0]), noun), true
}

// createTokenArticle returns the indefinite article a created creature token's
// spec is printed with: "an" before a power that reads with a leading vowel
// sound (8, 11, 18) or a variable "X/X" token ("ex"), provided no leading
// "tapped" or "legendary" adjective intervenes, and "a" otherwise. The article
// agrees with the first spoken word of the rendered spec body, so a "tapped" or
// "legendary" lead reads "a" while the power/toughness lead governs the rest.
// It serves both the single-token and multi-token byte-exact reconstructions.
func createTokenArticle(spec *EffectSyntax) string {
	if spec.Selection.Tapped && !spec.Selection.Attacking {
		return "a"
	}
	if len(spec.Selection.Supertypes) != 0 {
		return "a"
	}
	if spec.TokenPTVariableX {
		return "an"
	}
	switch spec.TokenPower {
	case 8, 11, 18:
		return "an"
	default:
		return "a"
	}
}
