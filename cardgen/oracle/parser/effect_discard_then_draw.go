package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// fuseDiscardThenDrawSentences folds each "discard {up to N|any number of}
// cards, then draw that many cards[ plus K]" looter, parsed as an adjacent
// discard + draw effect pair, into a single annotated discard effect. The
// player-chosen discard count feeds the "that many" draw back-reference, which
// the ordinary effect vocabulary cannot express as two independent
// instructions, so the pair lowers to one variable-looter primitive. This is
// the only place the looter's Oracle wording is inspected; the annotated typed
// fields carry the count bound and draw offset downstream.
func fuseDiscardThenDrawSentences(sentences []Sentence) {
	for si := range sentences {
		fuseDiscardThenDrawEffects(&sentences[si])
	}
}

func fuseDiscardThenDrawEffects(sentence *Sentence) {
	for i := 0; i+1 < len(sentence.Effects); i++ {
		maxCount, offset, ok := matchDiscardThenDrawPair(sentence, &sentence.Effects[i], &sentence.Effects[i+1])
		if !ok {
			continue
		}
		discard := &sentence.Effects[i]
		draw := sentence.Effects[i+1]
		discard.DiscardThenDraw = true
		discard.DiscardThenDrawMax = maxCount
		discard.DiscardThenDrawOffset = offset
		// A single fused effect must not demand ordered lowering, and its span
		// must cover both the discard and the consumed draw clause.
		discard.RequiresOrderedLowering = false
		discard.Span = spanCover(discard.Span, draw.Span)
		discard.ClauseSpan = spanCover(discard.ClauseSpan, draw.ClauseSpan)
		sentence.Effects = append(sentence.Effects[:i+1:i+1], sentence.Effects[i+2:]...)
		return
	}
}

// matchDiscardThenDrawPair reports whether the discard/draw effect pair is a
// controller looter and, if so, the discard upper bound (0 for "any number of
// cards") and the draw offset ("plus K", 0 when absent).
func matchDiscardThenDrawPair(sentence *Sentence, discard, draw *EffectSyntax) (maxCount, offset int, ok bool) {
	if discard.Kind != EffectDiscard ||
		discard.Context != EffectContextController ||
		discard.Negated ||
		draw.Kind != EffectDraw ||
		draw.Context != EffectContextController ||
		draw.Connection != EffectConnectionThen ||
		draw.Negated ||
		draw.Amount.DynamicKind != EffectDynamicAmountTriggeringCounterCount {
		return 0, 0, false
	}
	maxCount, ok = discardThenDrawCount(tokensWithinParserSpan(sentence.Tokens, discard.Selection.Span))
	if !ok {
		return 0, 0, false
	}
	offset, ok = discardThenDrawOffset(tokensWithinParserSpan(sentence.Tokens, draw.Selection.Span))
	if !ok {
		return 0, 0, false
	}
	return maxCount, offset, true
}

// discardThenDrawCount reads the variable discard selection tokens, returning
// the upper bound for "up to <N> cards" (N) or "any number of cards" (0). Any
// other selection (a fixed "two cards", a filtered selection) is not a variable
// looter and is rejected.
func discardThenDrawCount(tokens []shared.Token) (int, bool) {
	if len(tokens) == 4 &&
		isWord(tokens[0], "any") && isWord(tokens[1], "number") &&
		isWord(tokens[2], "of") && isWord(tokens[3], "cards") {
		return 0, true
	}
	if len(tokens) == 4 &&
		isWord(tokens[0], "up") && isWord(tokens[1], "to") && isWord(tokens[3], "cards") {
		value, found := bottomHandThenDrawOffsetWords[tokens[2].Text]
		if found {
			return value, true
		}
	}
	return 0, false
}

// discardThenDrawOffset reads the "that many cards[ plus <K>]" draw selection
// tokens, returning the spelled-out offset K (0 when no "plus" rider is
// present). A "plus" without a known spelled count is rejected.
func discardThenDrawOffset(tokens []shared.Token) (int, bool) {
	for i := range tokens {
		if !isWord(tokens[i], "plus") {
			continue
		}
		if i+1 >= len(tokens) {
			return 0, false
		}
		value, found := bottomHandThenDrawOffsetWords[tokens[i+1].Text]
		if !found {
			return 0, false
		}
		return value, true
	}
	return 0, true
}

func isWord(token shared.Token, text string) bool {
	return token.Kind == shared.Word && token.Text == text
}

// spanCover returns the smallest span covering both inputs.
func spanCover(a, b shared.Span) shared.Span {
	result := a
	if b.Start.Offset < result.Start.Offset {
		result.Start = b.Start
	}
	if b.End.Offset > result.End.Offset {
		result.End = b.End
	}
	return result
}
