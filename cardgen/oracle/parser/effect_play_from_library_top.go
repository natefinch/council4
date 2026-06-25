package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePlayFromLibraryTopEffect recognizes the controller-scoped, turn-scoped
// grant "until end of turn, you may look at the top card of your library any
// time and you may play cards from the top of your library." (Gwenom,
// Remorseless). The leading "until end of turn," duration clause is stripped and
// recorded as the effect duration; the two "you may" permissions are folded into
// a single unconditional turn-scoped allowance (the player is never forced to
// look at or play the top card). Any other wording fails closed and flows
// through the generic effect parser.
func parsePlayFromLibraryTopEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	rest, duration := stripLeadingDurationClause(tokens, atoms)
	if duration != EffectDurationUntilEndOfTurn {
		return nil, false
	}
	words := wordsOnly(rest)
	want := []string{
		"you", "may", "look", "at", "the", "top", "card", "of", "your",
		"library", "any", "time", "and", "you", "may", "play", "cards",
		"from", "the", "top", "of", "your", "library",
	}
	if len(words) != len(want) {
		return nil, false
	}
	for i, w := range want {
		if !equalWord(words[i], w) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:       EffectPlayFromLibraryTop,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[2].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationUntilEndOfTurn,
		Exact:      true,
	}}, true
}

// playFromTopPayLifeRiderWords is the fixed word sequence of the "If you cast a
// spell this way, pay life equal to its mana value rather than pay its mana
// cost." rider, with punctuation removed.
var playFromTopPayLifeRiderWords = []string{
	"if", "you", "cast", "a", "spell", "this", "way", "pay", "life",
	"equal", "to", "its", "mana", "value", "rather", "than", "pay",
	"its", "mana", "cost",
}

// isPlayFromTopPayLifeRiderTokens reports whether the sentence tokens are the
// "If you cast a spell this way, pay life equal to its mana value rather than
// pay its mana cost." rider that folds onto a preceding play-from-library-top
// grant (Gwenom, Remorseless; Bolas's Citadel).
func isPlayFromTopPayLifeRiderTokens(tokens []shared.Token) bool {
	words := wordsOnly(tokens)
	if len(words) != len(playFromTopPayLifeRiderWords) {
		return false
	}
	for i, w := range playFromTopPayLifeRiderWords {
		if !equalWord(words[i], w) {
			return false
		}
	}
	return true
}

// playFromTopPayLifeRiderConditionAt reports whether the "if" condition intro at
// index i begins the pay-life rider ("If you cast a spell this way, pay life
// ..."). That rider is folded onto the play-from-library-top grant, so its "if"
// must not also surface as a standalone intervening condition.
func playFromTopPayLifeRiderConditionAt(tokens []shared.Token, i int) bool {
	words := normalizedWords(tokens[i:])
	if len(words) < len(playFromTopPayLifeRiderWords) {
		return false
	}
	for j, w := range playFromTopPayLifeRiderWords {
		if words[j] != w {
			return false
		}
	}
	return true
}

// creditPlayFromTopPayLifeRider folds the "If you cast a spell this way, pay
// life equal to its mana value rather than pay its mana cost." rider sentence
// onto the ability's lone play-from-library-top grant: it sets PlayFromTopPayLife
// plus a coverage span on the grant and clears the rider sentence's effects so
// reference and coverage scans credit it. It credits only when the ability holds
// exactly one play-from-library-top grant, that grant is exact, and the rider
// sentence is exactly the recognized pay-life clause; otherwise the rider stays
// uncredited and the card fails closed.
func creditPlayFromTopPayLifeRider(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	grant := lonePlayFromLibraryTopEffect(sentences)
	if grant == nil || !grant.Exact {
		return 0, 0, false
	}
	for i := range sentences {
		if len(sentences[i].Effects) == 0 {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		if !isPlayFromTopPayLifeRiderTokens(tokens) {
			continue
		}
		grant.PlayFromTopPayLife = true
		grant.PlayFromTopPayLifeRiderSpan = sentences[i].Span
		foldedEffects = len(sentences[i].Effects)
		if sentences[i].LegacyEffects {
			foldedLegacy = orderedEffectCount(tokens, atoms)
		}
		sentences[i].Effects = nil
		sentences[i].LegacyEffects = false
		sentences[i].PlayFromTopPayLifeRider = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// lonePlayFromLibraryTopEffect returns the single play-from-library-top grant
// effect across the sentences, or nil when the sentences hold zero or more than
// one such effect.
func lonePlayFromLibraryTopEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if sentences[i].Effects[j].Kind != EffectPlayFromLibraryTop {
				continue
			}
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	return found
}

// wordsOnly returns the tokens with comma and period punctuation removed, so
// fixed-wording recognizers can compare a clean word sequence.
func wordsOnly(tokens []shared.Token) []shared.Token {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Comma || token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	return words
}
