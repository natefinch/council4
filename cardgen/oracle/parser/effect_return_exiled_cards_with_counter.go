package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// parseReturnExiledCardsWithCounterEffect recognizes the resolution-time mass
// return "Put all exiled cards you own with <kind> counters on them into your
// hand." (Flamewar, Brash Veteran): every card the resolving controller owns in
// exile that bears the named marker counter returns to that controller's hand.
// The recognized effect carries the source zone (Exile), the destination zone
// (Hand), and the marker-counter filter (on the shared CounterKind/CounterKnown
// fields); it lowers on its own to the exile-with-named-counter substrate's
// return companion.
//
// Only the exact "you own ... with <kind> counters on them into your hand"
// wording is accepted; any other owner, destination, or counter phrasing fails
// closed and flows through the generic effect parser, where the owner scope and
// counter filter would otherwise leave a dangling "them" reference that fails
// activation-reference support. Reading the counter name through
// counterNameBefore keeps the recognizer text-blind to which named counter the
// card uses, so every named-counter-exile card benefits.
func parseReturnExiledCardsWithCounterEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	if !staticWordsAt(tokens, 0, "put", "all", "exiled", "cards", "you", "own", "with") {
		return nil, false
	}
	nameStart := 7
	counterIndex := -1
	for i := nameStart; i < len(tokens); i++ {
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			counterIndex = i
			break
		}
	}
	if counterIndex <= nameStart {
		return nil, false
	}
	kind, span, ok := counterNameBefore(tokens, counterIndex)
	if !ok || span.Start.Offset != tokens[nameStart].Span.Start.Offset {
		return nil, false
	}
	// The clause must close exactly "<kind> counters on them into your hand."
	// with no trailing tokens, so an unmodeled rider never rides in unrecognized.
	if counterIndex+6 != len(tokens)-1 ||
		!equalWord(tokens[counterIndex+1], "on") ||
		!equalWord(tokens[counterIndex+2], "them") ||
		!equalWord(tokens[counterIndex+3], "into") ||
		!equalWord(tokens[counterIndex+4], "your") ||
		!equalWord(tokens[counterIndex+5], "hand") ||
		tokens[counterIndex+6].Kind != shared.Period {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:         EffectReturnExiledCardsWithCounter,
		Span:         sentence.Span,
		ClauseSpan:   sentence.Span,
		VerbSpan:     tokens[0].Span,
		Text:         sentence.Text,
		Tokens:       append([]shared.Token(nil), tokens...),
		Context:      EffectContextController,
		FromZone:     zone.Exile,
		ToZone:       zone.Hand,
		CounterKind:  kind,
		CounterKnown: true,
		Exact:        true,
	}}, true
}
