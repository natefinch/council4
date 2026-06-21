package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// bottomHandThenDrawPrefix is the fixed lead-in for the hand-to-library + draw
// sequence up to (but excluding) the library-end word ("bottom" or "top").
var bottomHandThenDrawPrefix = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "Put"}, {shared.Word, "any"}, {shared.Word, "number"},
	{shared.Word, "of"}, {shared.Word, "cards"}, {shared.Word, "from"},
	{shared.Word, "your"}, {shared.Word, "hand"}, {shared.Word, "on"},
}

// bottomHandThenDrawMiddle is the fixed run from the library-of clause through
// "draw that many cards"; the optional "plus <n>" rider and the final period
// follow it.
var bottomHandThenDrawMiddle = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "of"}, {shared.Word, "your"}, {shared.Word, "library"},
	{shared.Comma, ","}, {shared.Word, "then"}, {shared.Word, "draw"},
	{shared.Word, "that"}, {shared.Word, "many"}, {shared.Word, "cards"},
}

var bottomHandThenDrawOffsetWords = map[string]int{
	"one":   1,
	"two":   2,
	"three": 3,
	"four":  4,
	"five":  5,
}

// recognizeBottomHandThenDrawSequence matches the spell body "Put any number of
// cards from your hand on the {bottom|top} of your library, then draw that many
// cards[ plus <n>]." and records it as an exact sequence. The "draw that many"
// back-reference and player-chosen count are not expressible through the normal
// effect vocabulary, so the body is captured whole and lowered as a fixed
// instruction template. It matches only a plain spell body with no trigger or
// cost.
func recognizeBottomHandThenDrawSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilitySpell || ability.Trigger != nil || ability.CostSyntax != nil {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	pos := 0
	if !matchTokenRun(tokens, &pos, bottomHandThenDrawPrefix) {
		return false
	}
	// Optional "the" before the library-end word ("on the bottom" / "on top").
	if pos < len(tokens) && tokens[pos].Kind == shared.Word && tokens[pos].Text == "the" {
		pos++
	}
	bottom, ok := matchBottomHandThenDrawEnd(tokens, &pos)
	if !ok {
		return false
	}
	if !matchTokenRun(tokens, &pos, bottomHandThenDrawMiddle) {
		return false
	}
	offset := 0
	if pos < len(tokens) && tokens[pos].Kind == shared.Word && tokens[pos].Text == "plus" {
		pos++
		value, valueOK := matchBottomHandThenDrawOffset(tokens, &pos)
		if !valueOK {
			return false
		}
		offset = value
	}
	if pos >= len(tokens) || tokens[pos].Kind != shared.Period {
		return false
	}
	pos++
	if pos != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:       ExactSequenceBottomHandThenDraw,
		Span:       ability.BodySpan,
		Bottom:     bottom,
		DrawOffset: offset,
	}
	return true
}

// matchBottomHandThenDrawEnd consumes the library-end word and reports whether
// it is the bottom of the library (bottom) and whether a known end word matched
// (matched).
func matchBottomHandThenDrawEnd(tokens []shared.Token, pos *int) (bottom, matched bool) {
	if *pos >= len(tokens) || tokens[*pos].Kind != shared.Word {
		return false, false
	}
	switch tokens[*pos].Text {
	case "bottom":
		*pos++
		return true, true
	case "top":
		*pos++
		return false, true
	default:
		return false, false
	}
}

// matchBottomHandThenDrawOffset consumes the spelled-out "plus <n>" count.
func matchBottomHandThenDrawOffset(tokens []shared.Token, pos *int) (int, bool) {
	if *pos >= len(tokens) || tokens[*pos].Kind != shared.Word {
		return 0, false
	}
	value, found := bottomHandThenDrawOffsetWords[tokens[*pos].Text]
	if !found {
		return 0, false
	}
	*pos++
	return value, true
}

// matchTokenRun matches a fixed run of tokens at pos, advancing pos past it on
// success.
func matchTokenRun(tokens []shared.Token, pos *int, run []struct {
	kind shared.Kind
	text string
}) bool {
	if *pos+len(run) > len(tokens) {
		return false
	}
	for i, want := range run {
		token := tokens[*pos+i]
		if token.Kind != want.kind || token.Text != want.text {
			return false
		}
	}
	*pos += len(run)
	return true
}
