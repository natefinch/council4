package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// splitSpellTargetRestrictionTail detects a trailing "that targets <X>" relative
// clause on a counter spell target ("spell that targets a permanent you
// control") and splits it from the selection's head tokens. The clause is owned
// here, not by the runtime, so the head ("spell") flows through parseSelection
// as an ordinary spell selection while the restriction is captured as typed
// alternatives. It fires only when the clause directly follows the noun "spell"
// and every restriction alternative parses, so any unrecognized wording leaves
// the tokens intact and the selection unsupported, preserving fail-closed
// behavior.
func splitSpellTargetRestrictionTail(tokens []shared.Token) (head []shared.Token, restrictions []SpellTargetRestriction, ok bool) {
	for i := 0; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "that") || !equalWord(tokens[i+1], "targets") {
			continue
		}
		if i == 0 || !equalWord(tokens[i-1], "spell") {
			return nil, nil, false
		}
		parsed, parsedOK := parseSpellTargetRestrictionElements(tokens[i+2:])
		if !parsedOK {
			return nil, nil, false
		}
		return tokens[:i], parsed, true
	}
	return nil, nil, false
}

// trimSpellTargetRestrictionTail returns the tokens preceding a "spell that
// targets <X>" restriction clause, dropping the relative clause that qualifies
// the target spell. The articles and nouns inside the restriction ("a creature",
// "a player") belong to the target's filter, not to the counter effect's amount,
// so amount parsing must not see them. It mirrors splitSpellTargetRestrictionTail
// and leaves the tokens untouched when no such clause is present.
func trimSpellTargetRestrictionTail(tokens []shared.Token) []shared.Token {
	for i := 0; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "that") || !equalWord(tokens[i+1], "targets") {
			continue
		}
		if i == 0 || !equalWord(tokens[i-1], "spell") {
			return tokens
		}
		return tokens[:i]
	}
	return tokens
}

// parseSpellTargetRestrictionElements parses one or more "that targets"
// alternatives joined by "or" ("a permanent you control", "you or a permanent
// you control"). It fails closed on any unrecognized element or separator so the
// whole restriction is rejected rather than partially recognized.
func parseSpellTargetRestrictionElements(tokens []shared.Token) ([]SpellTargetRestriction, bool) {
	words := normalizedWords(tokens)
	if len(words) != len(tokens) {
		// A punctuation token inside the clause is not part of the supported
		// "that targets <X>" wording; reject so the selection stays unsupported.
		return nil, false
	}
	var restrictions []SpellTargetRestriction
	i := 0
	for i < len(words) {
		restriction, consumed, restrictionOK := parseSpellTargetRestrictionElement(words[i:])
		if !restrictionOK {
			return nil, false
		}
		restrictions = append(restrictions, restriction)
		i += consumed
		if i >= len(words) {
			break
		}
		if words[i] != "or" {
			return nil, false
		}
		i++
	}
	if len(restrictions) == 0 {
		return nil, false
	}
	return restrictions, true
}

// parseSpellTargetRestrictionElement parses one restriction alternative and
// returns the number of words it consumed. It recognizes "you", "a player", and
// "a[n] <permanent noun>" optionally followed by a "you control" or "an opponent
// controls" controller clause.
func parseSpellTargetRestrictionElement(words []string) (SpellTargetRestriction, int, bool) {
	if words[0] == "you" {
		return SpellTargetRestriction{Kind: SpellTargetRestrictionPlayer, Controller: SelectionControllerYou}, 1, true
	}
	if words[0] != "a" && words[0] != "an" {
		return SpellTargetRestriction{}, 0, false
	}
	if len(words) < 2 {
		return SpellTargetRestriction{}, 0, false
	}
	if words[1] == "player" {
		return SpellTargetRestriction{Kind: SpellTargetRestrictionPlayer, Controller: SelectionControllerAny}, 2, true
	}
	var permanentType CardType
	if words[1] != "permanent" {
		cardType, typeOK := recognizeCardTypeWord(words[1])
		if !typeOK {
			return SpellTargetRestriction{}, 0, false
		}
		permanentType = cardType
	}
	consumed := 2
	controller := SelectionControllerAny
	rest := words[2:]
	switch {
	case len(rest) >= 2 && rest[0] == "you" && rest[1] == "control":
		controller = SelectionControllerYou
		consumed += 2
	case len(rest) >= 3 && rest[0] == "an" && rest[1] == "opponent" && rest[2] == "controls":
		controller = SelectionControllerOpponent
		consumed += 3
	default:
	}
	return SpellTargetRestriction{Kind: SpellTargetRestrictionPermanent, PermanentType: permanentType, Controller: controller}, consumed, true
}

// spellTargetRestrictionsClause reconstructs the canonical " that targets <X>"
// Oracle suffix from typed restrictions, joining alternatives with " or ". It
// fails closed for any restriction it cannot render so the byte-exact target
// round-trip stays honest.
func spellTargetRestrictionsClause(restrictions []SpellTargetRestriction) (string, bool) {
	if len(restrictions) == 0 {
		return "", false
	}
	parts := make([]string, 0, len(restrictions))
	for _, restriction := range restrictions {
		text, textOK := spellTargetRestrictionElementText(restriction)
		if !textOK {
			return "", false
		}
		parts = append(parts, text)
	}
	return " that targets " + strings.Join(parts, " or "), true
}

// spellTargetRestrictionElementText reconstructs one restriction alternative's
// canonical Oracle phrase.
func spellTargetRestrictionElementText(restriction SpellTargetRestriction) (string, bool) {
	switch restriction.Kind {
	case SpellTargetRestrictionPlayer:
		switch restriction.Controller {
		case SelectionControllerYou:
			return "you", true
		case SelectionControllerAny:
			return "a player", true
		default:
			return "", false
		}
	case SpellTargetRestrictionPermanent:
		noun := "permanent"
		if restriction.PermanentType != CardTypeUnknown {
			typeNoun, nounOK := permanentCardTypeNoun(restriction.PermanentType)
			if !nounOK {
				return "", false
			}
			noun = typeNoun
		}
		text := indefiniteArticle(noun) + " " + noun
		switch restriction.Controller {
		case SelectionControllerAny:
		case SelectionControllerYou:
			text += " you control"
		case SelectionControllerOpponent:
			text += " an opponent controls"
		default:
			return "", false
		}
		return text, true
	default:
		return "", false
	}
}
