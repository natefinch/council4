package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// isPileSplitMiddleTokens reports whether the sentence tokens are one of the two
// recognized zero-effect pile-split middle sentences ("An opponent separates
// those cards into two piles." / "An opponent chooses one of those piles."). It
// gates the middle sentence as a pile-split candidate so it is not treated as an
// unrecognized sibling before the full sequence is validated.
func isPileSplitMiddleTokens(tokens []shared.Token) bool {
	_, _, ok := pileSplitMiddleRole(normalizedWords(tokens))
	return ok
}

// recognizePileSplitSequence recognizes the closed pile-split effect family
// "reveal the top N cards of your library[ and separate them into two piles]. An
// opponent {separates those cards into|chooses one of} two piles. Put {one|that}
// pile into your hand and the other into your graveyard." (Fact or Fiction,
// Steam Augury, Sphinx of Uthuun). The controller reveals the top N cards; one
// player divides them into two piles and another chooses which pile the
// controller keeps; the kept pile goes to the controller's hand and the other to
// the controller's graveyard.
//
// The parser owns the wording: it confirms the [Reveal, <zero-effect middle>,
// Put] sentence shape, that the reveal draws the top N of the controller's
// library, that the separate and choose roles are split between the controller
// and an opponent consistently across the reveal and middle sentences, and that
// the put clause moves one pile into the controller's hand and the other into
// the controller's graveyard. It marks both effects with PileSplitSequence,
// records the roles, destination, amount, and middle-sentence span on the put
// effect, and marks the middle sentence as a credited rider so the text-blind
// lowering can emit a single PileSplit primitive. Any other wording fails closed.
func recognizePileSplitSequence(sentences []Sentence) bool {
	revealIdx, putIdx, ok := pileSplitSentenceShape(sentences)
	if !ok {
		return false
	}
	midIdx := revealIdx + 1
	reveal := &sentences[revealIdx].Effects[0]
	put := &sentences[putIdx].Effects[0]
	if reveal.Kind != EffectReveal || put.Kind != EffectPut ||
		reveal.Context != EffectContextController || put.Context != EffectContextController ||
		!pileSplitFixedAmount(reveal.Amount) {
		return false
	}

	separateInReveal, ok := pileSplitRevealShape(normalizedWords(semanticEffectTokens(sentences[revealIdx].Tokens)))
	if !ok {
		return false
	}
	separatorOpponent, chooserOpponent, ok := pileSplitMiddleRole(normalizedWords(semanticEffectTokens(sentences[midIdx].Tokens)))
	if !ok {
		return false
	}
	otherZone, ok := pileSplitPutShape(normalizedWords(semanticEffectTokens(sentences[putIdx].Tokens)))
	if !ok {
		return false
	}
	// The reveal sentence's "and separate them into two piles" clause and the
	// middle sentence's role must name complementary roles: either the reveal
	// states the controller separates and the middle states an opponent chooses,
	// or the reveal omits the separate clause and the middle states an opponent
	// separates (the controller then chooses by putting one pile into hand).
	if separateInReveal {
		if separatorOpponent || !chooserOpponent {
			return false
		}
	} else {
		if !separatorOpponent || chooserOpponent {
			return false
		}
	}

	reveal.Exact = true
	reveal.PileSplitSequence = true
	put.Exact = true
	put.PileSplitSequence = true
	put.PileSplitSeparatorOpponent = separatorOpponent
	put.PileSplitChooserOpponent = chooserOpponent
	put.PileSplitOtherZone = otherZone
	put.PileSplitAmount = reveal.Amount.Value
	put.PileSplitMiddleSpan = sentences[midIdx].Span
	sentences[midIdx].PileSplitRider = true
	return true
}

// pileSplitFixedAmount reports whether an amount is a plain, known, positive
// fixed integer with no variable, dynamic, range, or "plus N" rider. The
// PileSplit primitive reveals a fixed count, so amounts such as "X plus one"
// (Epiphany at the Drownyard) or any rules-derived count fail closed here rather
// than silently lowering to a wrong fixed number.
func pileSplitFixedAmount(amount EffectAmountSyntax) bool {
	return amount.Known && amount.Value >= 1 &&
		!amount.VariableX && !amount.RangeKnown && !amount.AnyNumber &&
		amount.DynamicKind == EffectDynamicAmountNone && amount.Addend == 0 && amount.Multiplier == 0 &&
		len(amount.Operands) == 0 && amount.Selection == nil
}

