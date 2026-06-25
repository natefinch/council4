package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// SpreeCostClause is a Spree option's additional mana cost and the em dash
// separating it from the option's rules text (CR 702.171). It mirrors
// ModeLabelClause so the lowering can credit the cost's source span when it
// asserts that a modal option's text is fully recognized.
type SpreeCostClause struct {
	Cost          cost.Mana   `json:",omitempty"`
	Text          string      `json:",omitempty"`
	Span          shared.Span `json:"-"`
	SeparatorSpan shared.Span `json:"-"`
}

// isSpreeHeader reports whether a paragraph is a Spree keyword header. A Spree
// header is the word "Spree" followed only by reminder text, e.g. "Spree
// (Choose one or more additional costs.)".
func isSpreeHeader(tokens []shared.Token) bool {
	if !startsWithWord(tokens, "Spree") {
		return false
	}
	for _, token := range tokensOutsideParens(tokens[1:]) {
		if token.Kind != shared.Period {
			return false
		}
	}
	return true
}

// parseSpreeMode parses one "+ {cost} — effect" Spree option. The leading "+"
// has already been removed by the caller. The mana cost and its em dash are
// recorded as a SpreeCostClause; the remaining tokens are the option's rules
// text, parsed exactly like an ordinary modal option body.
func parseSpreeMode(source string, tokens []shared.Token) (Mode, []shared.Diagnostic) {
	mode := Mode{
		Span:   shared.SpanOf(tokens),
		Text:   shared.SliceSpan(source, shared.SpanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	bodyTokens := tokens
	if manaCost, end, ok := parseKeywordManaCost(tokens, 0); ok &&
		end < len(tokens) && tokens[end].Kind == shared.EmDash {
		costSpan := shared.SpanOf(tokens[:end])
		mode.SpreeCost = &SpreeCostClause{
			Cost:          manaCost,
			Text:          shared.SliceSpan(source, costSpan),
			Span:          costSpan,
			SeparatorSpan: tokens[end].Span,
		}
		bodyTokens = tokens[end+1:]
	}
	mode.Body = phraseFromTokens(source, bodyTokens)
	mode.Sentences = ParseSentences(source, bodyTokens)
	var diagnostics []shared.Diagnostic
	mode.Reminders, mode.Quoted, diagnostics = parseDelimited(source, bodyTokens, diagnostics)
	return mode, diagnostics
}
