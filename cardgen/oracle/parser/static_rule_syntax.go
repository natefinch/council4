package parser

import (
	"slices"
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
		operation, opNext, ok := parseProhibitedStaticRuleOperation(tokens, next)
		if !ok {
			return nil, false
		}
		rule.Operation = operation
		if qualifier, qualifierNext, ok := parseStaticBlockerRestrictionQualifier(tokens, opNext, len(tokens)-1); ok {
			rule.Qualifiers = append(rule.Qualifiers, qualifier)
			opNext = qualifierNext
		}
		if opNext != len(tokens)-1 {
			return nil, false
		}
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
	if block, ok := parseRequiredBlockRule(tokens, next); ok {
		rule.Constraint = block.Constraint
		rule.Operation = block.Operation
		rule.Qualifiers = block.Qualifiers
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	if constraint, operation, ok := parseStaticDoesntUntapRule(tokens, next); ok {
		rule.Constraint = constraint
		rule.Operation = operation
		if !validStaticRuleSyntax(*rule) {
			return nil, false
		}
		return rule, true
	}
	return nil, false
}

func parseStaticRuleSubject(tokens []shared.Token) (StaticRuleSubject, int, bool) {
	if !staticRuleWordsAt(tokens, 0, "this") {
		if staticRuleWordsAt(tokens, 0, "enchanted", "creature") ||
			staticRuleWordsAt(tokens, 0, "equipped", "creature") {
			return StaticRuleSubject{
				Kind: StaticRuleSubjectAttachedObject,
				Span: shared.SpanOf(tokens[:2]),
			}, 2, true
		}
		return StaticRuleSubject{}, 0, false
	}
	if len(tokens) < 2 {
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
	if staticRuleWordsAt(tokens, start, "attack", "or", "block") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationAttackOrBlock,
			Voice: StaticRuleVoiceActive,
			Span:  shared.SpanOf(tokens[start : start+3]),
		}, start + 3, true
	}
	if staticRuleWordsAt(tokens, start, "attack") {
		return StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[start].Span,
		}, start + 1, true
	}
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

func parseRequiredBlockRule(tokens []shared.Token, start int) (requiredAttackRuleSyntax, bool) {
	if !staticRuleWordsAt(tokens, start, "must", "be", "blocked", "if", "able") ||
		start+5 != len(tokens)-1 {
		return requiredAttackRuleSyntax{}, false
	}
	return requiredAttackRuleSyntax{
		Constraint: StaticRuleConstraint{
			Kind: StaticRuleConstraintRequirement,
			Span: tokens[start].Span,
		},
		Operation: StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[start+1 : start+3]),
		},
		Qualifiers: []StaticRuleQualifier{
			{
				Kind: StaticRuleQualifierIfAble,
				Span: shared.SpanOf(tokens[start+3 : start+5]),
			},
		},
	}, true
}

// parseStaticDoesntUntapRule recognizes "doesn't untap during your untap step"
// or "doesn't untap during its controller's untap step", modeling the frozen
// permanent as a prohibition on the untap operation. The trailing "untap step"
// phrasing is fixed and fully consumed.
func parseStaticDoesntUntapRule(tokens []shared.Token, start int) (StaticRuleConstraint, StaticRuleOperation, bool) {
	if !staticRuleWordsAt(tokens, start, "doesn't", "untap", "during") {
		return StaticRuleConstraint{}, StaticRuleOperation{}, false
	}
	cursor := start + 3
	switch {
	case staticRuleWordsAt(tokens, cursor, "your"):
		cursor++
	case staticRuleWordsAt(tokens, cursor, "its", "controller's"):
		cursor += 2
	default:
		return StaticRuleConstraint{}, StaticRuleOperation{}, false
	}
	if !staticRuleWordsAt(tokens, cursor, "untap", "step") || cursor+2 != len(tokens)-1 {
		return StaticRuleConstraint{}, StaticRuleOperation{}, false
	}
	constraint := StaticRuleConstraint{
		Kind: StaticRuleConstraintProhibition,
		Span: shared.SpanOf(tokens[start : start+1]),
	}
	operation := StaticRuleOperation{
		Kind:  StaticRuleOperationUntap,
		Voice: StaticRuleVoiceActive,
		Span:  shared.SpanOf(tokens[start+1 : cursor+2]),
	}
	return constraint, operation, true
}

