package compiler

import (
	"testing"
)

// TestCompileCombatAloneStaticRules covers the active-voice combat "alone"
// restrictions: each compiles to a single source-domain static declaration
// carrying the matching StaticRuleKind.
func TestCompileCombatAloneStaticRules(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		rule   StaticRuleKind
	}{
		"cant attack alone": {
			source: "This creature can't attack alone.",
			rule:   StaticRuleCantAttackAlone,
		},
		"cant block alone": {
			source: "This creature can't block alone.",
			rule:   StaticRuleCantBlockAlone,
		},
		"cant attack or block alone": {
			source: "This creature can't attack or block alone.",
			rule:   StaticRuleCantAttackOrBlockAlone,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			static := compilation.Abilities[0].Static
			if static == nil || len(static.Declarations) != 1 {
				t.Fatalf("static semantics = %#v, want one declaration", static)
			}
			declaration := static.Declarations[0]
			if declaration.Rule == nil || declaration.Rule.Kind != test.rule {
				t.Fatalf("rule = %#v, want %v", declaration.Rule, test.rule)
			}
			if declaration.Rule.Domain != staticRuleDomain(test.rule) {
				t.Fatalf("rule domain = %v, want %v", declaration.Rule.Domain, staticRuleDomain(test.rule))
			}
			if declaration.Group.Domain != StaticGroupSource {
				t.Fatalf("group domain = %v, want StaticGroupSource", declaration.Group.Domain)
			}
		})
	}
}

// TestCompileCombatAloneStaticRulesFailClosed confirms the "alone" recognizer
// stays bounded: a trailing duration clause must not compile to a static
// declaration carrying an alone rule kind.
func TestCompileCombatAloneStaticRulesFailClosed(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"attack alone this turn": "This creature can't attack alone this turn.",
		"permission attack":      "This creature can attack alone.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(source, pipelineContext{})
			for _, ability := range compilation.Abilities {
				if ability.Static == nil {
					continue
				}
				for _, declaration := range ability.Static.Declarations {
					if declaration.Rule == nil {
						continue
					}
					kind := declaration.Rule.Kind
					if kind == StaticRuleCantAttackAlone ||
						kind == StaticRuleCantBlockAlone ||
						kind == StaticRuleCantAttackOrBlockAlone {
						t.Fatalf("source %q compiled to an alone static rule: %#v", source, declaration.Rule)
					}
				}
			}
		})
	}
}
