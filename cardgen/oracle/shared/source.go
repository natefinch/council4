package shared

import (
	"strings"
)

// SpanOf returns the span covering tokens.
func SpanOf(tokens []Token) Span {
	if len(tokens) == 0 {
		return Span{}
	}
	return Span{Start: tokens[0].Span.Start, End: tokens[len(tokens)-1].Span.End}
}

// SliceSpan returns the source text covered by span, or an empty string when
// span is outside source.
func SliceSpan(source string, span Span) string {
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		return ""
	}
	return source[span.Start.Offset:span.End.Offset]
}

// TopLevelIndex returns the first wanted token outside parentheses and quotes.
func TopLevelIndex(tokens []Token, wanted Kind) int {
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if !quoted {
				depth++
			}
		case RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case Quote:
			quoted = !quoted
		default:
			if token.Kind == wanted && depth == 0 && !quoted {
				return i
			}
		}
	}
	return -1
}

// NormalizedWords returns the lower-cased text of word tokens.
func NormalizedWords(tokens []Token) []string {
	words := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == Word {
			words = append(words, strings.ToLower(token.Text))
		}
	}
	return words
}
