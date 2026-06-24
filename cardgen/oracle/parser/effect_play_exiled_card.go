package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePlayExiledCardEffect recognizes the Hideaway activated-ability effect
// "You may play the exiled card without paying its mana cost" (CR 702.75c),
// optionally followed by an "if <condition>" activation gate that is parsed
// separately as the ability's condition. It plays the card the source permanent
// hid away with its Hideaway enters action: a land is put onto the battlefield
// and any other card is cast for free.
//
// The recognizer is intentionally narrow — it matches only the exact
// "play the exiled card without paying its mana cost" wording — so it never
// reinterprets other "play"/"cast" effects. Any other wording fails closed and
// flows through the generic effect parser.
func parsePlayExiledCardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := wordsOnly(tokens)
	rest, ok := cutTokenPrefix(words, "you", "may")
	if !ok {
		return nil, false
	}
	optionalSpan := shared.Span{Start: words[0].Span.Start, End: words[1].Span.End}
	verbSpan := rest[0].Span
	after, ok := cutTokenPrefix(rest, "play", "the", "exiled", "card", "without", "paying", "its", "mana", "cost")
	if !ok {
		return nil, false
	}
	// The only permitted continuation is the "if <condition>" activation gate,
	// which the condition parser extracts separately.
	if len(after) != 0 && !equalWord(after[0], "if") {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                      EffectPlay,
		Span:                      sentence.Span,
		ClauseSpan:                sentence.Span,
		VerbSpan:                  verbSpan,
		Text:                      sentence.Text,
		Tokens:                    append([]shared.Token(nil), tokens...),
		Context:                   EffectContextController,
		Optional:                  true,
		OptionalSpan:              optionalSpan,
		PlayHideawayExiledCard:    true,
		CastWithoutPayingManaCost: true,
		Exact:                     true,
	}}, true
}
