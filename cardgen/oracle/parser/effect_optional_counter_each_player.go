package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitOptionalCounterForEachPlayerSequences folds a per-player optional counter
// placement followed by "Goad each creature that had counters put on it this
// way." into one typed clause. The reminder text explaining goad is left alone.
func emitOptionalCounterForEachPlayerSequences(abilities []Ability) {
	for i := range abilities {
		recognizeOptionalCounterForEachPlayerSequence(&abilities[i])
	}
}

func recognizeOptionalCounterForEachPlayerSequence(ability *Ability) {
	if ability == nil || ability.OptionalCounterForEachPlayer != nil ||
		ability.Vote != nil || ability.Modal != nil || ability.DiceTable != nil ||
		ability.CoinFlip != nil {
		return
	}
	sentences := nonReminderSentences(ability.Sentences)
	if len(sentences) != 2 {
		return
	}
	counterSentence := sentences[0]
	goadSentence := sentences[1]
	if len(counterSentence.Effects) != 1 {
		return
	}
	put := &counterSentence.Effects[0]
	if put.Kind != EffectPut ||
		!put.Optional ||
		!put.CounterKnown ||
		!put.Amount.Known ||
		put.Amount.Value <= 0 ||
		put.Selection.Kind == SelectionUnknown ||
		(put.Context != EffectContextEachPlayer && put.Context != EffectContextEachOpponent) ||
		!optionalCounterSentenceWords(counterSentence.Tokens, put.Context) ||
		!goadCountersPlacedThisWayWords(goadSentence.Tokens) {
		return
	}

	construct := counterSentence.Span
	if goadSentence.Span.End.Offset > construct.End.Offset {
		construct.End = goadSentence.Span.End
	}
	ability.OptionalCounterForEachPlayer = &OptionalCounterForEachPlayerClause{
		PlayerContext: put.Context,
		Pool:          put.Selection,
		Amount:        put.Amount,
		CounterKind:   put.CounterKind,
		Spans:         []shared.Span{counterSentence.Span, goadSentence.Span},
		ConstructSpan: construct,
	}
	counterSentence.Effects = nil
	goadSentence.Effects = nil
	ability.SemanticReferences = nil
}

func nonReminderSentences(sentences []Sentence) []*Sentence {
	result := make([]*Sentence, 0, len(sentences))
	for i := range sentences {
		if !isReminderSentence(sentences[i]) {
			result = append(result, &sentences[i])
		}
	}
	return result
}

func optionalCounterSentenceWords(tokens []shared.Token, context EffectContextKind) bool {
	tokens = semanticEffectTokens(tokens)
	if len(tokens) < 10 || tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	groupWord := "player"
	if context == EffectContextEachOpponent {
		groupWord = "opponent"
	}
	return effectWordsAt(tokens, 0, "each", groupWord, "may", "put") &&
		effectWordsAt(tokens, len(tokens)-3, "they", "control")
}

func goadCountersPlacedThisWayWords(tokens []shared.Token) bool {
	tokens = semanticEffectTokens(tokens)
	words := []string{"goad", "each", "creature", "that", "had", "counters", "put", "on", "it", "this", "way"}
	if len(tokens) != len(words)+1 || tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	return effectWordsAt(tokens, 0, words...)
}
