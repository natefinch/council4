package parser

import "testing"

func TestParseNontokenCreatureSacrificeCost(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{U}, {T}, Sacrifice another nontoken creature: Create a 1/1 blue Zombie creature token.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	components := document.Abilities[0].CostSyntax.Components
	for _, component := range components {
		if component.Kind == CostComponentSacrifice {
			if !component.ObjectNonToken || component.ObjectNoun != ObjectNounCreature {
				t.Fatalf("sacrifice component = %#v", component)
			}
			return
		}
	}
	t.Fatal("sacrifice component not found")
}
