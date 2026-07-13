package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// beseechSearchCastPayoffPrefix is the fixed token run of the Beseech-the-Mirror
// spell body up to (but excluding) the "<N> or less" mana-value bound: "Search
// your library for a card, exile it face down, then shuffle. If this spell was
// bargained, you may cast the exiled card without paying its mana cost if that
// spell's mana value is".
var beseechSearchCastPayoffPrefix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "Search"}, {shared.Word, "your"}, {shared.Word, "library"},
	{shared.Word, "for"}, {shared.Word, "a"}, {shared.Word, "card"},
	{shared.Comma, ","}, {shared.Word, "exile"}, {shared.Word, "it"},
	{shared.Word, "face"}, {shared.Word, "down"}, {shared.Comma, ","},
	{shared.Word, "then"}, {shared.Word, "shuffle"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "this"}, {shared.Word, "spell"},
	{shared.Word, "was"}, {shared.Word, "bargained"}, {shared.Comma, ","},
	{shared.Word, "you"}, {shared.Word, "may"}, {shared.Word, "cast"},
	{shared.Word, "the"}, {shared.Word, "exiled"}, {shared.Word, "card"},
	{shared.Word, "without"}, {shared.Word, "paying"}, {shared.Word, "its"},
	{shared.Word, "mana"}, {shared.Word, "cost"}, {shared.Word, "if"},
	{shared.Word, "that"}, {shared.Word, "spell's"}, {shared.Word, "mana"},
	{shared.Word, "value"}, {shared.Word, "is"},
}

// beseechSearchCastPayoffSuffix is the fixed token run after the mana-value bound:
// "or less. Put the exiled card into your hand if it wasn't cast this way.".
var beseechSearchCastPayoffSuffix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "or"}, {shared.Word, "less"}, {shared.Period, "."},
	{shared.Word, "Put"}, {shared.Word, "the"}, {shared.Word, "exiled"},
	{shared.Word, "card"}, {shared.Word, "into"}, {shared.Word, "your"},
	{shared.Word, "hand"}, {shared.Word, "if"}, {shared.Word, "it"},
	{shared.Word, "wasn't"}, {shared.Word, "cast"}, {shared.Word, "this"},
	{shared.Word, "way"}, {shared.Period, "."},
}

// recognizeBargainSearchCastPayoffSequence matches the exact spell body "Search
// your library for a card, exile it face down, then shuffle. If this spell was
// bargained, you may cast the exiled card without paying its mana cost if that
// spell's mana value is <N> or less. Put the exiled card into your hand if it
// wasn't cast this way." (Beseech the Mirror) and records it as an exact
// sequence. The cast-time bargained gate, the face-down linked exile, the free
// cast of the exiled card, and the "if it wasn't cast this way" fallback compose
// several primitives whose linkage the normal effect vocabulary cannot express,
// so the body is captured whole and lowered from the typed mana-value bound. It
// matches only a plain spell body with no trigger or cost.
func recognizeBargainSearchCastPayoffSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilitySpell || ability.Trigger != nil || ability.CostSyntax != nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	if !matchTokenRun(tokens, &cursor, beseechSearchCastPayoffPrefix) {
		return false
	}
	maxManaValue, ok := matchBeseechManaValueBound(tokens, &cursor)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &cursor, beseechSearchCastPayoffSuffix) {
		return false
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:         ExactSequenceBargainSearchCastPayoff,
		Span:         ability.BodySpan,
		MaxManaValue: maxManaValue,
	}
	return true
}

// matchBeseechManaValueBound consumes the "<N>" mana-value bound as a
// non-negative integer literal.
func matchBeseechManaValueBound(tokens []shared.Token, cursor *int) (int, bool) {
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
