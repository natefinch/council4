package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// monocoloredChosenColorSpendWords is the selector tail of the Throne of Eldraine
// mana-spend restriction "Spend this mana only to cast monocolored spells of that
// color." It follows the shared "spend this mana only to cast" prefix.
var monocoloredChosenColorSpendWords = []string{
	"monocolored", "spells", "of", "that", "color",
}

// recognizeMonocoloredChosenColorManaSpendRider reports whether the sentence
// tokens are exactly "Spend this mana only to cast monocolored spells of that
// color." (Throne of Eldraine) and, if so, returns its typed syntax. The tagged
// mana is produced in the source's entry-time chosen color, and "that color"
// refers back to that same color; the restriction admits only a monocolored
// spell whose single color matches. Any other selector or trailing content fails
// closed.
func recognizeMonocoloredChosenColorManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	prefix := len(spellTypeSpendPrefixWords)
	if len(tokens) <= prefix || !effectWordsAt(tokens, 0, spellTypeSpendPrefixWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	index := prefix
	if index < len(tokens) && equalWord(tokens[index], "to") {
		index++
	}
	if index >= len(tokens) || !equalWord(tokens[index], "cast") {
		return ManaSpendRiderSyntax{}, false
	}
	index++
	selectorEnd := index + len(monocoloredChosenColorSpendWords)
	if len(tokens) < selectorEnd ||
		!effectWordsAt(tokens, index, monocoloredChosenColorSpendWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	for i := selectorEnd; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:selectorEnd]),
		Condition:     ManaSpendCastMonocoloredSpellOfChosenColor,
		Effect:        ManaSpendRiderEffectUnknown,
		Restricted:    true,
	}, true
}
