package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// staticGroupDoesntUntapSubject recognizes the battlefield-wide creature group
// frozen by a mass "don't untap during their controllers' untap steps."
// restriction ("Creatures ...", "Red creatures ...", "Mercenaries ...",
// "Creatures with power 3 or greater ...") when it precedes the plural untap
// verb "don't". The group is every matching creature on the battlefield, named
// either by the bare "creatures" noun (carrying an optional leading color filter
// and an optional trailing "with power/toughness <comparison>" qualifier) or by
// a creature subtype plural ("Mercenaries"). Unlike the can't-block subject
// parsers this one is gated on "don't", so it never shadows an existing
// prohibition shape. It returns the group subject and the index of the "don't"
// verb that follows the noun phrase.
func staticGroupDoesntUntapSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, int, bool) {
	if subject, verb, ok := staticAllCreaturesDoesntUntapSubject(tokens, atoms); ok {
		return subject, verb, true
	}
	if subject, verb, ok := staticNonbasicLandsDoesntUntapSubject(tokens); ok {
		return subject, verb, true
	}
	if subject, verb, ok := staticNonlandPermanentsDoesntUntapSubject(tokens); ok {
		return subject, verb, true
	}
	if subject, verb, ok := staticSnowPermanentsDoesntUntapSubject(tokens); ok {
		return subject, verb, true
	}
	if subject, verb, ok := staticSubtypeDoesntUntapSubject(tokens, atoms); ok {
		return subject, verb, true
	}
	return staticLandSubtypeDoesntUntapSubject(tokens, atoms)
}

// staticNonbasicLandsDoesntUntapSubject recognizes the "Nonbasic lands don't"
// head (Back to Basics), mapping it onto the battlefield-wide nonbasic-land
// group. The leading word must be the single "nonbasic" supertype-exclusion glyph
// followed by the "lands" plural and the untap verb "don't".
func staticNonbasicLandsDoesntUntapSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, int, bool) {
	if !staticWordsAt(tokens, 0, "nonbasic", "lands", "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	return EffectStaticSubjectSyntax{
		Kind: EffectStaticSubjectNonbasicLands,
		Span: shared.SpanOf(tokens[:2]),
	}, 2, true
}

// staticNonlandPermanentsDoesntUntapSubject recognizes the "Nonland permanents
// don't" head (Embargo), mapping it onto the battlefield-wide nonland-permanent
// group.
func staticNonlandPermanentsDoesntUntapSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, int, bool) {
	if !staticWordsAt(tokens, 0, "nonland", "permanents", "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	return EffectStaticSubjectSyntax{
		Kind: EffectStaticSubjectNonlandPermanents,
		Span: shared.SpanOf(tokens[:2]),
	}, 2, true
}

// staticSnowPermanentsDoesntUntapSubject recognizes the "Snow permanents don't"
// head (Freyalise's Radiance), mapping it onto the battlefield-wide
// snow-permanent group.
func staticSnowPermanentsDoesntUntapSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, int, bool) {
	if !staticWordsAt(tokens, 0, "snow", "permanents", "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	return EffectStaticSubjectSyntax{
		Kind: EffectStaticSubjectSnowPermanents,
		Span: shared.SpanOf(tokens[:2]),
	}, 2, true
}

