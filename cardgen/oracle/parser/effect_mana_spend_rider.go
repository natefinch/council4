package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// manaSpendRiderWords is the exact leading token sequence of the
// commander-creature-type spend condition.
var manaSpendRiderWords = []string{
	"when", "that", "mana", "is", "spent", "to", "cast",
	"a", "creature", "spell", "that", "shares", "a", "creature",
	"type", "with", "your", "commander",
}

var chosenTypeManaSpendConditionWords = []string{
	"spend", "this", "mana", "only", "to", "cast", "a", "creature", "spell",
	"of", "the", "chosen", "type",
}

var cantBeCounteredSpendEffectWords = []string{
	"and", "that", "spell", "can't", "be", "countered",
}

// recognizeManaSpendRider reports whether the sentence tokens are exactly the
// Path of Ancestry mana-spend rider "When that mana is spent to cast a creature
// spell that shares a creature type with your commander, scry N." and, if so,
// returns its typed syntax. It matches the entire token stream so that an
// unmodeled rider effect or trailing qualifier fails closed.
func recognizeManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	n := len(manaSpendRiderWords)
	// Layout: <n condition words> , scry <integer> [periods...]
	if len(tokens) < n+3 {
		return ManaSpendRiderSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, manaSpendRiderWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	if tokens[n].Kind != shared.Comma {
		return ManaSpendRiderSyntax{}, false
	}
	if !equalWord(tokens[n+1], "scry") {
		return ManaSpendRiderSyntax{}, false
	}
	amountToken := tokens[n+2]
	if amountToken.Kind != shared.Integer {
		return ManaSpendRiderSyntax{}, false
	}
	amount, err := strconv.Atoi(amountToken.Text)
	if err != nil || amount < 1 {
		return ManaSpendRiderSyntax{}, false
	}
	// Only trailing periods may follow the scry amount; any further word or
	// punctuation means extra unmodeled content, so fail closed.
	for i := n + 3; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:n]),
		EffectSpan:    shared.SpanOf(tokens[n+1 : n+3]),
		Condition:     ManaSpendCastCommanderCreatureType,
		Effect:        ManaSpendRiderEffectScry,
		ScryAmount:    amount,
	}, true
}

func recognizeChosenTypeManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	conditionEnd := len(chosenTypeManaSpendConditionWords)
	effectStart := conditionEnd + 1
	effectEnd := effectStart + len(cantBeCounteredSpendEffectWords)
	if len(tokens) <= effectEnd ||
		!effectWordsAt(tokens, 0, chosenTypeManaSpendConditionWords...) ||
		tokens[conditionEnd].Kind != shared.Comma ||
		!effectWordsAt(tokens, effectStart, cantBeCounteredSpendEffectWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	for i := effectEnd; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:conditionEnd]),
		EffectSpan:    shared.SpanOf(tokens[effectStart:effectEnd]),
		Condition:     ManaSpendCastChosenCreatureType,
		Effect:        ManaSpendRiderEffectCantBeCountered,
		Restricted:    true,
	}, true
}

// collapseManaSpendRiderSentence replaces a recognized mana-spend rider
// sentence's generic effects with a single typed EffectManaSpendRider effect
// that spans the whole sentence, so the rider rides on the preceding add-mana
// effect rather than splitting into uncoordinated cast/scry effects. It returns
// true when it collapsed the sentence. The synthesized effect credits the full
// sentence span for coverage and round-trips exactly.
func collapseManaSpendRiderSentence(sentence *Sentence, tokens []shared.Token) bool {
	rider, ok := recognizeManaSpendRider(tokens)
	if !ok {
		rider, ok = recognizeChosenTypeManaSpendRider(tokens)
		if !ok {
			return false
		}
	}
	span := shared.SpanOf(tokens)
	riderCopy := rider
	sentence.Effects = []EffectSyntax{{
		Kind:           EffectManaSpendRider,
		VerbSpan:       tokens[0].Span,
		ClauseSpan:     span,
		Span:           span,
		Text:           sentence.Text,
		Tokens:         tokens,
		Exact:          true,
		ManaSpendRider: &riderCopy,
	}}
	return true
}
