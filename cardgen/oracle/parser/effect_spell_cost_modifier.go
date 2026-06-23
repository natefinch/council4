package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// spellCostModifierMatch holds the parsed pieces of a resolving spell cost
// modifier sentence.
type spellCostModifierMatch struct {
	caster        SpellCostCasterKind
	amount        int
	increase      bool
	requiredTypes []CardType
	excludedTypes []CardType
	duration      EffectDurationKind
	verbSpan      shared.Span
}

// matchSpellCostModifier reports whether tokens form the one-shot,
// duration-bounded spell cost modifier "[<type filter>] spells <caster> cast
// cost {N} more/less to cast" scoped by a recognized finite duration ("Artifact
// spells you cast this turn cost {1} less to cast.", Armor Wars chapter II;
// "Until your next turn, spells your opponents cast cost {1} more to cast.", Tax
// Collector; "Noncreature spells your opponents cast cost {2} more to cast until
// your next turn.", Elspeth Conquers Death chapter II). The optional leading
// filter constrains the affected spells to a single card type (a bare type such
// as "Artifact") or exempts a single card type (a "non"-prefixed word such as
// "Noncreature"). The caster phrase is "you cast" (controller), "your opponents
// cast" (opponents), or absent (every player). The duration may lead the
// sentence ("Until your next turn, ..."), sit before the cost clause ("... you
// cast this turn cost {1} less ..."), or trail it. A sentence without a
// recognized finite duration fails closed so the permanent static cost-modifier
// wording is never hijacked by the resolving recognizer.
func matchSpellCostModifier(tokens []shared.Token) (spellCostModifierMatch, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	words, duration := extractSpellCostDuration(words)
	if duration == EffectDurationNone {
		return spellCostModifierMatch{}, false
	}

	index := 0
	var requiredTypes, excludedTypes []CardType
	if len(words) > 0 {
		if cardType, ok := recognizeCardTypeWord(words[0].Text); ok {
			requiredTypes = []CardType{cardType}
			index++
		} else if cardType, ok := recognizeExcludedCardTypeWord(words[0].Text); ok {
			excludedTypes = []CardType{cardType}
			index++
		}
	}

	if index >= len(words) || !equalWord(words[index], "spells") {
		return spellCostModifierMatch{}, false
	}
	index++

	caster := SpellCostCasterAll
	switch {
	case index+1 < len(words) && equalWord(words[index], "your") && equalWord(words[index+1], "opponents"):
		caster = SpellCostCasterOpponents
		index += 2
		if index >= len(words) || !equalWord(words[index], "cast") {
			return spellCostModifierMatch{}, false
		}
		index++
	case index < len(words) && equalWord(words[index], "you"):
		caster = SpellCostCasterController
		index++
		if index >= len(words) || !equalWord(words[index], "cast") {
			return spellCostModifierMatch{}, false
		}
		index++
	default:
	}

	// The remaining words must be exactly "cost(s) {N} more/less to cast".
	rest := words[index:]
	if len(rest) != 5 ||
		(!equalWord(rest[0], "cost") && !equalWord(rest[0], "costs")) ||
		rest[1].Kind != shared.Symbol ||
		!equalWord(rest[3], "to") ||
		!equalWord(rest[4], "cast") {
		return spellCostModifierMatch{}, false
	}
	amount, ok := staticGenericSymbolValue(rest[1].Text)
	if !ok || amount <= 0 {
		return spellCostModifierMatch{}, false
	}
	increase := false
	switch {
	case equalWord(rest[2], "more"):
		increase = true
	case equalWord(rest[2], "less"):
		increase = false
	default:
		return spellCostModifierMatch{}, false
	}

	return spellCostModifierMatch{
		caster:        caster,
		amount:        amount,
		increase:      increase,
		requiredTypes: requiredTypes,
		excludedTypes: excludedTypes,
		duration:      duration,
		verbSpan:      rest[0].Span,
	}, true
}

// parseSpellCostModifierEffect recognizes a resolving spell cost modifier
// sentence and emits the matching EffectSpellCostModifier clause.
func parseSpellCostModifierEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	match, ok := matchSpellCostModifier(tokens)
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                           EffectSpellCostModifier,
		Span:                           sentence.Span,
		ClauseSpan:                     sentence.Span,
		VerbSpan:                       match.verbSpan,
		Text:                           sentence.Text,
		Tokens:                         append([]shared.Token(nil), tokens...),
		Context:                        EffectContextController,
		Duration:                       match.duration,
		SpellCostModifierCaster:        match.caster,
		SpellCostModifierAmount:        match.amount,
		SpellCostModifierIncrease:      match.increase,
		SpellCostModifierRequiredTypes: match.requiredTypes,
		SpellCostModifierExcludedTypes: match.excludedTypes,
		Exact:                          true,
	}}, true
}

// extractSpellCostDuration removes a recognized finite duration phrase ("this
// turn", "until your next turn", "until end of turn", "until the end of your
// next turn") from anywhere in the word run, returning the remaining words and
// the duration it names. The phrase may lead the sentence, sit between the cast
// verb and the cost clause ("... you cast this turn cost {1} less ..."), or
// trail it ("... cost {2} more to cast until your next turn"). It returns
// EffectDurationNone with the words unchanged when no phrase matches.
func extractSpellCostDuration(words []shared.Token) ([]shared.Token, EffectDurationKind) {
	type phrase struct {
		words    []string
		duration EffectDurationKind
	}
	phrases := []phrase{
		{[]string{"until", "the", "end", "of", "your", "next", "turn"}, EffectDurationUntilEndOfYourNextTurn},
		{[]string{"until", "your", "next", "turn"}, EffectDurationUntilYourNextTurn},
		{[]string{"until", "end", "of", "turn"}, EffectDurationUntilEndOfTurn},
		{[]string{"this", "turn"}, EffectDurationThisTurn},
	}
	for _, candidate := range phrases {
		for start := 0; start+len(candidate.words) <= len(words); start++ {
			if !spellCostWordsAt(words, start, candidate.words) {
				continue
			}
			remaining := make([]shared.Token, 0, len(words)-len(candidate.words))
			remaining = append(remaining, words[:start]...)
			remaining = append(remaining, words[start+len(candidate.words):]...)
			// A leading "Until your next turn," clause leaves a separating comma
			// at the front of the remaining words; drop it so the type filter or
			// "spells" word leads the parse.
			for len(remaining) > 0 && remaining[0].Kind == shared.Comma {
				remaining = remaining[1:]
			}
			return remaining, candidate.duration
		}
	}
	return words, EffectDurationNone
}

func spellCostWordsAt(words []shared.Token, start int, want []string) bool {
	for offset, word := range want {
		if !equalWord(words[start+offset], word) {
			return false
		}
	}
	return true
}

// spellCostModifierCastAt reports whether the "cast" token at index belongs to a
// resolving spell cost modifier sentence. Such a sentence carries two "cast"
// tokens ("spells you cast ... cost {N} less to cast"); the dedicated recognizer
// produces a single effect, so the legacy verb counter must not double-count its
// casts as separate effects and force the clause down the ordered-sequence path.
func spellCostModifierCastAt(tokens []shared.Token, index int) bool {
	_ = index
	_, ok := matchSpellCostModifier(tokens)
	return ok
}
