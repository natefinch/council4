package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// conditionalLookAtTopBattlefieldPutRun is the fixed token run that follows the
// variable card-type disjunction and the trailing "card" word: ", you may put
// it onto the battlefield". The optional "tapped" entry rider and the closing
// period are matched separately so both the plain and tapped entries are
// recognized.
var conditionalLookAtTopBattlefieldPutRun = []struct {
	kind shared.Kind
	text string
}{
	{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "may"},
	{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "onto"},
	{shared.Word, "the"}, {shared.Word, "battlefield"},
}

// conditionalLookAtTopBattlefieldHandRuns are the recognized trailing fallback
// clauses that send the card into the controller's hand when it is not put onto
// the battlefield. Each begins after the put clause's closing period and must
// consume the remainder of the body. The three wordings ("the card" / "it" in
// the conditional, and the "Otherwise," form) are equivalent: a mandatory move
// into the hand.
var conditionalLookAtTopBattlefieldHandRuns = [][]struct {
	kind shared.Kind
	text string
}{
	{
		{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
		{shared.Word, "put"}, {shared.Word, "the"}, {shared.Word, "card"},
		{shared.Word, "onto"}, {shared.Word, "the"}, {shared.Word, "battlefield"},
		{shared.Comma, ","}, {shared.Word, "put"}, {shared.Word, "it"},
		{shared.Word, "into"}, {shared.Word, "your"}, {shared.Word, "hand"},
		{shared.Period, "."},
	},
	{
		{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
		{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "onto"},
		{shared.Word, "the"}, {shared.Word, "battlefield"}, {shared.Comma, ","},
		{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "into"},
		{shared.Word, "your"}, {shared.Word, "hand"}, {shared.Period, "."},
	},
	{
		{shared.Word, "Otherwise"}, {shared.Comma, ","}, {shared.Word, "put"},
		{shared.Word, "it"}, {shared.Word, "into"}, {shared.Word, "your"},
		{shared.Word, "hand"}, {shared.Period, "."},
	},
}

// recognizeConditionalLookAtTopBattlefieldSequence matches the exact resolving
// body "look at the top card of your library. If it's a <type[ or type...]>
// card, you may put it onto the battlefield[ tapped]." optionally followed by a
// mandatory "if you don't put it onto the battlefield, put it into your hand."
// fallback. It records the disjunctive card types, the tapped entry rider, and
// the fallback disposition text-blind so the compiler and lowering stay free of
// Oracle wording. The whole body must match exactly, so any extra text fails
// closed.
func recognizeConditionalLookAtTopBattlefieldSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	if !matchTokenRun(tokens, &cursor, conditionalLookAtTopPrefix) {
		return false
	}
	cardTypes, next, ok := scanCardTypeDisjunction(tokens, cursor)
	if !ok {
		return false
	}
	cursor = next
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Word || tokens[cursor].Text != "card" {
		return false
	}
	cursor++
	if !matchTokenRun(tokens, &cursor, conditionalLookAtTopBattlefieldPutRun) {
		return false
	}
	entersTapped := false
	if cursor < len(tokens) && tokens[cursor].Kind == shared.Word && tokens[cursor].Text == "tapped" {
		entersTapped = true
		cursor++
	}
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Period || tokens[cursor].Text != "." {
		return false
	}
	cursor++

	elseDisposition := LookAtTopBattlefieldElseNone
	if cursor != len(tokens) {
		matched := false
		for _, run := range conditionalLookAtTopBattlefieldHandRuns {
			runCursor := cursor
			if matchTokenRun(tokens, &runCursor, run) && runCursor == len(tokens) {
				elseDisposition = LookAtTopBattlefieldElseHand
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:                  ExactSequenceConditionalLookAtTopBattlefield,
		Span:                  ability.BodySpan,
		LookAtTopCardTypes:    cardTypes,
		LookAtTopEntersTapped: entersTapped,
		LookAtTopBattlefield:  elseDisposition,
	}
	return true
}
