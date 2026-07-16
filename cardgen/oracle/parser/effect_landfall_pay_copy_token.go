package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeLandfallPayCopyTokenSequence matches Springheart Nantuko's exact
// triggered resolving body:
//
//	you may pay {1}{G} if this permanent is attached to a creature you control.
//	If you do, create a token that's a copy of that creature. If you didn't
//	create a token this way, create a 1/1 green Insect creature token.
//
// The whole body is fixed, so it carries no extra data: lowering composes the
// gated optional payment, the reflexive token-copy of the attached creature, and
// the fixed 1/1 green Insect fallback from the recognized kind alone. Every token
// including the {1}{G} cost and the 1/1 green Insect wording is matched exactly,
// so any near-miss (a different cost, a different token, a missing branch, a
// non-landfall trigger) fails closed and leaves the body for the generic path.
func recognizeLandfallPayCopyTokenSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := bodyTokensExcludingReminders(ability)
	cursor := 0
	takeWord := func(word string) bool {
		if cursor >= len(tokens) || !equalWord(tokens[cursor], word) {
			return false
		}
		cursor++
		return true
	}
	takeWords := func(words ...string) bool {
		for _, word := range words {
			if !takeWord(word) {
				return false
			}
		}
		return true
	}
	takeKind := func(kind shared.Kind) bool {
		if cursor >= len(tokens) || tokens[cursor].Kind != kind {
			return false
		}
		cursor++
		return true
	}
	takeSymbol := func(text string) bool {
		if cursor >= len(tokens) || tokens[cursor].Kind != shared.Symbol || tokens[cursor].Text != text {
			return false
		}
		cursor++
		return true
	}
	takeInteger := func(text string) bool {
		if cursor >= len(tokens) || tokens[cursor].Kind != shared.Integer || tokens[cursor].Text != text {
			return false
		}
		cursor++
		return true
	}
	// "you may pay {1}{G} if this permanent is attached to a creature you control."
	if !takeWords("you", "may", "pay") || !takeSymbol("{1}") || !takeSymbol("{G}") {
		return false
	}
	if !takeWords("if", "this", "permanent", "is", "attached", "to", "a", "creature", "you", "control") ||
		!takeKind(shared.Period) {
		return false
	}
	// "If you do, create a token that's a copy of that creature."
	if !takeWords("if", "you", "do") || !takeKind(shared.Comma) {
		return false
	}
	if !takeWords("create", "a", "token", "that's", "a", "copy", "of", "that", "creature") ||
		!takeKind(shared.Period) {
		return false
	}
	// "If you didn't create a token this way, create a 1/1 green Insect creature token."
	if !takeWords("if", "you", "didn't", "create", "a", "token", "this", "way") || !takeKind(shared.Comma) {
		return false
	}
	if !takeWords("create", "a") ||
		!takeInteger("1") || !takeKind(shared.Slash) || !takeInteger("1") ||
		!takeWords("green", "Insect", "creature", "token") ||
		!takeKind(shared.Period) {
		return false
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind: ExactSequenceLandfallPayCopyToken,
		Span: ability.BodySpan,
	}
	return true
}
