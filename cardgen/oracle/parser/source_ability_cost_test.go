package parser

import "testing"

func TestParseSourceAbilityCostReduction(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"{1}{G}, Discard this card: Draw a card. This ability costs {1} less to activate for each legendary creature you control.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	reduction := document.Abilities[0].SourceAbilityCostReduction
	if reduction == nil || reduction.Amount != 1 ||
		reduction.CountSelection.Kind != SelectionCreature ||
		len(reduction.CountSelection.Supertypes) != 1 ||
		reduction.CountSelection.Supertypes[0] != SupertypeLegendary ||
		reduction.CountSelection.Controller != SelectionControllerYou {
		t.Fatalf("reduction = %#v, want {1} per legendary creature you control", reduction)
	}
}

func TestSourceAbilityCostReductionPreservesUnsupportedMainSentence(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"{1}: Draw a card, then you become the monarch. This ability costs {1} less to activate for each legendary creature you control.",
		"{1}: Draw a card, then venture into the dungeon. This ability costs {1} less to activate for each legendary creature you control.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if ability.SourceAbilityCostReduction == nil {
				t.Fatalf("source %q did not produce a source-ability reduction", source)
			}
			effects := ability.Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact {
				t.Fatalf("effects = %#v, want one inexact effect", effects)
			}
			if coverage := DocumentCoverage(document); coverage.Complete {
				t.Fatalf("coverage = %#v, want unsupported main-sentence content", coverage)
			}
		})
	}
}
