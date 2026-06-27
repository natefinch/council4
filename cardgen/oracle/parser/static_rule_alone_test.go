package parser

import (
	"testing"
)

// TestParseCombatAloneStaticRuleSentences covers the active-voice combat "alone"
// restrictions ("can't attack alone", "can't block alone", "can't attack or
// block alone"): each parses to a prohibition over the stated operation carrying
// exactly one StaticRuleQualifierAlone.
func TestParseCombatAloneStaticRuleSentences(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		operation StaticRuleOperationKind
	}{
		"cant attack alone": {
			source:    "This creature can't attack alone.",
			operation: StaticRuleOperationAttack,
		},
		"cant block alone": {
			source:    "This creature can't block alone.",
			operation: StaticRuleOperationBlock,
		},
		"cant attack or block alone": {
			source:    "This creature can't attack or block alone.",
			operation: StaticRuleOperationAttackOrBlock,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := parseStaticRuleSentence(t, test.source)
			if rule == nil {
				t.Fatalf("StaticRule = nil, want a static rule for %q", test.source)
			}
			if rule.Subject.Kind != StaticRuleSubjectSourceCreature {
				t.Fatalf("subject = %s, want %s", rule.Subject.Kind, StaticRuleSubjectSourceCreature)
			}
			if rule.Constraint.Kind != StaticRuleConstraintProhibition {
				t.Fatalf("constraint = %s, want %s", rule.Constraint.Kind, StaticRuleConstraintProhibition)
			}
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != StaticRuleVoiceActive {
				t.Fatalf("operation = %#v, want %s active", rule.Operation, test.operation)
			}
			if len(rule.Qualifiers) != 1 || rule.Qualifiers[0].Kind != StaticRuleQualifierAlone {
				t.Fatalf("qualifiers = %#v, want one %s", rule.Qualifiers, StaticRuleQualifierAlone)
			}
		})
	}
}

// TestParseCombatAloneStaticRuleSentencesFailClosed guards the "alone"
// restriction against near misses: a trailing clause after "alone", the passive
// "can't be blocked alone" wording, and a non-creature subject must not parse to
// a static rule carrying the alone qualifier.
func TestParseCombatAloneStaticRuleSentencesFailClosed(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"attack alone this turn": "This creature can't attack alone this turn.",
		"passive be blocked":     "This creature can't be blocked alone.",
		"permission attack":      "This creature can attack alone.",
		"unknown operation":      "This creature can't wobble alone.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					if sentence.StaticRule != nil &&
						staticRuleQualifiersAre(sentence.StaticRule.Qualifiers, StaticRuleQualifierAlone) {
						t.Fatalf("source %q parsed to an alone static rule: %#v", source, sentence.StaticRule)
					}
				}
			}
		})
	}
}
