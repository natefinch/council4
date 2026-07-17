package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseGroupControlledBecomeTypeEffect recognizes the resolving group type-grant
// a reanimation spell applies to every creature its controller controls once the
// returned cards have entered ("Then each creature you control becomes a
// Phyrexian in addition to its other types." — Breach the Multiverse). Unlike the
// targeted Liquimetal form (parseBecomeTypeEffect) and the single back-reference
// rider (parseReferencedTypeGrantEffect), the subject is the controlled-creature
// group "each creature you control"; the effect adds one or more colors, card
// types, and/or creature subtypes to each snapshotted member without removing
// its existing characteristics, and lasts for those permanents' lifetime on the
// battlefield (no "until end of turn" duration). It emits an EffectBecomeType
// carrying the controlled-creatures static subject; lowering expands it into a
// per-member continuous grant over the group snapshotted at resolution. Any other
// shape (a different subject, a plural "become <types>" without an indefinite
// article, an until-end-of-turn duration, or an unrecognized color/type word)
// fails closed.
func parseGroupControlledBecomeTypeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	content := body[:len(body)-1]
	// A "Then" connective links this clause to the preceding reanimation put in
	// the same sentence sequence; it carries no meaning beyond ordering, so drop
	// it before matching the group subject.
	if len(content) > 0 && equalWord(content[0], "then") {
		content = content[1:]
	}
	if len(content) < 6 ||
		!effectWordsAt(content, 0, "each", "creature", "you", "control") ||
		!equalWord(content[4], "becomes") {
		return nil, false
	}
	subjectSpan := shared.SpanOf(content[:4])
	grant, ok := parseAdditiveTypeGrantBody(normalizedWords(content[5:]))
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectBecomeType,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), body...),
		StaticSubject: EffectStaticSubjectSyntax{
			Kind: EffectStaticSubjectControlledCreatures,
			Span: subjectSpan,
		},
		BecomeTypeAddTypes:    grant.Types,
		BecomeTypeAddColors:   grant.Colors,
		BecomeTypeAddSubtypes: grant.Subtypes,
		References:            referencesInSpan(atoms, sentence.Span),
	}
	return []EffectSyntax{effect}, true
}
