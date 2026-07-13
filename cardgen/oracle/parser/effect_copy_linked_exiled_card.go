package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseCopyLinkedExiledCardEffect recognizes the imprint copy-consent effect
// "You may copy the exiled card." (CR 707.12; Isochron Scepter, Spellbinder),
// the enabling half of the "You may copy the exiled card. If you do, you may
// cast the copy without paying its mana cost." idiom. The exiled card is the one
// this source imprinted, identified at resolution through the source's imprint
// link rather than a bound object, so the effect carries no pronoun reference.
// The paired parseCastLinkedExiledCopyEffect casts the copy.
//
// The recognizer is intentionally narrow — it matches only the exact
// "(you may) copy the exiled card" wording — so it never reinterprets other
// "copy" effects. Any other wording fails closed and flows through the generic
// effect parser.
func parseCopyLinkedExiledCardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := wordsOnly(tokens)
	optional := false
	var optionalSpan shared.Span
	if rest, ok := cutTokenPrefix(words, "you", "may"); ok {
		optional = true
		optionalSpan = shared.Span{Start: words[0].Span.Start, End: words[1].Span.End}
		words = rest
	}
	if len(words) == 0 || !equalWord(words[0], "copy") {
		return nil, false
	}
	verbSpan := words[0].Span
	after, ok := cutTokenPrefix(words, "copy", "the", "exiled", "card")
	if !ok || len(after) != 0 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                 EffectCopyStackObject,
		Span:                 sentence.Span,
		ClauseSpan:           sentence.Span,
		VerbSpan:             verbSpan,
		Text:                 sentence.Text,
		Tokens:               append([]shared.Token(nil), tokens...),
		Context:              EffectContextController,
		Optional:             optional,
		OptionalSpan:         optionalSpan,
		CopyLinkedExiledCard: true,
		Exact:                true,
	}}, true
}

// parseCastLinkedExiledCopyEffect recognizes the imprint cast-the-copy effect
// "(If you do,) you may cast the copy without paying its mana cost." (CR 707.12;
// Isochron Scepter, Spellbinder), the consequence half of the imprint copy/cast
// idiom paired with parseCopyLinkedExiledCardEffect. The leading "If you do,"
// reflexive condition is parsed separately as the ability's prior-instruction
// gate; the "its" of "its mana cost" is consumed wholesale into the free-cast
// rider, so the effect emits no pronoun reference.
//
// The recognizer is intentionally narrow — it matches only the exact "(you may)
// cast the copy without paying its mana cost" wording — so it never reinterprets
// other "cast" effects. Any other wording fails closed and flows through the
// generic effect parser.
func parseCastLinkedExiledCopyEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	tokens = stripLeadingIfYouDoClause(tokens)
	words := wordsOnly(tokens)
	optional := false
	var optionalSpan shared.Span
	if rest, ok := cutTokenPrefix(words, "you", "may"); ok {
		optional = true
		optionalSpan = shared.Span{Start: words[0].Span.Start, End: words[1].Span.End}
		words = rest
	}
	if len(words) == 0 || !equalWord(words[0], "cast") {
		return nil, false
	}
	verbSpan := words[0].Span
	after, ok := cutTokenPrefix(words, "cast", "the", "copy", "without", "paying", "its", "mana", "cost")
	if !ok || len(after) != 0 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                      EffectCast,
		Span:                      sentence.Span,
		ClauseSpan:                sentence.Span,
		VerbSpan:                  verbSpan,
		Text:                      sentence.Text,
		Tokens:                    append([]shared.Token(nil), tokens...),
		Context:                   EffectContextController,
		Optional:                  optional,
		OptionalSpan:              optionalSpan,
		CastLinkedExiledCopy:      true,
		CastWithoutPayingManaCost: true,
		Exact:                     true,
	}}, true
}
