package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// devotionLookWinPrefix is the fixed token run that precedes the variable
// devotion color word: "look at the top X cards of your library , where X is
// your devotion to".
var devotionLookWinPrefix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "look"}, {shared.Word, "at"}, {shared.Word, "the"},
	{shared.Word, "top"}, {shared.Word, "X"}, {shared.Word, "cards"},
	{shared.Word, "of"}, {shared.Word, "your"}, {shared.Word, "library"},
	{shared.Comma, ","}, {shared.Word, "where"}, {shared.Word, "X"},
	{shared.Word, "is"}, {shared.Word, "your"}, {shared.Word, "devotion"},
	{shared.Word, "to"},
}

// devotionLookWinSuffix is the fixed token run that follows the devotion color
// word: ". Put up to one of them on top of your library and the rest on the
// bottom of your library in a random order. If X is greater than or equal to
// the number of cards in your library , you win the game .".
var devotionLookWinSuffix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Period, "."},
	{shared.Word, "Put"}, {shared.Word, "up"}, {shared.Word, "to"},
	{shared.Word, "one"}, {shared.Word, "of"}, {shared.Word, "them"},
	{shared.Word, "on"}, {shared.Word, "top"}, {shared.Word, "of"},
	{shared.Word, "your"}, {shared.Word, "library"}, {shared.Word, "and"},
	{shared.Word, "the"}, {shared.Word, "rest"}, {shared.Word, "on"},
	{shared.Word, "the"}, {shared.Word, "bottom"}, {shared.Word, "of"},
	{shared.Word, "your"}, {shared.Word, "library"}, {shared.Word, "in"},
	{shared.Word, "a"}, {shared.Word, "random"}, {shared.Word, "order"},
	{shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "X"}, {shared.Word, "is"},
	{shared.Word, "greater"}, {shared.Word, "than"}, {shared.Word, "or"},
	{shared.Word, "equal"}, {shared.Word, "to"}, {shared.Word, "the"},
	{shared.Word, "number"}, {shared.Word, "of"}, {shared.Word, "cards"},
	{shared.Word, "in"}, {shared.Word, "your"}, {shared.Word, "library"},
	{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "win"},
	{shared.Word, "the"}, {shared.Word, "game"}, {shared.Period, "."},
}

// recognizeDevotionLookWinSequence matches Thassa's Oracle's exact resolving
// body "look at the top X cards of your library, where X is your devotion to
// <color>. Put up to one of them on top of your library and the rest on the
// bottom of your library in a random order. If X is greater than or equal to
// the number of cards in your library, you win the game." text-blind, recording
// the single devotion color so the compiler and lowering stay free of Oracle
// wording. Any trailing reminder text (the parenthetical devotion reminder) is
// excluded before matching, and the whole remaining body must match exactly, so
// any extra text fails closed.
func recognizeDevotionLookWinSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := bodyTokensExcludingReminders(ability)
	cursor := 0
	for _, want := range devotionLookWinPrefix {
		if cursor >= len(tokens) || tokens[cursor].Kind != want.kind || tokens[cursor].Text != want.text {
			return false
		}
		cursor++
	}
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Word {
		return false
	}
	color, ok := recognizeColorWord(tokens[cursor].Text)
	if !ok {
		return false
	}
	cursor++
	for _, want := range devotionLookWinSuffix {
		if cursor >= len(tokens) || tokens[cursor].Kind != want.kind || tokens[cursor].Text != want.text {
			return false
		}
		cursor++
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:          ExactSequenceDevotionLookWin,
		Span:          ability.BodySpan,
		DevotionColor: color,
	}
	return true
}

// bodyTokensExcludingReminders returns the ability's body-span tokens with every
// token that falls inside a reminder-text delimiter removed. Reminder text
// carries no game meaning, so an exact-sequence recognizer matches the semantic
// body without the trailing parenthetical reminder.
func bodyTokensExcludingReminders(ability *Ability) []shared.Token {
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	if len(ability.Reminders) == 0 {
		return tokens
	}
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		inReminder := false
		for _, reminder := range ability.Reminders {
			if token.Span.Start.Offset >= reminder.Span.Start.Offset &&
				token.Span.End.Offset <= reminder.Span.End.Offset {
				inReminder = true
				break
			}
		}
		if !inReminder {
			result = append(result, token)
		}
	}
	return result
}
