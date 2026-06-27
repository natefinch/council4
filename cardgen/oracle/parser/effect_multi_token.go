package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// multiTokenCreateSpecs splits a multi-token create clause into one parsed
// creature-token spec per created token ("Create a 1/1 green Snake creature
// token, a 2/2 green Wolf creature token, and a 3/3 green Elephant creature
// token." -> three specs). The clause is the run of tokens after the "Create"
// verb. It returns the specs in source order with ok=true only when the clause
// names two or more distinct tokens, every one of which is a fixed
// power/toughness creature token the single-token path already synthesizes.
// Every other shape — a single token, any quoted granted ability, a named or
// predefined artifact token, a variable "X/X" token, or a trailing count phrase
// — returns ok=false so the clause stays on its existing single-token path and
// fails closed when unsupported.
func multiTokenCreateSpecs(kind EffectKind, clause []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if kind != EffectCreate {
		return nil, false
	}
	segments := splitMultiTokenClause(clause)
	if len(segments) < 2 {
		return nil, false
	}
	specs := make([]EffectSyntax, 0, len(segments))
	for _, segment := range segments {
		power, toughness, ptKnown := parseTokenPowerToughness(EffectCreate, segment)
		if !ptKnown {
			return nil, false
		}
		selection := parseSelection(segment, atoms)
		if len(selection.SubtypesAny) < 1 {
			return nil, false
		}
		// parseSelection folds a token's single "with <keyword>" rider into the
		// selection's Keyword slot, but a created token's complete keyword list is
		// captured separately by parseTokenKeywords. Clear the selection keyword so
		// the token's keywords flow only through TokenKeywords; this keeps the
		// lowering from counting a single keyword twice (once from the selector and
		// once from the keyword list).
		selection.Keyword = KeywordUnknown
		specs = append(specs, EffectSyntax{
			Kind:                EffectCreate,
			Context:             EffectContextController,
			Selection:           selection,
			TokenPower:          power,
			TokenToughness:      toughness,
			TokenPTKnown:        ptKnown,
			TokenKeywords:       parseTokenKeywords(EffectCreate, segment, atoms),
			TokenName:           parseTokenName(EffectCreate, segment),
			TokenPredefinedName: parsePredefinedTokenName(EffectCreate, segment),
			Amount:              EffectAmountSyntax{Known: true, Value: 1},
			Tokens:              append([]shared.Token(nil), segment...),
			ClauseSpan:          shared.SpanOf(segment),
		})
	}
	return specs, true
}

// splitMultiTokenClause divides a create clause's post-verb tokens into one
// token-spec run per created token. Each spec begins with the article "a"/"an"
// at the start of the clause or immediately after a top-level comma or "and"
// connector, and must itself contain a "token"/"tokens" noun. The clause must
// hold two or more such specs and no quoted text; otherwise it returns nil so
// the caller leaves the clause on its single-token path. Trailing connector
// tokens (the comma and/or "and" that separate one spec from the next) are
// trimmed from each returned run.
func splitMultiTokenClause(clause []shared.Token) [][]shared.Token {
	for _, token := range clause {
		if token.Kind == shared.Quote {
			return nil
		}
	}
	var starts []int
	for i, token := range clause {
		if !equalWord(token, "a") && !equalWord(token, "an") {
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
// controller-creates form ("Create <spec>, <spec>, and <spec>.") with two or
// more fixed power/toughness creature-token specs; any other recipient,
// negation, or unreconstructable spec fails closed. It returns false when the
// effect carries no additional token specs, so single-token creates are
// untouched.
func exactCreateMultiTokenEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.AdditionalTokens) == 0 {
		return false
	}
	if effect.Context != EffectContextController || effect.Negated {
		return false
	}
	specs := make([]*EffectSyntax, 0, 1+len(effect.AdditionalTokens))
	specs = append(specs, effect)
	for i := range effect.AdditionalTokens {
		specs = append(specs, &effect.AdditionalTokens[i])
	}
	bodies := make([]string, 0, len(specs))
	for _, spec := range specs {
		if !spec.TokenPTKnown || spec.TokenPTVariableX || spec.TokenGrantedAbility != nil {
			return false
		}
		body, ok := creatureTokenSpecBody(spec)
		if !ok {
			return false
		}
		bodies = append(bodies, body(multiTokenArticle(spec), "token"))
	}
	var joined string
	if len(bodies) == 2 {
		joined = bodies[0] + " and " + bodies[1]
	} else {
		joined = strings.Join(bodies[:len(bodies)-1], ", ") + ", and " + bodies[len(bodies)-1]
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Create "+joined+".")
}

// multiTokenArticle returns the indefinite article a created creature token's
// spec is printed with: "an" before a power that reads with a leading vowel
// sound (8, 11, 18) when no leading "tapped" or "legendary" adjective intervenes,
// and "a" otherwise.
func multiTokenArticle(spec *EffectSyntax) string {
	if spec.Selection.Tapped && !spec.Selection.Attacking {
		return "a"
	}
	if len(spec.Selection.Supertypes) != 0 {
		return "a"
	}
	switch spec.TokenPower {
	case 8, 11, 18:
		return "an"
	default:
		return "a"
	}
}
