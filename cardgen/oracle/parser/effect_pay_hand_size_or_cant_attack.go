package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// payHandSizeOrCantAttackPrefix is the fixed token run that opens the Champions
// of Minas Tirith punisher body, up to the "{X}" mana symbol: "that opponent may
// pay". The intervening "if you're the monarch" clause is stripped before this
// run is matched.
var payHandSizeOrCantAttackPrefix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "that"}, {shared.Word, "opponent"}, {shared.Word, "may"},
	{shared.Word, "pay"},
}

// payHandSizeOrCantAttackSuffix is the fixed token run that follows the "{X}"
// mana symbol: ", where X is the number of cards in their hand. If they don't,
// they can't attack you this combat.".
var payHandSizeOrCantAttackSuffix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Comma, ","}, {shared.Word, "where"}, {shared.Word, "X"},
	{shared.Word, "is"}, {shared.Word, "the"}, {shared.Word, "number"},
	{shared.Word, "of"}, {shared.Word, "cards"}, {shared.Word, "in"},
	{shared.Word, "their"}, {shared.Word, "hand"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "they"}, {shared.Word, "don't"},
	{shared.Comma, ","}, {shared.Word, "they"}, {shared.Word, "can't"},
	{shared.Word, "attack"}, {shared.Word, "you"}, {shared.Word, "this"},
	{shared.Word, "combat"}, {shared.Period, "."},
}

// recognizePayHandSizeOrCantAttackSequence matches the exact triggered resolving
// body "that opponent may pay {X}, where X is the number of cards in their hand.
// If they don't, they can't attack you this combat." (Champions of Minas
// Tirith), following an intervening-if condition. On a match it records the
// ExactSequencePayHandSizeOrCantAttack syntax so lowering models the punisher
// from typed nodes rather than Oracle wording.
func recognizePayHandSizeOrCantAttackSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered {
		return false
	}
	strip := ability.interveningConditionStrip()
	if !strip.set {
		return false
	}
	tokens := strip.strip(tokensWithinParserSpan(ability.Tokens, ability.BodySpan))
	cursor := 0
	if !matchTokenRun(tokens, &cursor, payHandSizeOrCantAttackPrefix) {
		return false
	}
	mana, next, ok := parseKeywordManaCost(tokens, cursor)
	if !ok || len(mana) != 1 || mana[0] != cost.X {
		return false
	}
	cursor = next
	if !matchTokenRun(tokens, &cursor, payHandSizeOrCantAttackSuffix) {
		return false
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind: ExactSequencePayHandSizeOrCantAttack,
		Span: ability.BodySpan,
	}
	return true
}
