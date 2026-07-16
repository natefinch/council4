package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseStaticDevotionNotCreatureDeclaration recognizes the Theros Gods'
// devotion-gated type-changing static "As long as your devotion to <color> is
// less than N, <source> isn't a creature." and its two-color form "As long as
// your devotion to <color> and <color> is less than N, <source> isn't a
// creature." (Purphoros, God of the Forge; Athreos, God of Passage; the full
// God family). The parser owns the entire wording: it captures the one or two
// devotion colors and the numeric threshold N as typed data so downstream
// stages remove the creature type while devotion is below N without inspecting
// any Oracle text or card name.
//
// Per the Gods' rulings the type-changing ability functions only on the
// battlefield (a God is always a creature card in other zones and a creature
// spell on the stack), which the lowering models as a battlefield continuous
// LayerType effect gated by a devotion condition. Only this exact wording is
// recognized; any deviation leaves the clause unconsumed and the card fails
// closed.
func parseStaticDevotionNotCreatureDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "as", "long", "as", "your", "devotion", "to") {
		return StaticDeclarationSyntax{}, false
	}
	colors, next, ok := staticDevotionColors(tokens, 6)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, next, "is", "less", "than") {
		return StaticDeclarationSyntax{}, false
	}
	amountIdx := next + 3
	if amountIdx >= len(tokens) || tokens[amountIdx].Kind != shared.Word {
		return StaticDeclarationSyntax{}, false
	}
	threshold, ok := CardinalWordValue(tokens[amountIdx].Text)
	if !ok || threshold < 1 {
		return StaticDeclarationSyntax{}, false
	}
	commaIdx := amountIdx + 1
	if commaIdx >= len(tokens) || tokens[commaIdx].Kind != shared.Comma {
		return StaticDeclarationSyntax{}, false
	}
	nameIdx := commaIdx + 1
	if nameIdx >= len(tokens) {
		return StaticDeclarationSyntax{}, false
	}
	nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[nameIdx].Span)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	nameEnd := nameIdx
	for nameEnd < len(tokens) && tokens[nameEnd].Span.End.Offset <= nameSpan.End.Offset {
		nameEnd++
	}
	if !staticWordsAt(tokens, nameEnd, "isn't", "a", "creature") {
		return StaticDeclarationSyntax{}, false
	}
	if nameEnd+3 != len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:              StaticDeclarationDevotionNotCreature,
		Span:              shared.SpanOf(tokens),
		OperationSpan:     shared.SpanOf(tokens[nameEnd : nameEnd+3]),
		DevotionColors:    colors,
		DevotionThreshold: threshold,
	}, true
}

// staticDevotionColors reads the one or two devotion colors of a "your devotion
// to <color> [and <color>]" phrase starting at start and returns the colors in
// source order with the index of the first token after them. It fails closed
// for an unknown color word or a three-or-more-color list, so only the mono and
// two-color God wordings are recognized.
func staticDevotionColors(tokens []shared.Token, start int) ([]Color, int, bool) {
	if start >= len(tokens) || tokens[start].Kind != shared.Word {
		return nil, 0, false
	}
	first, ok := recognizeColorWord(tokens[start].Text)
	if !ok {
		return nil, 0, false
	}
	colors := []Color{first}
	next := start + 1
	if staticWordsAt(tokens, next, "and") &&
		next+1 < len(tokens) && tokens[next+1].Kind == shared.Word {
		second, ok := recognizeColorWord(tokens[next+1].Text)
		if !ok {
			return nil, 0, false
		}
		colors = append(colors, second)
		next += 2
	}
	return colors, next, true
}
