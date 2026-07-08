package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseOwnerPlayWhileExiledEffect recognizes "For as long as that card remains
// exiled, its owner may play it." — the owner-scoped play permission Prowl,
// Stoic Strategist grants a card it exiles from the battlefield ("exile up to
// one other target tapped creature or Vehicle. For as long as that card remains
// exiled, its owner may play it."). It is the owner-scoped, while-exiled sibling
// of the controller-scoped, turn-scoped "you may play it this turn." permission
// (parsePlayThatCardEffect): here the permission binds to the exiled card's
// OWNER, who may be an opponent of the resolving controller, and lasts for as
// long as the card remains in exile rather than a turn window.
//
// The recognizer requires the exact clause shape so it stays text-blind and
// corpus-safe: only "for as long as that card remains exiled, its owner may
// play it" yields the permission effect, and only when the trailing "it"
// resolves to a single referenced-object back-reference. Any other wording
// fails closed and flows through the generic effect parser. The recognized
// effect carries the referenced-object-owner recipient, the while-exiled
// duration, the optional ("may") marker, and the "it" back-reference so the
// paired exile clause and this permission lower together into a single
// exile-permanent-for-play primitive.
func parseOwnerPlayWhileExiledEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := wordsOnly(tokens)
	if !tokenWordsEqual(words,
		"for", "as", "long", "as", "that", "card", "remains", "exiled",
		"its", "owner", "may", "play", "it") {
		return nil, false
	}
	object := words[len(words)-1:]
	references := referencesInSpan(atoms, shared.SpanOf(object))
	if len(references) != 1 ||
		references[0].Kind != ReferencePronoun ||
		references[0].Pronoun != PronounIt {
		return nil, false
	}
	mayWord := words[len(words)-3]
	playWord := words[len(words)-2]
	itsWord := words[len(words)-5]
	return []EffectSyntax{{
		Kind:       EffectPlay,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   playWord.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextReferencedObjectOwner,
		Duration:   EffectDurationWhileExiled,
		Optional:   true,
		OptionalSpan: shared.Span{
			Start: itsWord.Span.Start,
			End:   mayWord.Span.End,
		},
		References: references,
		Exact:      true,
	}}, true
}
