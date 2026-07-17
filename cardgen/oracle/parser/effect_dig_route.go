package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// recognizeDigRouteSequence recognizes the closed look-and-route dig family
// "Look at the top N cards of your library. Put one of them into your hand, put
// one of them on the bottom of your library, and exile one of them. You may play
// the exiled card this turn." (Expressive Iteration): a single hidden look at
// the top N cards that fans the looked-at cards out into three ordered,
// mutually-exclusive destinations (hand, bottom of library, exile) with an
// impulse "play this turn" permission on the exiled card.
//
// It mirrors recognizeImpulseExileSequence: the parser owns the wording, folding
// all three sentences into one EffectDig whose typed DigRoute records the look
// count and the ordered hand / library-bottom / exile slots so the text-blind
// lowering can emit a single Dig with ordered slots. It matches only the exact
// three-way routing above: any other look/route counts (the routed cards must
// partition the looked-at cards), destination, ordering, permission duration, a
// cast-only ("you may cast ...") permission, or a free ("without paying its mana
// cost") permission fails closed, leaving the card unsupported.
func recognizeDigRouteSequence(sentences []Sentence) bool {
	// Trailing reminder text is parsed as its own parenthesized sentence and
	// carries no game meaning, so ignore it when matching the three-sentence
	// shape (parity with recognizeImpulseExileSequence).
	for len(sentences) > 3 && isReminderSentence(sentences[len(sentences)-1]) {
		sentences = sentences[:len(sentences)-1]
	}
	if len(sentences) != 3 {
		return false
	}
	look, ok := matchDigRouteLookClause(strings.TrimSpace(sentences[0].Text))
	if !ok {
		return false
	}
	slots, ok := matchDigRouteDistributeClause(strings.TrimSpace(sentences[1].Text))
	if !ok {
		return false
	}
	// Unique routing: each route moves exactly one of the looked-at cards to its
	// own destination, matching the "put one of them ... put one of them ...
	// exile one of them" wording. Any other per-route count is different wording
	// the exact recognizer leaves unsupported.
	for i := range slots {
		if slots[i].Count != 1 {
			return false
		}
	}
	// The exile slot is always the third route; its single card carries the
	// impulse "play the exiled card this turn" permission from the third
	// sentence. Only a plain this-turn "play" permission over exactly the one
	// exiled card is modeled; cast-only, free, and other-duration permissions
	// fail closed inside matchDigRoutePlayPermissionClause.
	exile := &slots[len(slots)-1]
	if !matchDigRoutePlayPermissionClause(strings.TrimSpace(sentences[2].Text)) {
		return false
	}
	exile.PlayThisTurn = true
	// The routes must partition the looked-at cards: their counts sum to the
	// look count, so no looked-at card is left with an unstated fate (a card
	// that left a remainder would carry an explicit "and the rest ..." clause).
	total := 0
	for i := range slots {
		total += slots[i].Count
	}
	if total != look {
		return false
	}

	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[2].Span.End}
	tokens := append(append(append([]shared.Token(nil), sentences[0].Tokens...), sentences[1].Tokens...), sentences[2].Tokens...)
	sentences[0].Effects = []EffectSyntax{{
		Kind:             EffectDig,
		Context:          EffectContextController,
		Span:             span,
		ClauseSpan:       span,
		Text:             sentences[0].Text + " " + sentences[1].Text + " " + sentences[2].Text,
		Tokens:           tokens,
		Amount:           EffectAmountSyntax{Value: look, Known: true},
		Exact:            true,
		DigRouteSequence: true,
		DigRoute:         DigRouteSyntax{Look: look, Slots: slots},
	}}
	return true
}

// matchDigRouteLookClause recognizes "Look at the top N cards of your library."
// and returns N (a cardinal word two or greater; the three-way routing always
// looks at more than one card). Any other wording fails closed.
func matchDigRouteLookClause(text string) (int, bool) {
	const prefix = "Look at the top "
	const suffix = " cards of your library."
	middle, ok := digRouteCardinalBetween(text, prefix, suffix)
	if !ok || middle < 2 {
		return 0, false
	}
	return middle, true
}

// matchDigRouteDistributeClause recognizes the three-way routing sentence "Put
// <a> of them into your hand, put <b> of them on the bottom of your library, and
// exile <c> of them." and returns the ordered hand, library-bottom, and exile
// slots. The three routes must appear in exactly this order with these
// destinations; any other order, destination, count, or shape fails closed.
func matchDigRouteDistributeClause(text string) ([]DigRouteSlotSyntax, bool) {
	body, ok := strings.CutSuffix(text, ".")
	if !ok {
		return nil, false
	}
	parts := strings.Split(body, ", ")
	if len(parts) != 3 {
		return nil, false
	}
	hand, ok := digRouteCardinalBetween(parts[0], "Put ", " of them into your hand")
	if !ok || hand < 1 {
		return nil, false
	}
	bottom, ok := digRouteCardinalBetween(parts[1], "put ", " of them on the bottom of your library")
	if !ok || bottom < 1 {
		return nil, false
	}
	exile, ok := digRouteCardinalBetween(parts[2], "and exile ", " of them")
	if !ok || exile < 1 {
		return nil, false
	}
	return []DigRouteSlotSyntax{
		{Count: hand, Destination: zone.Hand},
		{Count: bottom, Destination: zone.Library, Bottom: true},
		{Count: exile, Destination: zone.Exile},
	}, true
}

// matchDigRoutePlayPermissionClause reports whether text is exactly the impulse
// permission sentence "You may play the exiled card this turn." — a plain
// this-turn play permission over the single exiled card. A cast-only permission
// ("You may cast the exiled card this turn."), a free-cast rider ("... without
// paying its mana cost"), and any other play window (until end of turn, until
// your next turn) all fail closed, so the exile slot only ever carries the
// modeled "play this turn" grant.
func matchDigRoutePlayPermissionClause(text string) bool {
	window, ok := matchImpulsePermissionWindow(text, "play", "the exiled card")
	return ok && window == EffectDurationThisTurn
}

// digRouteCardinalBetween returns the cardinal-word integer that appears between
// prefix and suffix in text (case-insensitively), or ok=false when text does not
// have that exact frame or the middle is not a cardinal word.
func digRouteCardinalBetween(text, prefix, suffix string) (int, bool) {
	if len(text) <= len(prefix)+len(suffix) ||
		!strings.EqualFold(text[:len(prefix)], prefix) ||
		!strings.EqualFold(text[len(text)-len(suffix):], suffix) {
		return 0, false
	}
	return CardinalWordValue(text[len(prefix) : len(text)-len(suffix)])
}
