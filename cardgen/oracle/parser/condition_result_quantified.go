package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func recognizeQuantifiedResultCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest := body
	controller := startsWithWord(rest, "you")
	if controller {
		rest = rest[1:]
	}
	var noun []shared.Token
	var outcome EffectKind
	var cardDomain, ok bool
	if controller {
		if len(rest) < 4 {
			return ConditionClause{}, false
		}
		outcome, cardDomain, ok = resultThisWayActiveOutcomeKind(rest[0].Text)
		noun = rest[1 : len(rest)-2]
	} else {
		copula := -1
		for i, tok := range rest {
			if equalWord(tok, "is") || equalWord(tok, "are") || equalWord(tok, "was") || equalWord(tok, "were") {
				copula = i
				break
			}
		}
		if copula <= 0 || copula+4 != len(rest) {
			return ConditionClause{}, false
		}
		outcome, cardDomain, ok = resultThisWayOutcomeKind(rest[copula+1].Text)
		noun = rest[:copula]
	}
	if !ok || !tokenWordsEqual(rest[len(rest)-2:], "this", "way") {
		return ConditionClause{}, false
	}
	count, comparison := 0, ConditionComparisonNone
	switch {
	case startsWithWord(noun, "a"), startsWithWord(noun, "an"):
		noun = noun[1:]
	default:
		if parsed, tail, ok := parseLeadingCount(noun); ok {
			if parsed.Value <= 0 || parsed.Comparison != ConditionComparisonAtLeast {
				return ConditionClause{}, false
			}
			count, comparison, noun = parsed.Value, parsed.Comparison, tail
		} else if len(noun) > 0 {
			if value, ok := conditionNumberValue(noun[0]); ok {
				if value <= 0 {
					return ConditionClause{}, false
				}
				count, comparison, noun = value, ConditionComparisonNone, noun[1:]
			} else if !pluralResultNoun(noun) {
				return ConditionClause{}, false
			}
		}
	}
	if !validResultNounAtoms(noun, atoms) || !validResultThisWayNoun(noun, cardDomain, atoms) {
		return ConditionClause{}, false
	}

	selection := parseSelection(noun, atoms)
	return ConditionClause{
		Predicate: ConditionPredicateResultThisWay, ThisWayOutcome: outcome,
		ThisWaySelection: &selection, ThisWayController: controller,
		ThisWayCardNoun: equalWord(noun[len(noun)-1], "card") || equalWord(noun[len(noun)-1], "cards"),
		ThisWayCount:    count, ThisWayCountComparison: comparison,
	}, true
}

func pluralResultNoun(noun []shared.Token) bool {
	if len(noun) == 0 {
		return false
	}
	for _, word := range []string{"cards", "creatures", "artifacts", "enchantments", "lands", "permanents", "planeswalkers", "battles"} {
		if equalWord(noun[len(noun)-1], word) {
			return true
		}
	}
	return false
}

func validResultNounAtoms(noun []shared.Token, atoms Atoms) bool {
	if len(noun) >= 5 && tokenWordsEqual(noun[len(noun)-5:], "with", "one", "or", "more", "colors") {
		noun = noun[:len(noun)-5]
	}
	for _, token := range noun {
		if equalWord(token, "card") || equalWord(token, "cards") || equalWord(token, "or") || equalWord(token, "and") ||
			equalWord(token, "instant") || equalWord(token, "sorcery") {
			continue
		}
		atom := parseSelection([]shared.Token{token}, atoms)
		if atom.Kind == SelectionUnknown && len(atom.SubtypesAny) == 0 && len(atom.ColorsAny) == 0 &&
			len(atom.Supertypes) == 0 && len(atom.ExcludedTypes) == 0 && !atom.Colorless && !atom.Multicolored && !atom.Colored {
			return false
		}
	}
	return true
}

func resultThisWayActiveOutcomeKind(word string) (outcome EffectKind, cardDomain, recognized bool) {
	switch word {
	case "discard":
		return EffectDiscard, true, true
	case "exile":
		return EffectExile, true, true
	case "mill":
		return EffectMill, true, true
	case "sacrifice":
		return EffectSacrifice, false, true
	case "destroy":
		return EffectDestroy, false, true
	default:
		return resultThisWayOutcomeKind(word)
	}
}