// staticLandSubtypeDoesntUntapSubject recognizes the "<land subtype plural>
// don't" head ("Islands don't ...", Choke), mapping it onto the battlefield-wide
// permanent-subtype group. Only land subtypes are accepted here; creature
// subtypes are owned by staticSubtypeDoesntUntapSubject, which is tried first.
func staticLandSubtypeDoesntUntapSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, int, bool) {
	if len(tokens) == 0 {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	subtype, ok := atoms.SubtypeAt(tokens[0].Span)
	if !ok || !SubtypeMatchesAnyRuntimeCardType(subtype, []types.Card{types.Land}) {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	if !staticWordsAt(tokens, 1, "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	return EffectStaticSubjectSyntax{
		Kind:         EffectStaticSubjectAllPermanentSubtype,
		Span:         shared.SpanOf(tokens[:1]),
		Subtype:      subtype,
		SubtypeText:  tokens[0].Text,
		SubtypeKnown: true,
	}, 1, true
}

// staticAllCreaturesDoesntUntapSubject recognizes the "[<color>] creatures
// [with <power/toughness comparison>] don't" head, mapping it onto the
// all-creatures group narrowed by the optional color and power/toughness
// filters.
func staticAllCreaturesDoesntUntapSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, int, bool) {
	subject := EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures}
	idx := 0
	if filter, width, ok := staticColorFilterAt(tokens, idx); ok {
		subject.Colors = filter.colors
		subject.Colorless = filter.colorless
		subject.Multicolored = filter.multicolored
		idx += width
	}
	if !staticWordsAt(tokens, idx, "creatures") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	idx++
	if match, ok := controlledGroupProhibitionPowerToughnessQualifier(tokens, idx, atoms); ok {
		subject.Power = match.power
		subject.MatchPower = match.matchPower
		subject.Toughness = match.toughness
		subject.MatchToughness = match.matchToughness
		subject.PowerOrToughness = match.powerOrToughness
		subject.PowerLessThanSource = match.powerLessThanSource
		subject.PowerGreaterThanSource = match.powerGreaterThanSource
		idx = match.end
	} else if match, ok := counterQualifierKind(tokens, idx); ok && subject.applyCounterQualifier(match) {
		idx = match.End
	}
	if !staticWordsAt(tokens, idx, "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	subject.Span = shared.SpanOf(tokens[:idx])
	return subject, idx, true
}

// staticSubtypeDoesntUntapSubject recognizes the "<creature subtype plural>
// don't" head ("Mercenaries don't ..."), mapping it onto the battlefield-wide
// creature-subtype group. A leading color filter is not accepted here; the
// subtype plural names the whole group.
func staticSubtypeDoesntUntapSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, int, bool) {
	if len(tokens) == 0 {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	subtype, ok := atoms.SubtypeAt(tokens[0].Span)
	if !ok || !SubtypeMatchesAnyRuntimeCardType(subtype, []types.Card{types.Creature, types.Kindred}) {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	if !staticWordsAt(tokens, 1, "don't") {
		return EffectStaticSubjectSyntax{}, 0, false
	}
	return EffectStaticSubjectSyntax{
		Kind:         EffectStaticSubjectAllCreatureSubtype,
		Span:         shared.SpanOf(tokens[:1]),
		Subtype:      subtype,
		SubtypeText:  tokens[0].Text,
		SubtypeKnown: true,
	}, 1, true
}

// parseStaticGroupDoesntUntapRuleOperation recognizes the mass untap-prohibition
// tail "don't untap during their controllers' untap steps." beginning at index
// (the "don't" verb) for a battlefield-wide creature group subject. The plural
// "their controllers' untap steps" phrasing is fixed and fully consumed; any
// deviation fails closed. It is the group counterpart of the singular
// parseStaticDoesntUntapRule.
func parseStaticGroupDoesntUntapRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if !staticGroupDoesntUntapSubjectKind(subject) {
		return StaticDeclarationSyntax{}, 0, false
	}
	if !staticWordsAt(tokens, index, "don't", "untap", "during", "their", "controllers") {
		return StaticDeclarationSyntax{}, 0, false
	}
	apostrophe := index + 5
	if apostrophe >= len(tokens) || tokens[apostrophe].Kind != shared.Apostrophe {
		return StaticDeclarationSyntax{}, 0, false
	}
	if !staticWordsAt(tokens, apostrophe+1, "untap", "steps") {
		return StaticDeclarationSyntax{}, 0, false
	}
	next := apostrophe + 3
	if next != end {
		return StaticDeclarationSyntax{}, 0, false
	}
	constraint := StaticRuleConstraint{Kind: StaticRuleConstraintProhibition, Span: tokens[index].Span}
	operation := StaticRuleOperation{
		Kind:  StaticRuleOperationUntap,
		Voice: StaticRuleVoiceActive,
		Span:  shared.SpanOf(tokens[index+1 : next]),
	}
	return staticRuleOperation(tokens, index, next, subject, constraint, operation, nil)
}

// staticGroupDoesntUntapSubjectKind reports whether subject is one of the
// battlefield-wide creature groups the mass untap prohibition recognizes.
func staticGroupDoesntUntapSubjectKind(subject StaticDeclarationSubject) bool {
	if subject.Kind != StaticDeclarationSubjectGroup {
		return false
	}
	switch subject.Group.Kind {
	case EffectStaticSubjectAllCreatures, EffectStaticSubjectAllCreatureSubtype,
		EffectStaticSubjectNonbasicLands, EffectStaticSubjectNonlandPermanents,
		EffectStaticSubjectSnowPermanents, EffectStaticSubjectAllPermanentSubtype:
		return true
	default:
		return false
	}
}
