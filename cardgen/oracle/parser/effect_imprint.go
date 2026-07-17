package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

var replaceLinkedExiledCardBody = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "you"}, {shared.Word, "may"}, {shared.Word, "exile"},
	{shared.Word, "that"}, {shared.Word, "card"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "do"},
	{shared.Comma, ","}, {shared.Word, "return"}, {shared.Word, "each"},
	{shared.Word, "other"}, {shared.Word, "card"}, {shared.Word, "exiled"},
	{shared.Word, "with"}, {shared.Word, "this"}, {shared.Word, "artifact"},
	{shared.Word, "to"}, {shared.Word, "its"}, {shared.Word, "owner's"},
	{shared.Word, "graveyard"}, {shared.Period, "."},
}

var linkedExiledCopyTokenBody = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "Create"}, {shared.Word, "a"}, {shared.Word, "token"},
	{shared.Word, "that's"}, {shared.Word, "a"}, {shared.Word, "copy"},
	{shared.Word, "of"}, {shared.Word, "a"}, {shared.Word, "card"},
	{shared.Word, "exiled"}, {shared.Word, "with"}, {shared.Word, "this"},
	{shared.Word, "artifact"}, {shared.Period, "."},
	{shared.Word, "It"}, {shared.Word, "gains"}, {shared.Word, "haste"},
	{shared.Period, "."},
	{shared.Word, "Exile"}, {shared.Word, "it"}, {shared.Word, "at"},
	{shared.Word, "the"}, {shared.Word, "beginning"}, {shared.Word, "of"},
	{shared.Word, "the"}, {shared.Word, "next"}, {shared.Word, "end"},
	{shared.Word, "step"}, {shared.Period, "."},
}

func recognizeReplaceLinkedExiledCardSequence(ability *Ability) bool {
	return recognizeImprintSequence(
		ability,
		AbilityTriggered,
		replaceLinkedExiledCardBody,
		ExactSequenceReplaceLinkedExiledCard,
	)
}

func recognizeLinkedExiledCopyTokenSequence(ability *Ability) bool {
	return recognizeImprintSequence(
		ability,
		AbilityActivated,
		linkedExiledCopyTokenBody,
		ExactSequenceLinkedExiledCopyToken,
	)
}

func recognizeImprintSequence(
	ability *Ability,
	kind AbilityKind,
	body []struct {
		kind shared.Kind
		text string
	},
	sequence ExactSequenceKind,
) bool {
	if ability == nil || ability.Kind != kind {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	if !matchTokenRun(tokens, &cursor, body) || cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{Kind: sequence, Span: ability.BodySpan}
	return true
}
