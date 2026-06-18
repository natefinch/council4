package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// counterAbilityStackKinds records which stack-object kinds a counter-ability
// target list allows.
type counterAbilityStackKinds struct {
	activated bool
	triggered bool
	spell     bool
}

// counterAbilitySelectionSyntax recognizes qualified and mixed ability-counter
// target phrases that the plain ability selectors do not cover, such as
// "activated ability from an artifact source", "activated or triggered ability
// you don't control", and "activated ability, triggered ability, or legendary
// spell". It returns false for plain, unqualified ability lists so the existing
// exact selectors keep owning them, and fails closed for any unrecognized
// qualifier so unsupported wordings stay rejected.
func counterAbilitySelectionSyntax(tokens []shared.Token, span shared.Span, text string) (SelectionSyntax, bool) {
	words := normalizedWords(tokens)

	words, controller, hadController := stripCounterControllerClause(words)
	words, sourceType, hadSource := stripCounterSourceClause(words)

	kinds, qualifier, ok := parseCounterAbilityKindList(words)
	if !ok {
		return SelectionSyntax{}, false
	}

	qualified := hadController || hadSource || qualifier.colorless ||
		len(qualifier.supertypes) > 0
	if !qualified {
		return SelectionSyntax{}, false
	}

	kind, ok := counterAbilitySelectionKind(kinds)
	if !ok {
		return SelectionSyntax{}, false
	}
	// Source restrictions only apply to single-kind ability targets in the
	// supported corpus; mixed lists never combine with a source clause.
	if hadSource && (kinds.spell || (kinds.activated && kinds.triggered)) {
		return SelectionSyntax{}, false
	}
	// Spell qualifiers only make sense when a spell is among the allowed kinds.
	if (qualifier.colorless || len(qualifier.supertypes) > 0) && !kinds.spell {
		return SelectionSyntax{}, false
	}

	selection := SelectionSyntax{Span: span, Text: text, Kind: kind}
	if hadController {
		selection.Controller = controller
	}
	if hadSource {
		selection.SourceTypes = []CardType{sourceType}
	}
	selection.Supertypes = qualifier.supertypes
	selection.Colorless = qualifier.colorless
	return selection, true
}

type counterSpellQualifier struct {
	supertypes []Supertype
	colorless  bool
}

// parseCounterAbilityKindList parses an "or"-separated list of stack-object
// kinds, where the spell element may carry supertype or colorless qualifiers.
// Commas are already dropped by normalizedWords, so elements are separated only
// by "or" (or by nothing, as in "activated ability triggered ability").
func parseCounterAbilityKindList(words []string) (counterAbilityStackKinds, counterSpellQualifier, bool) {
	var kinds counterAbilityStackKinds
	var qualifier counterSpellQualifier
	seenSpell := false
	i := 0
	for i < len(words) {
		switch {
		case words[i] == "or":
			i++
		case i+1 < len(words) && words[i] == "activated" && words[i+1] == "ability":
			if kinds.activated {
				return counterAbilityStackKinds{}, counterSpellQualifier{}, false
			}
			kinds.activated = true
			i += 2
		case i+1 < len(words) && words[i] == "triggered" && words[i+1] == "ability":
			if kinds.triggered {
				return counterAbilityStackKinds{}, counterSpellQualifier{}, false
			}
			kinds.triggered = true
			i += 2
		default:
			next, ok := parseCounterSpellElement(words[i:], &qualifier)
			if !ok || seenSpell {
				return counterAbilityStackKinds{}, counterSpellQualifier{}, false
			}
			kinds.spell = true
			seenSpell = true
			i += next
		}
	}
	if !kinds.activated && !kinds.triggered && !kinds.spell {
		return counterAbilityStackKinds{}, counterSpellQualifier{}, false
	}
	return kinds, qualifier, true
}

// parseCounterSpellElement consumes a "[qualifier...] spell" element, recording
// supertype and colorless qualifiers. It fails closed on any qualifier word that
// is not a known supertype or the colorless color qualifier.
func parseCounterSpellElement(words []string, qualifier *counterSpellQualifier) (int, bool) {
	i := 0
	for i < len(words) && words[i] != "spell" {
		if words[i] == "or" {
			return 0, false
		}
		if supertype, ok := recognizeSupertypeWord(words[i]); ok {
			if !slices.Contains(qualifier.supertypes, supertype) {
				qualifier.supertypes = append(qualifier.supertypes, supertype)
			}
			i++
			continue
		}
		if colorQualifier, ok := recognizeColorQualifierWord(words[i]); ok && colorQualifier == ColorQualifierColorless {
			qualifier.colorless = true
			i++
			continue
		}
		return 0, false
	}
	if i >= len(words) || words[i] != "spell" {
		return 0, false
	}
	return i + 1, true
}

