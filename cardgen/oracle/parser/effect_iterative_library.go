package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// recognizeTaintedPactSequence recognizes the closed iterative library-processing
// family "Exile the top card of your library. You may put that card into your
// hand unless it has the same name as another card exiled this way. Repeat this
// process until you put a card into your hand or you exile two cards with the
// same name, whichever comes first." (Tainted Pact). The controller exiles cards
// from the top one at a time, may take each into hand, and stops the moment two
// exiled cards share a name.
//
// The parser owns the wording: it confirms the four-effect [Exile, may-Put,
// Put, Exile] shape, that every clause acts on the controller's own library and
// hand, and that the repeat clause names the duplicate-name stop. It marks each
// effect with IterativeLibraryProcess and records the duplicate-name stop plus
// the optional-take knob on the head Exile so the text-blind lowering can emit a
// single IterativeLibraryProcess primitive. Any other wording fails closed.
func recognizeTaintedPactSequence(sentences []Sentence) bool {
	effects := orderedRevealUntilEffects(sentences)
	if len(effects) != 4 {
		return false
	}
	exileTop, mayPut, put, exileDup := effects[0], effects[1], effects[2], effects[3]
	if exileTop.Kind != EffectExile || mayPut.Kind != EffectPut ||
		put.Kind != EffectPut || exileDup.Kind != EffectExile {
		return false
	}
	if !allControllerEffects(effects) {
		return false
	}
	if !mayPut.Optional || mayPut.ToZone != zone.Hand || put.ToZone != zone.Hand {
		return false
	}
	if !containsAll(strings.ToLower(exileTop.Text), "exile", "top card", "library") {
		return false
	}
	if !containsAll(strings.ToLower(mayPut.Text),
		"you may put", "into your hand", "same name as another card exiled this way") {
		return false
	}
	if !containsAll(strings.ToLower(put.Text),
		"repeat this process until", "put a card into your hand", "two cards with the same name") {
		return false
	}
	if !exileDup.Amount.Known || exileDup.Amount.Value != 2 {
		return false
	}
	markIterativeLibraryEffects(effects)
	exileTop.IterativeLibraryStop = IterativeLibraryStopDuplicateName
	exileTop.IterativeLibraryOptionalTake = true
	return true
}

// recognizeDemonicConsultationSequence recognizes the closed iterative library-
// processing family "Choose a card name. Exile the top six cards of your
// library, then reveal cards from the top of your library until you reveal a
// card with the chosen name. Put that card into your hand and exile all other
// cards revealed this way." (Demonic Consultation). The controller names a card,
// exiles a fixed six from the top, then reveals from the top until the named
// card appears; that card goes to hand and every other revealed card is exiled.
//
// The parser owns the wording: it confirms the zero-effect "Choose a card name."
// prelude, the five-effect [Exile, Reveal, Reveal, Put, Exile] body, that every
// clause acts on the controller's own library and hand, and that the reveal
// stops on the chosen name. It marks each effect with IterativeLibraryProcess,
// records the chosen-name stop, reveal, choose-name, and pre-exile count on the
// head Exile, and credits the prelude sentence so the text-blind lowering can
// emit a single IterativeLibraryProcess primitive. Any other wording fails
// closed. It returns true only when the full shape matches.
func recognizeDemonicConsultationSequence(sentences []Sentence) bool {
	preludeIdx := -1
	for i := range sentences {
		if len(sentences[i].Effects) == 0 &&
			isChooseCardNamePreludeTokens(semanticEffectTokens(sentences[i].Tokens)) {
			preludeIdx = i
			break
		}
	}
	if preludeIdx < 0 {
		return false
	}
	effects := orderedRevealUntilEffects(sentences)
	if len(effects) != 5 {
		return false
	}
	preExile, firstReveal, matchReveal, put, exileRest :=
		effects[0], effects[1], effects[2], effects[3], effects[4]
	if preExile.Kind != EffectExile || firstReveal.Kind != EffectReveal ||
		matchReveal.Kind != EffectReveal || put.Kind != EffectPut ||
		exileRest.Kind != EffectExile {
		return false
	}
	if !allControllerEffects(effects) {
		return false
	}
	if !preExile.Amount.Known || preExile.Amount.Value < 1 {
		return false
	}
	if !containsAll(strings.ToLower(preExile.Text),
		"exile", "cards of your library",
		"reveal cards from the top of your library", "until you reveal a card with the chosen name") {
		return false
	}
	if !strings.Contains(strings.ToLower(matchReveal.Selection.Text), "chosen name") {
		return false
	}
	if put.ToZone != zone.Hand ||
		!strings.Contains(strings.ToLower(put.Text), "put that card into your hand") {
		return false
	}
	if !strings.Contains(strings.ToLower(exileRest.Selection.Text), "all other cards revealed this way") {
		return false
	}
	markIterativeLibraryEffects(effects)
	preExile.IterativeLibraryStop = IterativeLibraryStopChosenName
	preExile.IterativeLibraryChooseName = true
	preExile.IterativeLibraryReveal = true
	preExile.IterativeLibraryPreExile = preExile.Amount.Value
	preExile.IterativeLibraryPreludeSpan = sentences[preludeIdx].Span
	sentences[preludeIdx].ChooseCardNamePrelude = true
	return true
}

// isChooseCardNamePreludeTokens reports whether the sentence tokens are the bare
// "Choose a card name." naming prelude, allowing only a trailing period.
func isChooseCardNamePreludeTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "choose", "a", "card", "name") {
		return false
	}
	for _, tok := range tokens[4:] {
		if tok.Kind != shared.Period {
			return false
		}
	}
	return true
}

// allControllerEffects reports whether every effect resolves in the controller's
// own context, so the single IterativeLibraryProcess primitive's controller
// player reference is faithful.
func allControllerEffects(effects []*EffectSyntax) bool {
	for _, e := range effects {
		if e.Context != EffectContextController {
			return false
		}
	}
	return true
}

// markIterativeLibraryEffects flags every effect of a recognized iterative
// library process as exact and part of the folded sequence.
func markIterativeLibraryEffects(effects []*EffectSyntax) {
	for _, e := range effects {
		e.Exact = true
		e.IterativeLibraryProcess = true
	}
}

// containsAll reports whether text contains every substring in subs.
func containsAll(text string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(text, sub) {
			return false
		}
	}
	return true
}
