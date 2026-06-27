package compiler

import "testing"

// TestCompileAssignsCombatDamageByToughnessStaticRules covers the three subject
// scopes of the combat-damage replacement "<subject> assigns combat damage equal
// to its toughness rather than its power." Each compiles to a single
// combat-damage-domain static declaration carrying StaticRuleAssignsCombatDamageByToughness
// with the matching affected-group domain.
func TestCompileAssignsCombatDamageByToughnessStaticRules(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		group  StaticGroupDomain
	}{
		"this creature": {
			source: "This creature assigns combat damage equal to its toughness rather than its power.",
			group:  StaticGroupSource,
		},
		"each creature you control": {
			source: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			group:  StaticGroupSourceControllerPermanents,
		},
		"each creature": {
			source: "Each creature assigns combat damage equal to its toughness rather than its power.",
			group:  StaticGroupBattlefield,
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
			if declaration.Rule == nil || declaration.Rule.Kind != StaticRuleAssignsCombatDamageByToughness {
				t.Fatalf("rule = %#v, want StaticRuleAssignsCombatDamageByToughness", declaration.Rule)
			}
			if declaration.Rule.Domain != staticRuleDomain(StaticRuleAssignsCombatDamageByToughness) {
				t.Fatalf("rule domain = %v, want %v", declaration.Rule.Domain, staticRuleDomain(StaticRuleAssignsCombatDamageByToughness))
			}
			if declaration.Group.Domain != test.group {
				t.Fatalf("group domain = %v, want %v", declaration.Group.Domain, test.group)
			}
		})
	}
}

// TestCompileAssignsCombatDamageByToughnessFailsClosed confirms the recognizer
// stays bounded: the attached-subject (Aura/Equipment) forms and the conditional
// "with toughness greater than its power" form must not compile to a static
// declaration carrying the assigns-by-toughness rule kind.
func TestCompileAssignsCombatDamageByToughnessFailsClosed(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"enchanted creature":  "Enchanted creature assigns combat damage equal to its toughness rather than its power.",
		"equipped creature":   "Equipped creature assigns combat damage equal to its toughness rather than its power.",
		"conditional filter":  "Each creature you control with toughness greater than its power assigns combat damage equal to its toughness, rather than its power.",
		"opponent controlled": "Each creature you don't control assigns combat damage equal to its toughness rather than its power.",
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
					if declaration.Rule != nil && declaration.Rule.Kind == StaticRuleAssignsCombatDamageByToughness {
						t.Fatalf("source %q compiled to an assigns-by-toughness static rule: %#v", source, declaration.Rule)
					}
				}
			}
		})
	}
}
