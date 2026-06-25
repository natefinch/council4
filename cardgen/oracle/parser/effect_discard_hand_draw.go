package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// discardHandThenDrawTail is the fixed run that follows the whole-hand discard
// clause: ", then draw that many cards." The "that many" back-reference reads
// the number of cards the controller just discarded, a count the normal effect
// vocabulary cannot express, so the body is captured whole and lowered as a
// fixed instruction template.
var discardHandThenDrawTail = []struct {
	kind shared.Kind
	text string
}{
	{shared.Comma, ","}, {shared.Word, "then"}, {shared.Word, "draw"},
	{shared.Word, "that"}, {shared.Word, "many"}, {shared.Word, "cards"},
}

// discardHandThenDrawHeads are the recognized controller whole-hand discard
// lead-ins, matched before the shared ", then draw that many cards." tail. Both
// the terse "Discard your hand" and the verbose "Discard all the cards in your
// hand" spell the same effect: the controller discards every card in hand.
var discardHandThenDrawHeads = [][]struct {
	kind shared.Kind
	text string
}{
	{
		{shared.Word, "Discard"}, {shared.Word, "your"}, {shared.Word, "hand"},
	},
	{
		{shared.Word, "Discard"}, {shared.Word, "all"}, {shared.Word, "the"},
		{shared.Word, "cards"}, {shared.Word, "in"}, {shared.Word, "your"},
		{shared.Word, "hand"},
	},
}

// recognizeDiscardHandThenDrawSequence matches the spell body "Discard {your
// hand | all the cards in your hand}, then draw that many cards." (Decaying Time
// Loop) and records it as an exact sequence. The controller discards their whole
// hand and then draws a number of cards equal to the number discarded. It
// matches only a plain spell body with no trigger or cost.
func recognizeDiscardHandThenDrawSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilitySpell || ability.Trigger != nil || ability.CostSyntax != nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	for _, head := range discardHandThenDrawHeads {
		pos := 0
		if !matchTokenRun(tokens, &pos, head) {
			continue
		}
		if !matchTokenRun(tokens, &pos, discardHandThenDrawTail) {
			continue
		}
		if pos >= len(tokens) || tokens[pos].Kind != shared.Period {
			continue
		}
		pos++
		if pos != len(tokens) {
			continue
		}
		ability.ExactSequence = &ExactSequenceSyntax{
			Kind: ExactSequenceDiscardHandThenDraw,
			Span: ability.BodySpan,
		}
		return true
	}
	return false
}
