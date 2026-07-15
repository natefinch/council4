package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileStaticGroupAnthemAlsoAdverb proves a stacked threshold anthem's
// follow-on clause ("Creatures you control also get +1/+0 and have trample as
// long as you control six or more creatures.", Jetmir, Nexus of Revels) compiles
// into the same composed [power/toughness, keyword] continuous declarations over
// the controller's creatures, gated by the same control-count condition, as the
// identical clause printed without the "also" adverb. The adverb only emphasizes
// that the bonus accumulates, so it must not change the compiled declarations.
func TestCompileStaticGroupAnthemAlsoAdverb(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"without also": "Creatures you control get +1/+0 and have trample as long as you control six or more creatures.",
		"with also":    "Creatures you control also get +1/+0 and have trample as long as you control six or more creatures.",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil || len(ability.Static.Declarations) != 2 {
				t.Fatalf("static semantics = %#v, want two declarations", ability.Static)
			}
			pt := ability.Static.Declarations[0]
			if pt.Kind != StaticDeclarationContinuous ||
				pt.Continuous == nil ||
				pt.Continuous.Layer != StaticLayerPowerToughnessModify ||
				pt.Continuous.Operation != StaticContinuousModifyPowerToughness ||
				pt.Continuous.PowerDelta.Value != 1 ||
				pt.Continuous.ToughnessDelta.Value != 0 {
				t.Fatalf("power/toughness declaration = %#v", pt)
			}
			keyword := ability.Static.Declarations[1]
			if keyword.Kind != StaticDeclarationContinuous ||
				keyword.Continuous == nil ||
				keyword.Continuous.Operation != StaticContinuousGrantKeywords {
				t.Fatalf("keyword declaration = %#v", keyword)
			}
			for i, declaration := range ability.Static.Declarations {
				if declaration.Group.Domain != StaticGroupSourceControllerPermanents ||
					!slices.Equal(declaration.Group.Selection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("declaration[%d] group = %#v, want controller's creatures", i, declaration.Group)
				}
				if declaration.Condition == nil {
					t.Fatalf("declaration[%d] missing control-count condition", i)
				}
			}
		})
	}
}
