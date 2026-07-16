package compiler

import "testing"

func TestCompileSakashimaCopyAndLegendRule(t *testing.T) {
	t.Parallel()
	source := "You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.\n" +
		"The \"legend rule\" doesn't apply to permanents you control.\n" +
		"Partner (You can have two commanders if both have partner.)"
	compilation, diagnostics := compileSource(source, pipelineContext{
		CardName:  "Sakashima of a Thousand Faces",
		Legendary: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var foundCopy, foundLegend bool
	for _, ability := range compilation.Abilities {
		for _, effect := range ability.Content.Effects {
			if effect.EntersAsCopy {
				foundCopy = effect.EntersAsCopyOptional &&
					effect.EntersAsCopyRetainName &&
					effect.EntersAsCopyAddOtherAbilities
			}
		}
		if ability.Static == nil {
			continue
		}
		for _, declaration := range ability.Static.Declarations {
			if declaration.Player != nil &&
				declaration.Player.Kind == StaticPlayerRuleLegendRuleDoesNotApply {
				foundLegend = true
			}
		}
	}
	if !foundCopy {
		t.Fatal("compiled copy effect did not preserve the typed copy exceptions")
	}
	if !foundLegend {
		t.Fatal("compiled static declarations did not include the legend-rule exemption")
	}
}
