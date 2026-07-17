package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseKeywordUnionTargetSelection(
	tokens []shared.Token,
	cardinality TargetCardinalitySyntax,
	atoms Atoms,
) (SelectionSyntax, bool) {
	with := -1
	for i := len(tokens) - 1; i >= 0; i-- {
		if equalWord(tokens[i], "with") {
			with = i
			break
		}
	}
	if with <= 0 {
		return SelectionSyntax{}, false
	}
	first, firstWidth, ok := recognizeKeywordNameAt(tokens, with+1)
	if !ok || with+1+firstWidth >= len(tokens) ||
		!equalWord(tokens[with+1+firstWidth], "or") {
		return SelectionSyntax{}, false
	}
	secondStart := with + 2 + firstWidth
	second, secondWidth, ok := recognizeKeywordNameAt(tokens, secondStart)
	if !ok || secondStart+secondWidth != len(tokens) ||
		first == KeywordUnknown || second == KeywordUnknown {
		return SelectionSyntax{}, false
	}
	baseTokens := tokens[:with]
	switch {
	case cardinality.Min == 0 && cardinality.Max == 99 &&
		len(baseTokens) >= 3 && effectWordsAt(baseTokens, 0, "any", "number", "of"):
		baseTokens = baseTokens[3:]
	case cardinality.Min == 0 && cardinality.Max > 0 &&
		len(baseTokens) >= 3 && effectWordsAt(baseTokens, 0, "up", "to"):
		baseTokens = baseTokens[3:]
	case cardinality.Min == cardinality.Max && cardinality.Max > 1 && len(baseTokens) > 1:
		if count, ok := effectNumber(baseTokens[0], atoms); ok && count == cardinality.Max {
			baseTokens = baseTokens[1:]
		}
	default:
	}
	base := parseSelection(baseTokens, atoms)
	if base.Kind == SelectionUnknown || base.Keyword != KeywordUnknown ||
		len(base.Alternatives) != 0 {
		return SelectionSyntax{}, false
	}
	base.Span = shared.SpanOf(tokens)
	base.Text = joinedEffectText(tokens)
	base.Alternatives = []SelectionSyntax{
		{
			Span:    shared.SpanOf(tokens[with+1 : with+1+firstWidth]),
			Text:    joinedEffectText(tokens[with+1 : with+1+firstWidth]),
			Keyword: first,
		},
		{
			Span:    shared.SpanOf(tokens[secondStart : secondStart+secondWidth]),
			Text:    joinedEffectText(tokens[secondStart : secondStart+secondWidth]),
			Keyword: second,
		},
	}
	return base, true
}

func exactKeywordUnionTargetSyntax(text string, selection SelectionSyntax) bool {
	if len(selection.Alternatives) != 2 {
		return false
	}
	keywords := make([]string, 0, 2)
	for i := range selection.Alternatives {
		alternative := selection.Alternatives[i]
		if alternative.Keyword == KeywordUnknown {
			return false
		}
		keyword := alternative.Keyword
		alternative.Keyword = KeywordUnknown
		if alternative.Kind != SelectionUnknown ||
			alternative.Controller != SelectionControllerAny ||
			len(alternative.RequiredTypesAny) != 0 ||
			len(alternative.Supertypes) != 0 ||
			len(alternative.SubtypesAny) != 0 ||
			len(alternative.ColorsAny) != 0 ||
			len(alternative.Alternatives) != 0 {
			return false
		}
		keywords = append(keywords, keyword.String())
	}
	base := selection
	base.Alternatives = nil
	expected, ok := exactPermanentTargetText(base)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected+" with "+keywords[0]+" or "+keywords[1])
}
