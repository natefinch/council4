package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// parsePluralSubtypeGroupSubject recognizes a static anthem group named directly
// by one or more plural creature-subtype nouns, with no "creatures" head noun:
//
//	"[Other] <Subtype>s you control get/have ..."          (controlled)
//	"[Other|All] <Subtype>s get/have ..."                  (battlefield-wide)
//	"<Sub1>s and <Sub2>s you control get/have ..."         (conjunction)
//	"<Sub1>s, <Sub2>s, and <Sub3>s you control get/have ..."
//	"Other <Sub1>s and <Sub2>s get/have ..."               (battlefield conjunction)
//
// Each named subtype rides SubtypesAny, so the affected group matches a permanent
// that has any one of them, exactly like the single-subtype "Other <Sub> creatures
// you control" group already does. A leading "Other" maps to the source-excluding
// subject kind. A leading "All" is the battlefield-wide non-excluding marker.
//
// To avoid shadowing the existing single-subtype controlled productions (which
// already own "<Subtype>s you control get ..." and "Other <Subtype>s you control
// get ..."), this recognizer fires for a controlled group only when the
// conjunction names two or more subtypes. The battlefield-wide bare-plural forms
// (with or without a conjunction) have no existing production, so they are
// admitted at any length.
func parsePluralSubtypeGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	idx := 0
	excluded := false
	switch {
	case len(tokens) > 0 && equalWord(tokens[0], "other"):
		excluded = true
		idx = 1
	case len(tokens) > 0 && equalWord(tokens[0], "all"):
		idx = 1
	default:
		idx = 0
	}
	subs, next, ok := parseCreatureSubtypeConjunctionList(tokens, idx, atoms)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	controlled := false
	if effectWordsAt(tokens, next, "you", "control") {
		controlled = true
		next += 2
	}
	if next >= len(tokens) || !staticGroupVerb(tokens[next]) {
		return EffectStaticSubjectSyntax{}, false
	}
	var kind EffectStaticSubjectKind
	switch {
	case controlled:
		// Defer single-subtype controlled groups to the established
		// single-subtype productions so their output is unchanged.
		if len(subs) < 2 {
			return EffectStaticSubjectSyntax{}, false
		}
		kind = EffectStaticSubjectControlledCreatureSubtype
		if excluded {
			kind = EffectStaticSubjectOtherControlledCreatureSubtype
		}
	default:
		kind = EffectStaticSubjectAllCreatureSubtype
		if excluded {
			kind = EffectStaticSubjectOtherCreatureSubtype
		}
	}
	return EffectStaticSubjectSyntax{
		Kind:         kind,
		Span:         shared.SpanOf(tokens[:next]),
		Subtype:      subs[0],
		SubtypeText:  string(subs[0]),
		SubtypeKnown: true,
		SubtypesAny:  subs,
	}, true
}

// parseCreatureSubtypeConjunctionList parses a list of one or more creature
// subtypes beginning at start, joined by "and" and optional Oxford commas
// ("Skeletons and Zombies", "Skeletons, Vampires, and Zombies", "Goblins"). Each
// element is a plural creature-subtype noun the atom layer resolved to its
// canonical typed identity. It returns the resolved subtypes and the token index
// just past the list, failing closed unless every element resolves to a creature
// or kindred subtype.
func parseCreatureSubtypeConjunctionList(tokens []shared.Token, start int, atoms Atoms) ([]types.Sub, int, bool) {
	var subs []types.Sub
	idx := start
	for {
		sub, width, ok := staticSubtypeAt(tokens, idx, len(tokens), atoms)
		if !ok || !SubtypeMatchesAnyRuntimeCardType(sub, []types.Card{types.Creature, types.Kindred}) {
			return nil, 0, false
		}
		subs = append(subs, sub)
		idx += width
		switch {
		case idx < len(tokens) && tokens[idx].Kind == shared.Comma:
			idx++
			if idx < len(tokens) && equalWord(tokens[idx], "and") {
				idx++
			}
		case idx < len(tokens) && equalWord(tokens[idx], "and"):
			idx++
		default:
			return subs, idx, true
		}
	}
}
