package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseGrantSpellKeywordEffect recognizes the controller-scoped, turn-scoped
// resolving buff "The next spell you cast this turn has <keyword>." (Archway of
// Innovation) and its all-spells form "Spells you cast this turn have
// <keyword>.". It grants a cost-affecting keyword (Improvise, Convoke, or Delve)
// to the matching spells the controller casts for the rest of the turn. The
// next-spell form sets GrantSpellKeywordNextOnly so the grant is consumed by the
// single next spell the controller casts. Any other wording, keyword, or shape
// fails closed so the generic effect classifier reports it unsupported.
func parseGrantSpellKeywordEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	index := 0
	nextOnly := false
	switch {
	case len(words) >= 3 && equalWord(words[0], "the") && equalWord(words[1], "next") && equalWord(words[2], "spell"):
		nextOnly = true
		index = 3
	case len(words) >= 1 && equalWord(words[0], "spells"):
		index = 1
	default:
		return nil, false
	}
	if index+6 != len(words) ||
		!equalWord(words[index], "you") ||
		!equalWord(words[index+1], "cast") ||
		!equalWord(words[index+2], "this") ||
		!equalWord(words[index+3], "turn") {
		return nil, false
	}
	verbToken := words[index+4]
	if nextOnly {
		if !equalWord(verbToken, "has") {
			return nil, false
		}
	} else if !equalWord(verbToken, "have") {
		return nil, false
	}
	keyword, ok := costGrantKeywordWord(words[index+5])
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                      EffectGrantSpellKeyword,
		Span:                      sentence.Span,
		ClauseSpan:                sentence.Span,
		VerbSpan:                  verbToken.Span,
		Text:                      sentence.Text,
		Tokens:                    append([]shared.Token(nil), tokens...),
		Context:                   EffectContextController,
		Duration:                  EffectDurationThisTurn,
		GrantSpellKeyword:         keyword,
		GrantSpellKeywordNextOnly: nextOnly,
		Exact:                     true,
	}}, true
}

// costGrantKeywordWord maps a single keyword word onto the cost-affecting
// KeywordKind a spell keyword grant may confer. Only the cost-reducing keywords
// Improvise, Convoke, and Delve the payment machinery honors are recognized; any
// other word fails closed.
func costGrantKeywordWord(token shared.Token) (KeywordKind, bool) {
	if token.Kind != shared.Word {
		return KeywordUnknown, false
	}
	switch {
	case equalWord(token, "improvise"):
		return KeywordImprovise, true
	case equalWord(token, "convoke"):
		return KeywordConvoke, true
	case equalWord(token, "delve"):
		return KeywordDelve, true
	default:
		return KeywordUnknown, false
	}
}

// grantSpellKeywordVerbAt reports whether the "has"/"have" verb at index is the
// grant verb of a "[The next] spell[s] you cast this turn has/have <keyword>."
// clause (Archway of Innovation). Its dedicated recognizer produces a single
// EffectGrantSpellKeyword, so the effect classifier must not also treat this
// grant verb as a separate effect boundary alongside the sentence's "cast" verb.
func grantSpellKeywordVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "has") && !equalWord(tokens[index], "have") {
		return false
	}
	if index+1 >= len(tokens) {
		return false
	}
	if _, ok := costGrantKeywordWord(tokens[index+1]); !ok {
		return false
	}
	if index+2 < len(tokens) && tokens[index+2].Kind != shared.Period {
		return false
	}
	if index < 2 || !equalWord(tokens[index-2], "this") || !equalWord(tokens[index-1], "turn") {
		return false
	}
	for i := 0; i < index-2; i++ {
		if equalWord(tokens[i], "cast") {
			return true
		}
	}
	return false
}
