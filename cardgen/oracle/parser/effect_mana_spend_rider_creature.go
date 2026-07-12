package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeCreatureCastOrActivateManaSpendRider reports whether the sentence
// tokens are the restricted "Spend this mana only to cast <creature spell> or
// activate <ability of a creature>" restriction (Castle Garenbrig) and, if so,
// returns its typed syntax. It shares the "Spend this mana only [to] cast"
// prefix with the other spell-type restrictions and then requires exactly the
// creature-spell selector followed by the creature-activation clause:
//
//   - the spell selector is "a creature spell" or "creature spells".
//   - the activation clause is "or activate" followed by "an ability of a
//     creature" (optionally "source") or "abilities of creatures" /
//     "abilities of creature sources".
//
// The two "creature" and "creature source" phrasings mean the same thing (an
// ability whose source is a creature permanent), so both map to the single
// ManaSpendCastOrActivateCreature condition. Any other selector, qualifier, or
// trailing content fails closed.
func recognizeCreatureCastOrActivateManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
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
	if index < len(tokens) && equalWord(tokens[index], "a") {
		index++
	}
	if index >= len(tokens) || !equalWord(tokens[index], "creature") {
		return ManaSpendRiderSyntax{}, false
	}
	nounEnd, ok := spellNounEnd(tokens, index+1)
	if !ok {
		return ManaSpendRiderSyntax{}, false
	}
	if !effectWordsAt(tokens, nounEnd, "or", "activate") {
		return ManaSpendRiderSyntax{}, false
	}
	clauseEnd, ok := creatureActivateClauseEnd(tokens, nounEnd+2)
	if !ok {
		return ManaSpendRiderSyntax{}, false
	}
	for i := clauseEnd; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:clauseEnd]),
		Condition:     ManaSpendCastOrActivateCreature,
		Effect:        ManaSpendRiderEffectUnknown,
		Restricted:    true,
	}, true
}

// creatureActivateClauseEnd matches the creature-activation clause of the
// cast-or-activate restriction beginning at start and returns the index just
// past it. It accepts the singular "an ability of a creature[ source]" and the
// plural "abilities of creatures" / "abilities of creature sources". Any other
// wording fails closed.
func creatureActivateClauseEnd(tokens []shared.Token, start int) (int, bool) {
	if effectWordsAt(tokens, start, "an", "ability", "of", "a", "creature") {
		end := start + 5
		if end < len(tokens) && equalWord(tokens[end], "source") {
			end++
		}
		return end, true
	}
	if effectWordsAt(tokens, start, "abilities", "of") {
		rest := start + 2
		if rest < len(tokens) && equalWord(tokens[rest], "creatures") {
			return rest + 1, true
		}
		if effectWordsAt(tokens, rest, "creature", "sources") {
			return rest + 2, true
		}
	}
	return 0, false
}
