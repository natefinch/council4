package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// parseReferencedTypeGrantEffect recognizes the permanent continuous
// type-and-color grant a rider applies to the creature an earlier clause in the
// same ability acted on ("That creature is a black Zombie in addition to its
// other colors and types." — Rise from the Grave, Liliana, Death's Majesty;
// "It's a Phyrexian in addition to its other types." — Portal to Phyrexia; "It
// becomes an Angel in addition to its other types." — Guide of Souls). The
// subject is a back-reference ("that creature", "the creature", "it", "it's")
// and the linking verb is "is" or "becomes"; the grant adds one or more colors
// and/or creature subtypes and card types to the permanent without removing its
// existing characteristics, and lasts as long as it remains on the battlefield
// (no "until end of turn" duration). It emits an EffectBecomeType whose
// referenced-object context and absent until-end-of-turn duration distinguish it
// from the targeted Liquimetal form in parseBecomeTypeEffect; lowering folds it
// into the preceding clause. Any other shape (a "target" subject, an
// until-end-of-turn duration, no added types, or an unrecognized color/type
// word) fails closed.
func parseReferencedTypeGrantEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := normalizedWords(body[:len(body)-1])
	rest, ok := stripReferencedTypeGrantSubject(words)
	if !ok {
		return nil, false
	}
	grant, ok := parseAdditiveTypeGrantBody(rest)
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                  EffectBecomeType,
		Context:               EffectContextReferencedObject,
		Span:                  sentence.Span,
		ClauseSpan:            sentence.Span,
		Text:                  sentence.Text,
		Tokens:                append([]shared.Token(nil), body...),
		BecomeTypeAddTypes:    grant.Types,
		BecomeTypeAddColors:   grant.Colors,
		BecomeTypeAddSubtypes: grant.Subtypes,
		References:            referencesInSpan(atoms, sentence.Span),
	}
	return []EffectSyntax{effect}, true
}

// additiveTypeGrant holds the colors, card types, and creature subtypes an
// additive "in addition to its other … types" grant confers without removing
// the object's existing characteristics.
type additiveTypeGrant struct {
	Colors   []Color
	Types    []types.Card
	Subtypes []types.Sub
}

// parseAdditiveTypeGrantBody parses the "a <color…> <type/subtype…> in addition
// to its other [colors and] [creature] types" body shared by the reanimation
// type-and-color grant riders, starting at the indefinite article that follows
// the subject's linking verb. It returns the added colors, card types, and
// creature subtypes the grant confers without removing existing
// characteristics. It fails closed when the body is not an additive type grant:
// a missing indefinite article, an "until end of turn" duration (a temporary
// Liquimetal-style change lowered through the ordinary target path), an absent
// additive "in addition to its other … types" tail, an unrecognized type or
// color word, or no added type at all.
func parseAdditiveTypeGrantBody(rest []string) (additiveTypeGrant, bool) {
	if len(rest) < 2 || (rest[0] != "a" && rest[0] != "an") {
		return additiveTypeGrant{}, false
	}
	rest = rest[1:]
	duration := []string{"until", "end", "of", "turn"}
	if len(rest) >= len(duration) && slices.Equal(rest[len(rest)-len(duration):], duration) {
		return additiveTypeGrant{}, false
	}
	additiveTypes := []string{"in", "addition", "to", "its", "other", "types"}
	additiveCreatureTypes := []string{"in", "addition", "to", "its", "other", "creature", "types"}
	additiveColorsTypes := []string{"in", "addition", "to", "its", "other", "colors", "and", "types"}
	switch {
	case len(rest) > len(additiveColorsTypes) &&
		slices.Equal(rest[len(rest)-len(additiveColorsTypes):], additiveColorsTypes):
		rest = rest[:len(rest)-len(additiveColorsTypes)]
	case len(rest) > len(additiveCreatureTypes) &&
		slices.Equal(rest[len(rest)-len(additiveCreatureTypes):], additiveCreatureTypes):
		rest = rest[:len(rest)-len(additiveCreatureTypes)]
	case len(rest) > len(additiveTypes) &&
		slices.Equal(rest[len(rest)-len(additiveTypes):], additiveTypes):
		rest = rest[:len(rest)-len(additiveTypes)]
	default:
		return additiveTypeGrant{}, false
	}
	addColors := make([]Color, 0)
	for len(rest) > 0 {
		parsedColor, colorOK := recognizeColorWord(rest[0])
		if !colorOK {
			break
		}
		addColors = append(addColors, parsedColor)
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return additiveTypeGrant{}, false
	}
	addTypes := make([]types.Card, 0, len(rest))
	addSubtypes := make([]types.Sub, 0, len(rest))
	for _, word := range rest {
		if cardType, typeOK := entersAsCopyAddTypeWord(word); typeOK {
			addTypes = append(addTypes, cardType)
			continue
		}
		if sub, subOK := recognizeSubtypePhrase(word); subOK {
			addSubtypes = append(addSubtypes, sub)
			continue
		}
		return additiveTypeGrant{}, false
	}
	if len(addTypes) == 0 && len(addSubtypes) == 0 {
		return additiveTypeGrant{}, false
	}
	return additiveTypeGrant{Colors: addColors, Types: addTypes, Subtypes: addSubtypes}, true
}

// stripReferencedTypeGrantSubject consumes the back-reference subject and its
// linking verb ("is" or "becomes") from a referenced-object type-grant rider,
// returning the words after the verb. It recognizes "that creature is/becomes",
// "the creature is/becomes", "it is/becomes", and the "it's" contraction; any
// other subject fails closed.
func stripReferencedTypeGrantSubject(words []string) ([]string, bool) {
	switch {
	case len(words) >= 1 && words[0] == "it's":
		return words[1:], true
	case len(words) >= 2 && words[0] == "it" && referencedTypeGrantVerb(words[1]):
		return words[2:], true
	case len(words) >= 3 && (words[0] == "that" || words[0] == "the") &&
		words[1] == "creature" && referencedTypeGrantVerb(words[2]):
		return words[3:], true
	}
	return nil, false
}

// referencedTypeGrantVerb reports whether a word is the linking verb of a
// referenced-object type grant: the stative "is" (Rise from the Grave) or the
// inchoative "becomes" (Guide of Souls). Both grant the added types for the
// permanent's lifetime on the battlefield.
func referencedTypeGrantVerb(word string) bool {
	return word == "is" || word == "becomes"
}