func counterAbilitySelectionKind(kinds counterAbilityStackKinds) (SelectionKind, bool) {
	switch {
	case kinds.spell && kinds.activated && kinds.triggered:
		return SelectionSpellActivatedOrTriggeredAbility, true
	case kinds.spell && kinds.triggered && !kinds.activated:
		return SelectionTriggeredAbilityOrSpell, true
	case kinds.activated && kinds.triggered && !kinds.spell:
		return SelectionActivatedOrTriggeredAbility, true
	case kinds.activated && !kinds.triggered && !kinds.spell:
		return SelectionActivatedAbility, true
	case kinds.triggered && !kinds.activated && !kinds.spell:
		return SelectionTriggeredAbility, true
	default:
		return SelectionUnknown, false
	}
}

func stripCounterControllerClause(words []string) ([]string, SelectionController, bool) {
	switch {
	case hasWordSuffix(words, "you", "don't", "control"):
		return words[:len(words)-3], SelectionControllerNotYou, true
	case hasWordSuffix(words, "you", "control"):
		return words[:len(words)-2], SelectionControllerYou, true
	case hasWordSuffix(words, "an", "opponent", "controls"):
		return words[:len(words)-3], SelectionControllerOpponent, true
	default:
		return words, SelectionControllerAny, false
	}
}

func stripCounterSourceClause(words []string) ([]string, CardType, bool) {
	if len(words) < 4 {
		return words, CardTypeUnknown, false
	}
	last := len(words) - 1
	if words[last] != "source" || (words[last-2] != "a" && words[last-2] != "an") || words[last-3] != "from" {
		return words, CardTypeUnknown, false
	}
	cardType, ok := recognizeCardTypeWord(words[last-1])
	if !ok {
		return words, CardTypeUnknown, false
	}
	return words[:last-3], cardType, true
}

func hasWordSuffix(words []string, suffix ...string) bool {
	if len(words) < len(suffix) {
		return false
	}
	return slices.Equal(words[len(words)-len(suffix):], suffix)
}

// selectionHasCounterAbilityQualifier reports whether a selection carries one of
// the qualifiers recognized only for ability-counter targets, marking it as a
// fully round-tripped qualified counter target for exactness.
func selectionHasCounterAbilityQualifier(selection SelectionSyntax) bool {
	return selection.Controller != SelectionControllerAny ||
		len(selection.SourceTypes) > 0 ||
		len(selection.Supertypes) > 0 ||
		selection.Colorless
}

// counterAbilityListEnd extends a target span across a comma-separated
// ability-counter kind list such as "activated ability, triggered ability, or
// legendary spell". It only fires for lists containing at least one ability kind
// and a comma, so non-counter comma phrases keep their existing boundaries.
func counterAbilityListEnd(tokens []shared.Token, start int) (int, bool) {
	i := start
	elements := 0
	abilityElements := 0
	sawComma := false
	for i < len(tokens) {
		switch {
		case effectWordsAt(tokens, i, "activated", "ability"),
			effectWordsAt(tokens, i, "triggered", "ability"):
			i += 2
			elements++
			abilityElements++
		default:
			next, ok := counterSpellElementEnd(tokens, i)
			if !ok {
				if elements >= 2 && abilityElements >= 1 && sawComma {
					return i, true
				}
				return start, false
			}
			i = next
			elements++
		}
		consumedSeparator := false
		if i < len(tokens) && tokens[i].Kind == shared.Comma {
			sawComma = true
			i++
			consumedSeparator = true
		}
		if i < len(tokens) && equalWord(tokens[i], "or") {
			i++
			consumedSeparator = true
		}
		if !consumedSeparator {
			break
		}
	}
	if elements >= 2 && abilityElements >= 1 && sawComma {
		return i, true
	}
	return start, false
}

func counterSpellElementEnd(tokens []shared.Token, start int) (int, bool) {
	i := start
	for i < len(tokens) && tokens[i].Kind == shared.Word && !equalWord(tokens[i], "spell") && !equalWord(tokens[i], "or") {
		if _, ok := recognizeSupertypeWord(tokens[i].Text); ok {
			i++
			continue
		}
		if colorQualifier, ok := recognizeColorQualifierWord(tokens[i].Text); ok && colorQualifier == ColorQualifierColorless {
			i++
			continue
		}
		return 0, false
	}
	if i < len(tokens) && equalWord(tokens[i], "spell") {
		return i + 1, true
	}
	return 0, false
}
