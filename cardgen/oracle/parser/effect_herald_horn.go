package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

var chosenTypeLibraryTopTokenSequence = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "look"}, {shared.Word, "at"}, {shared.Word, "the"},
	{shared.Word, "top"}, {shared.Word, "card"}, {shared.Word, "of"},
	{shared.Word, "your"}, {shared.Word, "library"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "it's"}, {shared.Word, "a"},
	{shared.Word, "creature"}, {shared.Word, "card"}, {shared.Word, "of"},
	{shared.Word, "the"}, {shared.Word, "chosen"}, {shared.Word, "type"},
	{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "may"},
	{shared.Word, "reveal"}, {shared.Word, "it"}, {shared.Word, "and"},
	{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "into"},
	{shared.Word, "your"}, {shared.Word, "hand"}, {shared.Period, "."},
}

func recognizeChosenTypeLibraryTopSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	if len(tokens) != len(chosenTypeLibraryTopTokenSequence) {
		return false
	}
	for i, want := range chosenTypeLibraryTopTokenSequence {
		if tokens[i].Kind != want.kind || tokens[i].Text != want.text {
			return false
		}
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind: ExactSequenceChosenTypeLibraryTopToHand,
		Span: ability.BodySpan,
	}
	return true
}
