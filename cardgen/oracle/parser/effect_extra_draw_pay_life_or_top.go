package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// extraDrawPayLifeOrTopCardinals maps the spelled-out draw and choose counts of
// the ExactSequenceExtraDrawThenPayLifeOrTop body to their integer values.
var extraDrawPayLifeOrTopCardinals = map[string]int{
	"one":   1,
	"two":   2,
	"three": 3,
	"four":  4,
	"five":  5,
}

// extraDrawPayLifeOrTopMiddle is the fixed token run between the draw count and
// the choose count: "additional cards. If you do, choose".
var extraDrawPayLifeOrTopMiddle = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "additional"}, {shared.Word, "cards"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "do"},
	{shared.Comma, ","}, {shared.Word, "choose"},
}

// extraDrawPayLifeOrTopBeforeLife is the fixed token run between the choose count
// and the life amount: "cards in your hand drawn this turn. For each of those
// cards, pay".
var extraDrawPayLifeOrTopBeforeLife = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "cards"}, {shared.Word, "in"}, {shared.Word, "your"},
	{shared.Word, "hand"}, {shared.Word, "drawn"}, {shared.Word, "this"},
	{shared.Word, "turn"}, {shared.Period, "."}, {shared.Word, "For"},
	{shared.Word, "each"}, {shared.Word, "of"}, {shared.Word, "those"},
	{shared.Word, "cards"}, {shared.Comma, ","}, {shared.Word, "pay"},
}

// extraDrawPayLifeOrTopSuffix is the fixed token run after the life amount:
// "life or put the card on top of your library.".
var extraDrawPayLifeOrTopSuffix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "life"}, {shared.Word, "or"}, {shared.Word, "put"},
	{shared.Word, "the"}, {shared.Word, "card"}, {shared.Word, "on"},
	{shared.Word, "top"}, {shared.Word, "of"}, {shared.Word, "your"},
	{shared.Word, "library"}, {shared.Period, "."},
}

// recognizeExtraDrawPayLifeOrTopSequence matches the exact triggered resolving
// body "you may draw <N> additional cards. If you do, choose <M> cards in your
// hand drawn this turn. For each of those cards, pay <L> life or put the card on
// top of your library." (Sylvan Library). The "cards in your hand drawn this
// turn" identity choice and the per-card "pay life or put on top" disjunction
// are not expressible through the normal effect vocabulary, so the body is
// captured whole and lowered from the typed counts N, M, and L. It matches only
// a triggered ability body with no other content; the lowering separately gates
// on the draw-step trigger, so this recognizer never inspects the trigger words.
func recognizeExtraDrawPayLifeOrTopSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	prefix := []struct {
		kind shared.Kind
		text string
	}{
		{shared.Word, "you"}, {shared.Word, "may"}, {shared.Word, "draw"},
	}
	if !matchTokenRun(tokens, &cursor, prefix) {
		return false
	}
	drawCount, ok := matchExtraDrawPayLifeOrTopCardinal(tokens, &cursor)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &cursor, extraDrawPayLifeOrTopMiddle) {
		return false
	}
	chooseCount, ok := matchExtraDrawPayLifeOrTopCardinal(tokens, &cursor)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &cursor, extraDrawPayLifeOrTopBeforeLife) {
		return false
	}
	payLife, ok := matchExtraDrawPayLifeOrTopInteger(tokens, &cursor)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &cursor, extraDrawPayLifeOrTopSuffix) {
		return false
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:        ExactSequenceExtraDrawThenPayLifeOrTop,
		Span:        ability.BodySpan,
		DrawCount:   drawCount,
		ChooseCount: chooseCount,
		PayLife:     payLife,
	}
	return true
}

// matchExtraDrawPayLifeOrTopCardinal consumes a spelled-out cardinal (one..five)
// and reports its value.
func matchExtraDrawPayLifeOrTopCardinal(tokens []shared.Token, cursor *int) (int, bool) {
	if *cursor >= len(tokens) || tokens[*cursor].Kind != shared.Word {
		return 0, false
	}
	value, found := extraDrawPayLifeOrTopCardinals[tokens[*cursor].Text]
	if !found {
		return 0, false
	}
	*cursor++
	return value, true
}

// matchExtraDrawPayLifeOrTopInteger consumes a non-negative integer literal.
func matchExtraDrawPayLifeOrTopInteger(tokens []shared.Token, cursor *int) (int, bool) {
	if *cursor >= len(tokens) || tokens[*cursor].Kind != shared.Integer {
		return 0, false
	}
	value, err := strconv.Atoi(tokens[*cursor].Text)
	if err != nil || value < 0 {
		return 0, false
	}
	*cursor++
	return value, true
}
