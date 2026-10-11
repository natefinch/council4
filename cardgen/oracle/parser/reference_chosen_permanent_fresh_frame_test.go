package parser

import "testing"

func TestChosenPermanentFreshFrameKeepsOriginalIdentity(t *testing.T) {
	for _, prefix := range []string{
		"Manifest the top card of your library.",
		"Bolster 1.",
		"Choose a creature type.",
	} {
		t.Run(prefix, func(t *testing.T) {
			document, diagnostics := Parse(prefix+" Choose a land you control. "+chosenNameSearchText,
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			effects := conditionEffects(ability.Sentences)
			choice := effects[1]
			if choice.Kind != EffectChoosePermanent || choice.ClauseID == 0 {
				t.Fatal("missing unique original permanent declaration")
			}
			found := false
			for _, reference := range ability.SemanticReferences {
				if reference.Text == "the chosen land" {
					found = true
					if reference.ChoiceClauseID != choice.ClauseID || reference.ProducerClauseID != 0 ||
						reference.SubjectNoun != ObjectNounLand || reference.LibraryCardObservation {
						t.Fatal("fresh frame borrowed earlier subject provenance")
					}
				}
			}
			if !found {
				t.Fatal("missing chosen-name reference")
			}
		})
	}
}
