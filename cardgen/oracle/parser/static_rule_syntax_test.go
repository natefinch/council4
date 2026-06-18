package parser

import "testing"

// parseStaticRuleSentence parses a single static-ability sentence and returns the
// whole-sentence StaticRule the parser emitted. It fails the test when the source
// did not produce exactly one ability with one static-rule sentence.
func parseStaticRuleSentence(t *testing.T, source string) *StaticRuleSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v, want one ability with one sentence", document.Abilities)
	}
	return document.Abilities[0].Sentences[0].StaticRule
}

func TestParseAttachedAndUntapStaticRuleSentences(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source     string
		subject    StaticRuleSubjectKind
		constraint StaticRuleConstraintKind
		operation  StaticRuleOperationKind
		voice      StaticRuleVoice
	}{
		"enchanted creature can't attack or block": {
			source:     "Enchanted creature can't attack or block.",
			subject:    StaticRuleSubjectAttachedObject,
			constraint: StaticRuleConstraintProhibition,
			operation:  StaticRuleOperationAttackOrBlock,
			voice:      StaticRuleVoiceActive,
		},
		"equipped creature can't be blocked": {
			source:     "Equipped creature can't be blocked.",
			subject:    StaticRuleSubjectAttachedObject,
			constraint: StaticRuleConstraintProhibition,
			operation:  StaticRuleOperationBlock,
			voice:      StaticRuleVoicePassive,
		},
		"this creature can't attack or block": {
			source:     "This creature can't attack or block.",
			subject:    StaticRuleSubjectSourceCreature,
			constraint: StaticRuleConstraintProhibition,
			operation:  StaticRuleOperationAttackOrBlock,
			voice:      StaticRuleVoiceActive,
		},
		"this creature doesn't untap your step": {
			source:     "This creature doesn't untap during your untap step.",
			subject:    StaticRuleSubjectSourceCreature,
			constraint: StaticRuleConstraintProhibition,
			operation:  StaticRuleOperationUntap,
			voice:      StaticRuleVoiceActive,
		},
		"enchanted creature doesn't untap controller step": {
			source:     "Enchanted creature doesn't untap during its controller's untap step.",
			subject:    StaticRuleSubjectAttachedObject,
			constraint: StaticRuleConstraintProhibition,
			operation:  StaticRuleOperationUntap,
			voice:      StaticRuleVoiceActive,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := parseStaticRuleSentence(t, test.source)
			if rule == nil {
				t.Fatalf("StaticRule = nil, want a static rule for %q", test.source)
			}
			if rule.Subject.Kind != test.subject {
				t.Fatalf("subject = %s, want %s", rule.Subject.Kind, test.subject)
			}
			if rule.Constraint.Kind != test.constraint {
				t.Fatalf("constraint = %s, want %s", rule.Constraint.Kind, test.constraint)
			}
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != test.voice {
				t.Fatalf("operation = %#v, want %s voice %s", rule.Operation, test.operation, test.voice)
			}
		})
	}
}

func TestParseAttachedAndUntapStaticRuleSentencesFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"attack and block":         "Enchanted creature can't attack and block.",
		"attack or block extra":    "Enchanted creature can't attack or block this turn.",
		"untap missing during":     "Enchanted creature doesn't untap.",
		"untap wrong step":         "Enchanted creature doesn't untap during your upkeep step.",
		"untap non creature":       "Enchanted permanent doesn't untap during your untap step.",
		"attack or block artifact": "Enchanted artifact can't attack or block.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", document.Abilities)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				if sentence.StaticRule != nil {
					t.Fatalf("sentence produced StaticRule %#v, want none (fail closed)", sentence.StaticRule)
				}
			}
		})
	}
}
