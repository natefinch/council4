package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// drawThenDiscardUnlessMiddle is the fixed run between the drawn-card count and
// the discarded-card count: "cards . Then discard".
var drawThenDiscardUnlessMiddle = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "cards"}, {shared.Period, "."}, {shared.Word, "Then"},
	{shared.Word, "discard"},
}

// drawThenDiscardUnlessTail is the fixed run between the discarded-card count
// and the variable exempt card-type disjunction: "cards unless you discard a".
var drawThenDiscardUnlessTail = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "cards"}, {shared.Word, "unless"}, {shared.Word, "you"},
	{shared.Word, "discard"},
}

// recognizeDrawThenDiscardUnlessSequence matches the spell body "Draw N cards.
// Then discard M cards unless you discard a <type[ or type...]> card." (Thirst
// for Knowledge family) text-blind: the controller draws N cards, then discards
// M cards unless they discard a single card of one of the recorded exempt
// types. The counts and exempt types travel on the compiled ability so the
// compiler and lowering never read Oracle words. It matches only a plain spell
// body with no trigger or cost, failing closed on any extra text.
func recognizeDrawThenDiscardUnlessSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilitySpell || ability.Trigger != nil || ability.CostSyntax != nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	pos := 0
	if pos >= len(tokens) || tokens[pos].Kind != shared.Word || tokens[pos].Text != "Draw" {
		return false
	}
	pos++
	drawCount, ok := cardinalCount(tokens, &pos)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &pos, drawThenDiscardUnlessMiddle) {
		return false
	}
	discardCount, ok := cardinalCount(tokens, &pos)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &pos, drawThenDiscardUnlessTail) {
		return false
	}
	if pos >= len(tokens) || tokens[pos].Kind != shared.Word || (tokens[pos].Text != "a" && tokens[pos].Text != "an") {
		return false
	}
	pos++
	cardTypes, next, ok := scanCardTypeDisjunction(tokens, pos)
	if !ok {
		return false
	}
	pos = next
	if pos >= len(tokens) || tokens[pos].Kind != shared.Word || tokens[pos].Text != "card" {
		return false
	}
	pos++
	if pos >= len(tokens) || tokens[pos].Kind != shared.Period {
		return false
	}
	pos++
	if pos != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:               ExactSequenceDrawThenDiscardUnlessType,
		Span:               ability.BodySpan,
		DrawCount:          drawCount,
		DiscardCount:       discardCount,
		LookAtTopCardTypes: cardTypes,
	}
	return true
}

// cardinalCount reads a spelled-out cardinal number word at pos, advancing pos
// past it on success.
func cardinalCount(tokens []shared.Token, pos *int) (int, bool) {
	if *pos >= len(tokens) || tokens[*pos].Kind != shared.Word {
		return 0, false
	}
	value, ok := CardinalWordValue(tokens[*pos].Text)
	if !ok {
		return 0, false
	}
	*pos++
	return value, true
}
