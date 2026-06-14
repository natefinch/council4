package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseStaticRuleSyntax(tokens []shared.Token) (*StaticRuleSyntax, bool) {
	if len(tokens) < 5 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	subject, next, ok := parseStaticRuleSubject(tokens)
	if !ok {
		return nil, false
	}
	rule := &StaticRuleSyntax{
		Span:    shared.SpanOf(tokens),
		Subject: subject,
	}
	if constraint, ok := parseStaticRuleProhibition(tokens, next); ok {
		rule.Constraint = constraint
		next++
		operation, next, ok := parseProhibitedStaticRuleOperation(tokens, next)
		if !ok || next != len(tokens)-1 {
			return nil, false
		}
		rule.Operation = operation
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if attack, ok := parseRequiredAttackRule(tokens, next); ok {
		rule.Constraint = attack.Constraint
		rule.Operation = attack.Operation
		rule.Qualifiers = attack.Qualifiers
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	return nil, false
}

func parseStaticRuleSubject(tokens []shared.Token) (StaticRuleSubject, int, bool) {
	if !staticRuleWordsAt(tokens, 0, "this") || len(tokens) < 2 {
		return StaticRuleSubject{}, 0, false
	}
	subject := StaticRuleSubject{Span: shared.SpanOf(tokens[:2])}
	switch {
	case staticRuleWordsAt(tokens, 1, "creature"):
		subject.Kind = StaticRuleSubjectSourceCreature
	case staticRuleWordsAt(tokens, 1, "spell"):
		subject.Kind = StaticRuleSubjectSourceSpell
	default:
		return StaticRuleSubject{}, 0, false
	}
	return subject, 2, true
}

func parseStaticRuleProhibition(tokens []shared.Token, start int) (StaticRuleConstraint, bool) {
	if !staticRuleWordsAt(tokens, start, "can't") && !staticRuleWordsAt(tokens, start, "cannot") {
		return StaticRuleConstraint{}, false
	}
	return StaticRuleConstraint{
		Kind: StaticRuleConstraintProhibition,
		Span: tokens[start].Span,
	}, true
}

func parseProhibitedStaticRuleOperation(tokens []shared.Token, start int) (StaticRuleOperation, int, bool) {
	if staticRuleWordsAt(tokens, start, "block") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start].Span,
		}, start + 1, true
	}
	if staticRuleWordsAt(tokens, start, "be", "blocked") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start : start+2]),
		}, start + 2, true
	}
	if staticRuleWordsAt(tokens, start, "be", "countered") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationCounter,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start : start+2]),
		}, start + 2, true
	}
	return StaticRuleOperation{}, start, false
}

type requiredAttackRuleSyntax struct {
	Constraint StaticRuleConstraint  `json:",omitzero"`
	Operation  StaticRuleOperation   `json:",omitzero"`
	Qualifiers []StaticRuleQualifier `json:",omitempty"`
}

func parseRequiredAttackRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, bool) {
	constraintStart := start
	operationStart := start
	explicit := staticRuleWordsAt(tokens, start, "must")
	if explicit {
		operationStart++
	}
	if (explicit && !staticRuleWordsAt(tokens, operationStart, "attack")) ||
		(!explicit && !staticRuleWordsAt(tokens, operationStart, "attacks")) {
		return requiredAttackRuleSyntax{}, false
	}
	qualifierStart := operationStart + 1
	if !staticRuleWordsAt(tokens, qualifierStart, "each", "combat", "if", "able") ||
		qualifierStart+4 != len(tokens)-1 {
		return requiredAttackRuleSyntax{}, false
	}
	constraintSpan := shared.SpanOf(tokens[constraintStart : qualifierStart+4])
	if operationStart != constraintStart {
		constraintSpan = tokens[constraintStart].Span
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: constraintSpan,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[operationStart].Span,
		},
		Qualifiers: []StaticRuleQualifier{
			{
				Kind: StaticRuleQualifierEachCombat,
				Span: shared.SpanOf(tokens[qualifierStart : qualifierStart+2]),
			},
			{
				Kind: StaticRuleQualifierIfAble,
				Span: shared.SpanOf(tokens[qualifierStart+2 : qualifierStart+4]),
			},
		},
	}, true
}

func validStaticRuleSyntax(rule StaticRuleSyntax) bool {
	switch rule.Subject.Kind {
	case StaticRuleSubjectSourceCreature:
		return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			len(rule.Qualifiers) == 0) ||
			(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
				rule.Operation.Kind == StaticRuleOperationAttack &&
				rule.Operation.Voice == StaticRuleVoiceActive &&
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble))
	case StaticRuleSubjectSourceSpell:
		return rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationCounter &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0
	default:
		return false
	}
}

func staticRuleQualifiersAre(qualifiers []StaticRuleQualifier, kinds ...StaticRuleQualifierKind) bool {
	if len(qualifiers) != len(kinds) {
		return false
	}
	for i := range qualifiers {
		if qualifiers[i].Kind != kinds[i] {
			return false
		}
	}
	return true
}

func staticRuleWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		token := tokens[start+i]
		if token.Kind != shared.Word || !strings.EqualFold(token.Text, word) {
			return false
		}
	}
	return true
}
