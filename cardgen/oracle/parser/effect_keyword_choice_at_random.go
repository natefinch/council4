package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// creditChooseKeywordAtRandomGrant folds the two-sentence "choose <keyword> or
// <keyword> at random. <source> gains that ability until end of turn."
// construction (Blitzwing, Adaptive Assailant) onto a single keyword-choice
// grant. The first sentence lists the candidate keywords and selects one at
// random; the second grants "that ability" (the chosen keyword) to the source
// until end of turn. The parser splits the two sentences, leaving the prelude
// with no effect and the grant sentence with a lone EffectGain whose body is the
// "that ability" anaphor. This recognizer marks that grant as an at-random
// keyword-choice grant, records the prelude span so lowering widens the trigger
// body to cover the listed keywords, and credits the prelude sentence so
// coverage and reference scans account for its tokens. It reports whether the
// fold applied. It credits only the exact adjacent prelude-then-grant shape;
// any other wording leaves the sentences untouched and the card fails closed.
func creditChooseKeywordAtRandomGrant(sentences []Sentence, atoms Atoms) bool {
	for i := 0; i+1 < len(sentences); i++ {
		prelude := &sentences[i]
		grantSentence := &sentences[i+1]
		if len(prelude.Effects) != 0 || len(grantSentence.Effects) != 1 {
			continue
		}
		if !isChooseKeywordAtRandomPrelude(prelude, atoms) {
			continue
		}
		grant := &grantSentence.Effects[0]
		if !isThatAbilityGrantEffect(grant) {
			continue
		}
		grant.KeywordGrantChoice = true
		grant.KeywordGrantChoiceAtRandom = true
		grant.Exact = true
		grant.KeywordChoiceAtRandomPreludeSpan = prelude.Span
		prelude.KeywordChoiceAtRandomPrelude = true
		return true
	}
	return false
}

// isChooseKeywordAtRandomPrelude reports whether the sentence tokens are exactly
// "choose <keyword> or <keyword>[ or <keyword> ...] at random." with two or more
// simple grantable keyword abilities. The keywords must be plain (no parameter)
// so lowering can grant the chosen one to the source; any parameterized keyword,
// missing "at random" suffix, or non-keyword body leaves the prelude unrecognized.
func isChooseKeywordAtRandomPrelude(sentence *Sentence, atoms Atoms) bool {
	tokens := semanticEffectTokens(sentence.Tokens)
	if len(tokens) < 6 || !equalWord(tokens[0], "choose") {
		return false
	}
	last := len(tokens) - 1
	if tokens[last].Kind != shared.Period ||
		!equalWord(tokens[last-1], "random") ||
		!equalWord(tokens[last-2], "at") {
		return false
	}
	listTokens := tokens[1 : last-2]
	keywords := atoms.KeywordsWithin(listTokens)
	if len(keywords) < 2 {
		return false
	}
	for _, keyword := range keywords {
		if keyword.Parameter.Kind != KeywordParameterNone {
			return false
		}
	}
	return exactKeywordChoiceList(joinedEffectText(listTokens))
}

// isThatAbilityGrantEffect reports whether the effect is the "<source> gains that
// ability until end of turn." grant that pairs with a choose-keyword-at-random
// prelude. The subject is the source, the grant lasts until end of turn, and its
// body is the "that ability" anaphor that denotes the randomly chosen keyword
// rather than a spelled-out keyword list (so it is not itself a keyword-choice
// grant yet).
func isThatAbilityGrantEffect(effect *EffectSyntax) bool {
	if effect.Kind != EffectGain ||
		effect.Context != EffectContextSource ||
		effect.Duration != EffectDurationUntilEndOfTurn ||
		effect.KeywordGrantChoice {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	_, after, ok := strings.Cut(text, " gains ")
	if !ok {
		return false
	}
	return after == "that ability until end of turn."
}
