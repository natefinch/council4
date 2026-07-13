package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePlayExiledCardEffect recognizes the Hideaway play effect "You may play
// the exiled card without paying its mana cost" (CR 702.75c). It accepts either
// a trailing activation gate or a leading resolving "Then if <condition>,"
// gate; both conditions are parsed separately and lower onto the instruction.
// The effect plays the card the source permanent hid away with its Hideaway
// enters action: a land is put onto the battlefield and any other card is cast
// for free.
//
// The recognizer is intentionally narrow — it matches only the exact
// "play the exiled card without paying its mana cost" wording — so it never
// reinterprets other "play"/"cast" effects. Any other wording fails closed and
// flows through the generic effect parser.
func parsePlayExiledCardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := tokens
	if len(body) > 0 && equalWord(body[0], "then") {
		body = body[1:]
	}
	if len(body) > 0 {
		if intro, _ := conditionIntroAt(body, 0); intro == ConditionIntroIf {
			end := conditionClauseEnd(body, 0)
			if end < len(body) && body[end].Kind == shared.Comma {
				body = body[end+1:]
			}
		}
	}
	words := wordsOnly(body)
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
		Tokens:                    append([]shared.Token(nil), body...),
		Context:                   EffectContextController,
		Optional:                  true,
		OptionalSpan:              optionalSpan,
		PlayHideawayExiledCard:    true,
		CastWithoutPayingManaCost: true,
		References:                referencesInSpan(atoms, sentence.Span),
		Exact:                     true,
	}}, true
}
