package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCombatAloneStaticRuleDeclarations proves the combat "alone"
// restrictions lower from typed semantics alone (no Oracle wording inspected) to
// the canonical runtime static bodies. Each lowered ability must name and
// byte-match the matching game.Cant*AloneStaticBody so the render path emits the
// shared constant.
func TestLowerCombatAloneStaticRuleDeclarations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		rule    compiler.StaticRuleKind
		domain  compiler.StaticRuleDomain
		varName string
		body    game.StaticAbility
	}{
		"cant attack alone": {
			rule:    compiler.StaticRuleCantAttackAlone,
			domain:  compiler.StaticRuleDomainAttack,
			varName: "game.CantAttackAloneStaticBody",
			body:    game.CantAttackAloneStaticBody,
		},
		"cant block alone": {
			rule:    compiler.StaticRuleCantBlockAlone,
			domain:  compiler.StaticRuleDomainBlock,
			varName: "game.CantBlockAloneStaticBody",
			body:    game.CantBlockAloneStaticBody,
		},
		"cant attack or block alone": {
			rule:    compiler.StaticRuleCantAttackOrBlockAlone,
			domain:  compiler.StaticRuleDomainAttackBlock,
			varName: "game.CantAttackOrBlockAloneStaticBody",
			body:    game.CantAttackOrBlockAloneStaticBody,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lowered, handled, diagnostic := lowerStaticDeclarations(compiler.CompiledAbility{
				Kind: compiler.AbilityStatic,
				Text: test.body.Text,
				Static: &compiler.CompiledStaticSemantics{
					Declarations: []compiler.StaticDeclaration{{
						Kind:  compiler.StaticDeclarationRule,
						Group: compiler.StaticGroupReference{Domain: compiler.StaticGroupSource},
						Rule: &compiler.StaticRuleDeclaration{
							Domain: test.domain,
							Kind:   test.rule,
							Zone:   compiler.StaticZoneBattlefield,
						},
					}},
				},
			}, &parser.Ability{})
			if !handled || diagnostic != nil || len(lowered.staticAbilities) != 1 {
				t.Fatalf("handled = %v, diagnostic = %#v, lowered = %#v", handled, diagnostic, lowered)
			}
			ability := lowered.staticAbilities[0]
			if ability.VarName != test.varName {
				t.Fatalf("var name = %q, want %q", ability.VarName, test.varName)
			}
			if !reflect.DeepEqual(ability.Body, test.body) {
				t.Fatalf("body = %#v, want %#v", ability.Body, test.body)
			}
		})
	}
}