// pileSplitSentenceShape locates the [Reveal, middle, Put] sentence triple of a
// pile-split sequence: exactly two effect-bearing sentences (the first holding a
// single reveal effect, the second a single put effect) separated by exactly one
// zero-effect middle sentence. The middle sentence is always at revealIdx+1. Any
// other distribution fails closed.
func pileSplitSentenceShape(sentences []Sentence) (revealIdx, putIdx int, ok bool) {
	var effectSentences []int
	totalEffects := 0
	for i := range sentences {
		totalEffects += len(sentences[i].Effects)
		if len(sentences[i].Effects) > 0 {
			effectSentences = append(effectSentences, i)
		}
	}
	if totalEffects != 2 || len(effectSentences) != 2 {
		return 0, 0, false
	}
	revealIdx = effectSentences[0]
	putIdx = effectSentences[1]
	if putIdx != revealIdx+2 ||
		len(sentences[revealIdx].Effects) != 1 ||
		len(sentences[putIdx].Effects) != 1 ||
		len(sentences[revealIdx+1].Effects) != 0 {
		return 0, 0, false
	}
	return revealIdx, putIdx, true
}

// pileSplitRevealShape validates the reveal sentence "reveal the top N cards of
// your library" and reports whether it carries the trailing "and separate them
// into two piles" clause (the controller-separates Steam Augury form). The count
// word between "top" and "cards" is treated as a wildcard; the parser already
// typed it as the reveal amount.
func pileSplitRevealShape(words []string) (separateInReveal, ok bool) {
	// The count occupies exactly one token between "top" and "cards"; multi-word
	// counts such as "X plus one" (Epiphany at the Drownyard) carry a variable the
	// fixed-amount PileSplit primitive cannot model and fail closed here.
	if !hasPrefix(words, "reveal", "the", "top") || len(words) < 8 {
		return false, false
	}
	if !slices.Equal(words[4:8], []string{"cards", "of", "your", "library"}) {
		return false, false
	}
	rest := words[8:]
	if len(rest) == 0 {
		return false, true
	}
	if slices.Equal(rest, []string{"and", "separate", "them", "into", "two", "piles"}) {
		return true, true
	}
	return false, false
}

// pileSplitMiddleRole validates the zero-effect middle sentence and reports which
// role an opponent fills: "an opponent separates those cards into two piles"
// (opponent separates) or "an opponent chooses one of those piles" (opponent
// chooses).
func pileSplitMiddleRole(words []string) (separatorOpponent, chooserOpponent, ok bool) {
	if slices.Equal(words, []string{"an", "opponent", "separates", "those", "cards", "into", "two", "piles"}) {
		return true, false, true
	}
	if slices.Equal(words, []string{"an", "opponent", "chooses", "one", "of", "those", "piles"}) {
		return false, true, true
	}
	return false, false, false
}

// pileSplitPutShape validates the put sentence "put {one|that} pile into your
// hand and the other into your graveyard" and returns the destination of the
// pile the controller does not keep (the kept pile always goes to hand).
func pileSplitPutShape(words []string) (zone.Type, bool) {
	if len(words) != 12 || words[0] != "put" ||
		(words[1] != "one" && words[1] != "that") {
		return zone.None, false
	}
	if !slices.Equal(words[2:11], []string{"pile", "into", "your", "hand", "and", "the", "other", "into", "your"}) {
		return zone.None, false
	}
	if words[11] == "graveyard" {
		return zone.Graveyard, true
	}
	return zone.None, false
}

// hasPrefix reports whether words begins with the given prefix words.
func hasPrefix(words []string, prefix ...string) bool {
	if len(words) < len(prefix) {
		return false
	}
	for i, w := range prefix {
		if words[i] != w {
			return false
		}
	}
	return true
}
