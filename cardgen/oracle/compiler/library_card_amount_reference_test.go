package compiler

import "testing"

func TestLibraryCardAmountsKeepTheirExactParserOwnedReference(t *testing.T) {
	for _, text := range []string{
		"Reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.",
		"Reveal the top card of your library. If it's a creature card, you gain life equal to that card's toughness, lose life equal to its power, then put it into your hand.",
	} {
		t.Run(text, func(t *testing.T) {
			compilation, diagnostics := compileSource(text, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			checked := 0
			for _, effect := range content.Effects {
				if effect.Amount.DynamicKind != DynamicAmountSourcePower && effect.Amount.DynamicKind != DynamicAmountSourceToughness {
					continue
				}
				found := false
				for _, reference := range effect.References {
					if reference.NodeID != effect.Amount.ReferenceNodeID {
						continue
					}
					if found || reference.ProducerClauseID != content.Effects[0].ClauseID ||
						reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction != 0 {
						t.Fatalf("amount consumed a different or ambiguous producer: %#v", reference)
					}
					found = true
				}
				if !found {
					t.Fatalf("amount lost parser NodeID %d", effect.Amount.ReferenceNodeID)
				}
				checked++
			}
			if checked != 2 {
				t.Fatalf("checked %d characteristic operands, want 2", checked)
			}
		})
	}
}
