package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// recognizeTibaltsTrickerySequence recognizes the closed six-effect sequence
// "Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that
// many cards, then exiles cards from the top of their library until they exile a
// nonland card with a different name than that spell. They may cast that card
// without paying its mana cost. Then they put the exiled cards on the bottom of
// their library in a random order." (Tibalt's Trickery).
//
// The parser owns the wording: it confirms the [Counter, Mill, Exile, Exile,
// Cast, Put] effect shape, the zero-effect "Choose 1, 2, or 3 at random."
// mill-count prelude, the "its controller"/"they" subject of the mill, exile,
// cast, and put effects, and every anchoring phrase (the different-name-nonland
// stop, the free cast, and the random-bottom remainder). It marks each effect
// with TibaltsTrickery, records the random mill-count range and prelude span on
// the head Counter effect, and credits the prelude sentence so the text-blind
// lowering can emit the counter, random-number choose, dynamic mill, and
// different-name-nonland IterativeLibraryProcess instructions. Any other wording
// fails closed. It returns true only when the full shape matches.
func recognizeTibaltsTrickerySequence(sentences []Sentence) bool {
	prelude, ok := tibaltRandomMillPrelude(sentences)
	if !ok {
		return false
	}
	effects := orderedRevealUntilEffects(sentences)
	if len(effects) != 6 {
		return false
	}
	counter, mill, exileHead, exileMatch, cast, put :=
		effects[0], effects[1], effects[2], effects[3], effects[4], effects[5]
	if counter.Kind != EffectCounter || mill.Kind != EffectMill ||
		exileHead.Kind != EffectExile || exileMatch.Kind != EffectExile ||
		cast.Kind != EffectCast || put.Kind != EffectPut {
		return false
	}
	if counter.Context != EffectContextController || counter.Optional || counter.Negated {
		return false
	}
	if !containsAll(strings.ToLower(counter.Text), "counter", "target spell") {
		return false
	}
	// The mill, exile, cast, and put all resolve in the countered spell's
	// controller context ("Its controller ... they ..."), never the caster's.
	if !isTargetControllerContext(mill.Context) ||
		!isTargetControllerContext(cast.Context) ||
		!isTargetControllerContext(put.Context) {
		return false
	}
	if mill.Amount.Known || !containsAll(strings.ToLower(mill.Selection.Text), "that many") {
		return false
	}
	if !containsAll(strings.ToLower(exileHead.Selection.Text),
		"from the top of their library", "until they exile") {
		return false
	}
	if !containsAll(strings.ToLower(exileMatch.Selection.Text),
		"nonland card", "different name than that spell") {
		return false
	}
	if !cast.Optional ||
		!containsAll(strings.ToLower(cast.Text), "cast that card", "without paying its mana cost") {
		return false
	}
	if put.ToZone != zone.Library ||
		!containsAll(strings.ToLower(put.Selection.Text),
			"the exiled cards", "on the bottom of their library", "in a random order") {
		return false
	}
	markTibaltsTrickeryEffects(effects)
	counter.TibaltRandomMillMin = prelude.millMin
	counter.TibaltRandomMillMax = prelude.millMax
	counter.TibaltPreludeSpan = sentences[prelude.index].Span
	sentences[prelude.index].ChooseNumberAtRandomPrelude = true
	return true
}

// isTargetControllerContext reports whether an effect resolves in the countered
// spell's controller context — the subject the mill, exile, cast, and put share
// through the "Its controller ... they ..." pronoun chain.
func isTargetControllerContext(context EffectContextKind) bool {
	return context == EffectContextReferencedObjectController ||
		context == EffectContextPriorSubject ||
		context == EffectContextEventPlayer
}

// markTibaltsTrickeryEffects flags every effect of a recognized Tibalt's
// Trickery sequence as exact and part of the folded sequence so coverage and
// reference scans credit them and the lowering identifies the sequence.
func markTibaltsTrickeryEffects(effects []*EffectSyntax) {
	for _, e := range effects {
		e.Exact = true
		e.TibaltsTrickery = true
	}
}

