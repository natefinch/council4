package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// escalateHeader recognizes an "Escalate <cost>" keyword header (CR 702.121),
// the paragraph printed above an ordinary modal that lets the controller choose
// extra modes by paying the escalate cost once per mode beyond the first. It
// returns the parsed mana cost and the header's source span. The header is the
// word "Escalate" followed by a mana cost and then only reminder text, e.g.
// "Escalate {G} (Pay this cost for each mode chosen beyond the first.)".
func escalateHeader(tokens []shared.Token) (cost.Mana, shared.Span, bool) {
	if !startsWithWord(tokens, "Escalate") {
		return nil, shared.Span{}, false
	}
	manaCost, end, ok := parseKeywordManaCost(tokens, 1)
	if !ok {
		return nil, shared.Span{}, false
	}
	for _, token := range tokensOutsideParens(tokens[end:]) {
		if token.Kind != shared.Period {
			return nil, shared.Span{}, false
		}
	}
	return manaCost, shared.SpanOf(tokens[:end]), true
}

// nextNonEmptyLineIsModalHeader reports whether the first non-empty line after
// index i begins a modal choose header. An Escalate keyword header is folded
// into the following modal ability only when one is present; otherwise the
// Escalate line is left to fail closed as an unrecognized ability.
func nextNonEmptyLineIsModalHeader(lines [][]shared.Token, i int) bool {
	for j := i + 1; j < len(lines); j++ {
		if len(lines[j]) == 0 {
			continue
		}
		return modalHeaderStart(lines[j]) >= 0
	}
	return false
}
