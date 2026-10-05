package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
)

// Numeric wording identifies a subject occurrence, never its runtime domain.
// The compiler binds that identity to a source, target, or triggering object.
func recognizeContextualAttributeCompareCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if len(body) < 5 {
		return ConditionClause{}, false
	}
	width := 1
	possessive := equalWord(body[0], "its")
	spell := false
	var subjectTypes []TriggerCardType
	if !possessive && !equalWord(body[0], "it") {
		if !equalWord(body[0], "that") && !equalWord(body[0], "this") {
			return ConditionClause{}, false
		}
		width = 2
		noun := strings.ToLower(body[1].Text)
		possessive = strings.HasSuffix(noun, "'s")
		noun = strings.TrimSuffix(noun, "'s")
		spell = noun == "spell"
		if !spell && !conditionAttributeComparePermanentNoun(noun) {
			return ConditionClause{}, false
		}
		if noun != "permanent" && !spell {
			cardType, ok := recognizeCardTypeWord(noun)
			if !ok {
				return ConditionClause{}, false
			}
			subjectTypes = []TriggerCardType{triggerCardTypeFromAtom(cardType)}
		}
	}
	rest := body[width:]
	past := false
	if !possessive {
		if len(rest) == 0 || !equalWord(rest[0], "has") && !equalWord(rest[0], "had") {
			return ConditionClause{}, false
		}
		past = equalWord(rest[0], "had")
		rest = rest[1:]
	}
	attribute, rest, ok := cutConditionAttributeWord(rest)
	if !ok {
		return ConditionClause{}, false
	}
	if possessive {
		if len(rest) == 0 || !equalWord(rest[0], "is") && !equalWord(rest[0], "was") {
			return ConditionClause{}, false
		}
		past = equalWord(rest[0], "was")
		rest = rest[1:]
	}
	comparison, ok := conditionFixedAttributeComparison(attribute, rest)
	if !ok {
		return ConditionClause{}, false
	}
	subject := shared.SpanOf(body[:width])
	binding := ConditionObjectBindingEventPermanent
	if equalWord(body[0], "this") {
		binding = ConditionObjectBindingSource
	}
	return ConditionClause{
		Predicate:      ConditionPredicateObjectMatches,
		ObjectBinding:  binding,
		Selection:      ConditionSelection{AttributeCompare: comparison},
		SubjectSpan:    subject,
		HasSubjectSpan: true,
		SubjectRefID:   atoms.ReferenceIDAt(subject),
		SubjectTypes:   subjectTypes,
		SubjectSpell:   spell,
		SubjectPast:    past,
	}, true
}

func conditionFixedAttributeComparison(attribute ConditionAttributeKind, tokens []shared.Token) (ConditionAttributeComparison, bool) {
	if len(tokens) != 3 || !equalWord(tokens[1], "or") {
		return ConditionAttributeComparison{}, false
	}
	value, ok := conditionNumberValue(tokens[0])
	if !ok {
		return ConditionAttributeComparison{}, false
	}
	var op compare.Op
	switch {
	case equalWord(tokens[2], "greater"):
		op = compare.GreaterOrEqual
	case equalWord(tokens[2], "less"):
		op = compare.LessOrEqual
	default:
		return ConditionAttributeComparison{}, false
	}
	return ConditionAttributeComparison{Attribute: attribute, Op: op, Value: value}, true
}
