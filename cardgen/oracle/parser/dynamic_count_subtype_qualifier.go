package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func dynamicCountSubtypeQualifierEnd(tokens []shared.Token, start int, atoms Atoms) (int, bool) {
	var next int
	switch {
	case effectWordsAt(tokens, start, "that's", "a"):
		next = start + 2
	case effectWordsAt(tokens, start, "that", "is", "a"):
		next = start + 3
	default:
		return 0, false
	}
	for {
		if next >= len(tokens) {
			return 0, false
		}
		if _, ok := atoms.SubtypeAt(tokens[next].Span); !ok {
			return 0, false
		}
		next++
		if dynamicAmountBoundary(tokens, next) {
			return next, true
		}
		if !effectWordsAt(tokens, next, "or", "a") {
			return 0, false
		}
		next += 2
	}
}
