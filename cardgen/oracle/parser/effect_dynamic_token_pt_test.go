package parser

import "testing"

func TestParseDynamicBasePowerToughnessToken(t *testing.T) {
	t.Parallel()
	source := "Create a green Fungus Dinosaur creature token with base power and toughness each equal to the total power of those creatures."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.Exact ||
		effect.TokenPTKnown ||
		effect.TokenPTVariableX ||
		effect.TokenPTDynamic != EffectDynamicAmountTriggeringEventTotalPower ||
		!effect.Amount.Known ||
		effect.Amount.Value != 1 {
		t.Fatalf("effect = %#v", effect)
	}
	if effect.Selection.Kind != SelectionCreature ||
		len(effect.Selection.ColorsAny) != 1 ||
		effect.Selection.ColorsAny[0] != ColorGreen ||
		len(effect.Selection.SubtypesAny) != 2 ||
		effect.Selection.SubtypesAny[0] != "Fungus" ||
		effect.Selection.SubtypesAny[1] != "Dinosaur" {
		t.Fatalf("token selection = %#v", effect.Selection)
	}
}