// tibaltMillPrelude locates the zero-effect "Choose 1, 2, or 3 at random."
// mill-count prelude: the sentence index it occupies and the inclusive [min, max]
// bounds of the offered consecutive run.
type tibaltMillPrelude struct {
	index   int
	millMin int
	millMax int
}

// tibaltRandomMillPrelude finds the zero-effect "Choose 1, 2, or 3 at random."
// mill-count prelude and returns its sentence index and the inclusive bounds of
// the chosen range. The prelude's integers must form the consecutive run
// [min..max] starting at a positive number, so a uniform pick over the offered
// numbers equals a uniform pick over the whole [min, max] range. Exactly one such
// prelude must be present; otherwise it fails closed.
func tibaltRandomMillPrelude(sentences []Sentence) (tibaltMillPrelude, bool) {
	prelude := tibaltMillPrelude{index: -1}
	for i := range sentences {
		if len(sentences[i].Effects) != 0 {
			continue
		}
		numbers, matched := parseChooseNumbersAtRandom(semanticEffectTokens(sentences[i].Tokens))
		if !matched {
			continue
		}
		if prelude.index >= 0 {
			return tibaltMillPrelude{}, false
		}
		lo, hi, consecutive := consecutiveRun(numbers)
		if !consecutive {
			return tibaltMillPrelude{}, false
		}
		prelude = tibaltMillPrelude{index: i, millMin: lo, millMax: hi}
	}
	if prelude.index < 0 {
		return tibaltMillPrelude{}, false
	}
	return prelude, true
}

// consecutiveRun reports whether numbers are strictly the ascending consecutive
// run min, min+1, ..., max with a positive start, and returns those bounds.
func consecutiveRun(numbers []int) (lo, hi int, ok bool) {
	if len(numbers) < 2 || numbers[0] < 1 {
		return 0, 0, false
	}
	for i := 1; i < len(numbers); i++ {
		if numbers[i] != numbers[i-1]+1 {
			return 0, 0, false
		}
	}
	return numbers[0], numbers[len(numbers)-1], true
}

// isChooseNumberAtRandomPreludeTokens reports whether the sentence tokens are a
// bare "Choose <n>, <n>, or <n> at random." mill-count naming prelude.
func isChooseNumberAtRandomPreludeTokens(tokens []shared.Token) bool {
	_, ok := parseChooseNumbersAtRandom(tokens)
	return ok
}

// parseChooseNumbersAtRandom parses the "Choose <numbers> at random." prelude,
// returning the offered integers in source order. The numbers may be separated
// by commas and a final "or"; the clause must end with "at random" followed only
// by a period. It requires at least two numbers so a genuine random choice is
// offered.
func parseChooseNumbersAtRandom(tokens []shared.Token) ([]int, bool) {
	if !effectWordsAt(tokens, 0, "choose") {
		return nil, false
	}
	i := 1
	var numbers []int
	expectNumber := true
	for i < len(tokens) {
		token := tokens[i]
		switch {
		case token.Kind == shared.Integer:
			if !expectNumber {
				return nil, false
			}
			value, err := strconv.Atoi(token.Text)
			if err != nil {
				return nil, false
			}
			numbers = append(numbers, value)
			expectNumber = false
		case token.Kind == shared.Comma || equalWord(token, "or"):
			expectNumber = true
		default:
			// The first non-number, non-separator token ends the number list.
			if !effectWordsAt(tokens, i, "at", "random") {
				return nil, false
			}
			for _, rest := range tokens[i+2:] {
				if rest.Kind != shared.Period {
					return nil, false
				}
			}
			if len(numbers) < 2 {
				return nil, false
			}
			return numbers, true
		}
		i++
	}
	return nil, false
}
