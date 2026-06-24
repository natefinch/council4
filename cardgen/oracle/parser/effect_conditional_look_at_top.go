package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// conditionalLookAtTopPrefix is the fixed token run that precedes the variable
// card-type disjunction in the conditional look-at-top reveal sequence:
// "look at the top card of your library . If it's a".
var conditionalLookAtTopPrefix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "look"}, {shared.Word, "at"}, {shared.Word, "the"},
	{shared.Word, "top"}, {shared.Word, "card"}, {shared.Word, "of"},
	{shared.Word, "your"}, {shared.Word, "library"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "it's"}, {shared.Word, "a"},
}

// conditionalLookAtTopSuffix is the fixed token run that follows the variable
// card-type disjunction: "card , you may reveal it and put it into your hand .
// If you don't put the card into your hand , you may put it into your
// graveyard .".
var conditionalLookAtTopSuffix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "card"}, {shared.Comma, ","}, {shared.Word, "you"},
	{shared.Word, "may"}, {shared.Word, "reveal"}, {shared.Word, "it"},
	{shared.Word, "and"}, {shared.Word, "put"}, {shared.Word, "it"},
	{shared.Word, "into"}, {shared.Word, "your"}, {shared.Word, "hand"},
	{shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
	{shared.Word, "put"}, {shared.Word, "the"}, {shared.Word, "card"},
	{shared.Word, "into"}, {shared.Word, "your"}, {shared.Word, "hand"},
	{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "may"},
	{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "into"},
	{shared.Word, "your"}, {shared.Word, "graveyard"}, {shared.Period, "."},
}

// recognizeConditionalLookAtTopSequence matches the exact resolving body
// "look at the top card of your library. If it's a <type[ or type...]> card,
// you may reveal it and put it into your hand. If you don't put the card into
// your hand, you may put it into your graveyard." text-blind, recording the
// disjunctive card types so the compiler and lowering stay free of Oracle
// wording. The whole body must match exactly, so any extra text fails closed.
func recognizeConditionalLookAtTopSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	for _, want := range conditionalLookAtTopPrefix {
		if cursor >= len(tokens) || tokens[cursor].Kind != want.kind || tokens[cursor].Text != want.text {
			return false
		}
		cursor++
	}
	cardTypes, next, ok := scanCardTypeDisjunction(tokens, cursor)
	if !ok {
		return false
	}
	cursor = next
	for _, want := range conditionalLookAtTopSuffix {
		if cursor >= len(tokens) || tokens[cursor].Kind != want.kind || tokens[cursor].Text != want.text {
			return false
		}
		cursor++
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:               ExactSequenceConditionalLookAtTopReveal,
		Span:               ability.BodySpan,
		LookAtTopCardTypes: cardTypes,
	}
	return true
}

// scanCardTypeDisjunction reads one or more card-type words joined by "or"
// starting at start, returning the recognized types and the index just past
// the run. It stops before the trailing "card" word that the suffix matches.
func scanCardTypeDisjunction(tokens []shared.Token, start int) ([]CardType, int, bool) {
	cursor := start
	var cardTypes []CardType
	for {
		if cursor >= len(tokens) || tokens[cursor].Kind != shared.Word {
			return nil, 0, false
		}
		cardType, ok := recognizeCardTypeWord(tokens[cursor].Text)
		if !ok {
			return nil, 0, false
		}
		cardTypes = append(cardTypes, cardType)
		cursor++
		if cursor < len(tokens) && tokens[cursor].Kind == shared.Word && tokens[cursor].Text == "or" {
			cursor++
			continue
		}
		return cardTypes, cursor, true
	}
}
