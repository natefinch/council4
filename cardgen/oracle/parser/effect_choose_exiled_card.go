package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// parseChooseExiledCardEffect recognizes the resolution-time choice "Choose an
// exiled card an opponent owns with a <kind> counter on it." (Dauthi
// Voidwalker): the resolving controller picks one card resting in exile that an
// opponent owns and that bears the named exile marker counter. The recognized
// effect carries the source zone (Exile), the opponent owner scope, and the
// marker-counter filter (on the shared CounterKind/CounterKnown fields); it
// lowers together with a following "You may play it this turn ..." permission
// into a single choose-then-play-from-exile primitive.
//
// Only the exact "an opponent owns" owner phrase and a "<kind> counter on it"
// filter are accepted; any other owner, source, or counter wording fails closed
// and flows through the generic effect parser so the owner scope, source zone,
// and counter filter are never silently dropped.
func parseChooseExiledCardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	if !staticWordsAt(tokens, 0, "choose", "an", "exiled", "card", "an", "opponent", "owns", "with") {
		return nil, false
	}
	articleIndex := 8
	if articleIndex >= len(tokens) || (!equalWord(tokens[articleIndex], "a") && !equalWord(tokens[articleIndex], "an")) {
		return nil, false
	}
	nameStart := articleIndex + 1
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
	// The filter must close exactly "<kind> counter on it." with no trailing
	// tokens, so an unmodeled rider never rides in unrecognized.
	if counterIndex+3 != len(tokens)-1 ||
		!equalWord(tokens[counterIndex+1], "on") ||
		!equalWord(tokens[counterIndex+2], "it") ||
		tokens[counterIndex+3].Kind != shared.Period {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                          EffectChooseExiledCard,
		Span:                          sentence.Span,
		ClauseSpan:                    sentence.Span,
		VerbSpan:                      tokens[0].Span,
		Text:                          sentence.Text,
		Tokens:                        append([]shared.Token(nil), tokens...),
		Context:                       EffectContextController,
		FromZone:                      zone.Exile,
		ChooseExiledCardOwnerOpponent: true,
		CounterKind:                   kind,
		CounterKnown:                  true,
		Exact:                         true,
	}}, true
}