// parseStaticBlockerRestrictionQualifier consumes the blocker-characteristic
// restriction "by creatures with flying", "by creatures with power N or less",
// or "by creatures with power N or greater" that bounds a passive "can't be
// blocked" prohibition to blockers matching that characteristic. The phrasing is
// fixed; any deviation fails closed. end is the exclusive bound (the period
// index) so the qualifier never consumes the terminating punctuation.
func parseStaticBlockerRestrictionQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if !staticRuleWordsAt(tokens, start, "by", "creatures", "with") {
		return StaticRuleQualifier{}, 0, false
	}
	cursor := start + 3
	if staticRuleWordsAt(tokens, cursor, "flying") && cursor < end {
		return StaticRuleQualifier{
			Kind: StaticRuleQualifierBlockerFlying,
			Span: shared.SpanOf(tokens[start : cursor+1]),
		}, cursor + 1, true
	}
	if !staticRuleWordsAt(tokens, cursor, "power") || cursor+3 > end {
		return StaticRuleQualifier{}, 0, false
	}
	amount, ok := staticUnsignedInteger(tokens[cursor+1])
	if !ok || !staticRuleWordsAt(tokens, cursor+2, "or") {
		return StaticRuleQualifier{}, 0, false
	}
	var kind StaticRuleQualifierKind
	switch {
	case staticRuleWordsAt(tokens, cursor+3, "less"):
		kind = StaticRuleQualifierBlockerPowerOrLess
	case staticRuleWordsAt(tokens, cursor+3, "greater"):
		kind = StaticRuleQualifierBlockerPowerOrGreater
	default:
		return StaticRuleQualifier{}, 0, false
	}
	return StaticRuleQualifier{
		Kind:   kind,
		Span:   shared.SpanOf(tokens[start : cursor+4]),
		Amount: amount,
	}, cursor + 4, true
}

func validStaticRuleSyntax(rule StaticRuleSyntax) bool {
	switch rule.Subject.Kind {
	case StaticRuleSubjectSourceCreature, StaticRuleSubjectAttachedObject:
		return validCreatureStaticRuleOperation(rule)
	case StaticRuleSubjectSourceSpell:
		return rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationCounter &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			len(rule.Qualifiers) == 0
	default:
		return false
	}
}

// validCreatureStaticRuleOperation reports whether a creature-scoped static rule
// (a creature source or the creature an Aura or Equipment is attached to) carries
// a recognized constraint, operation, voice, and qualifier set.
func validCreatureStaticRuleOperation(rule StaticRuleSyntax) bool {
	return (rule.Constraint.Kind == StaticRuleConstraintProhibition &&
		rule.Operation.Kind == StaticRuleOperationBlock &&
		rule.Operation.Voice == StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			(len(rule.Qualifiers) == 0 ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierByMoreThanOne) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerFlying) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerPowerOrLess) ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierBlockerPowerOrGreater))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationAttack &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			(len(rule.Qualifiers) == 0 ||
				staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierDefenderYou))) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationAttackOrBlock &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintProhibition &&
			rule.Operation.Kind == StaticRuleOperationUntap &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			len(rule.Qualifiers) == 0) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationAttack &&
			rule.Operation.Voice == StaticRuleVoiceActive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble)) ||
		(rule.Constraint.Kind == StaticRuleConstraintRequirement &&
			rule.Operation.Kind == StaticRuleOperationBlock &&
			rule.Operation.Voice == StaticRuleVoicePassive &&
			staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierIfAble))
}

func staticRuleQualifiersAre(qualifiers []StaticRuleQualifier, kinds ...StaticRuleQualifierKind) bool {
	actual := make([]StaticRuleQualifierKind, len(qualifiers))
	for i := range qualifiers {
		actual[i] = qualifiers[i].Kind
	}
	return slices.Equal(actual, kinds)
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
