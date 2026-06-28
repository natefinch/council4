package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// annotateSacrificeThenCountSentences marks each "sacrifice {all <X> | any
// number of <X>}, then <create|draw|add> that many/much ..." count-scaled
// sacrifice sequence so lowering can publish the number sacrificed and scale the
// following reward by it. The sacrifice clause and the reward clause are parsed
// as an adjacent EffectSacrifice + reward pair; the "that many/much"
// back-reference cannot be expressed by the ordinary effect vocabulary, so this
// is the only place that wording is inspected. The annotation sets typed fields
// the text-blind lowering reads; the effects themselves are left in place.
func annotateSacrificeThenCountSentences(sentences []Sentence) {
	for si := range sentences {
		annotateSacrificeThenCountEffects(&sentences[si])
	}
}

func annotateSacrificeThenCountEffects(sentence *Sentence) {
	for i := 0; i+1 < len(sentence.Effects); i++ {
		sacrifice := &sentence.Effects[i]
		reward := &sentence.Effects[i+1]
		anyNumber, ok := matchSacrificeThenCountPair(sentence, sacrifice, reward)
		if !ok {
			continue
		}
		sacrifice.SacrificeThenCount = true
		sacrifice.SacrificeAnyNumber = anyNumber
	}
}

// matchSacrificeThenCountPair reports whether the sacrifice/reward effect pair is
// a controller "sacrifice {all|any number of} <permanents>, then <reward> that
// many/much" count-scaled sequence and, if so, whether the sacrifice form is
// "any number of" (true) rather than "all" (false).
func matchSacrificeThenCountPair(sentence *Sentence, sacrifice, reward *EffectSyntax) (anyNumber, ok bool) {
	if sacrifice.Kind != EffectSacrifice ||
		sacrifice.Context != EffectContextController ||
		sacrifice.Negated {
		return false, false
	}
	if reward.Connection != EffectConnectionThen ||
		reward.Context != EffectContextController ||
		reward.Negated {
		return false, false
	}
	switch reward.Kind {
	case EffectCreate, EffectDraw, EffectAddMana:
	default:
		return false, false
	}
	if !rewardReferencesSacrificedCount(tokensWithinParserSpan(sentence.Tokens, reward.Span)) {
		return false, false
	}
	anyNumber, ok = sacrificeCountForm(tokensWithinParserSpan(sentence.Tokens, sacrifice.Selection.Span))
	if !ok {
		return false, false
	}
	return anyNumber, true
}

// sacrificeCountForm reads the sacrifice selection tokens, returning whether the
// count form is "any number of <noun>" (true) or "all <noun>" (false). Any other
// leading quantifier is not a count-scaled sacrifice and is rejected.
func sacrificeCountForm(tokens []shared.Token) (anyNumber, ok bool) {
	if len(tokens) >= 3 &&
		isWord(tokens[0], "any") && isWord(tokens[1], "number") && isWord(tokens[2], "of") {
		return true, true
	}
	if len(tokens) >= 1 && isWord(tokens[0], "all") {
		return false, true
	}
	return false, false
}

// rewardReferencesSacrificedCount reports whether the reward clause tokens carry
// the "that many" or "that much" back-reference to the sacrificed count.
func rewardReferencesSacrificedCount(tokens []shared.Token) bool {
	for i := 0; i+1 < len(tokens); i++ {
		if isWord(tokens[i], "that") &&
			(isWord(tokens[i+1], "many") || isWord(tokens[i+1], "much")) {
			return true
		}
	}
	return false
}
