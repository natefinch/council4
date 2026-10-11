package parser

import "testing"

const chosenNameSearchText = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."

func TestChosenPermanentNameParserIdentity(t *testing.T) {
	for _, text := range []string{
		"Choose a land you control. " + chosenNameSearchText,
		"You gain 1 life. Choose a land you control. You gain 2 life. " + chosenNameSearchText,
		"Choose a land you control. " + chosenNameSearchText + " You gain 3 life. " + chosenNameSearchText,
		"You may pay {3}. If you do, choose a land you control. Then " + chosenNameSearchText,
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			effects := conditionEffects(ability.Sentences)
			choice := 0
			for _, effect := range effects {
				if effect.Kind == EffectChoosePermanent {
					choice = effect.ClauseID
				}
			}
			found := 0
			for _, reference := range ability.SemanticReferences {
				if reference.Text != "the chosen land" {
					continue
				}
				found++
				if choice == 0 || reference.ChoiceClauseID != choice ||
					reference.SubjectNoun != ObjectNounLand || reference.ProducerClauseID != 0 ||
					reference.LibraryCardObservation || reference.ReachedCardProducerClauseID != 0 ||
					reference.Order.End <= reference.Order.Start {
					t.Fatalf("incorrect chosen-name provenance: %+v", reference)
				}
				copies := 0
				for _, effect := range effects {
					for _, copy := range effect.References {
						if copy.NodeID == reference.NodeID {
							copies++
							if copy.ChoiceClauseID != choice || copy.SubjectNoun != reference.SubjectNoun {
								t.Fatalf("effect copy lost identity: %+v", copy)
							}
						}
					}
				}
				if copies != 1 {
					t.Fatalf("reference copies=%d", copies)
				}
			}
			if found == 0 {
				t.Fatal("missing chosen-name reference")
			}
		})
	}
}

func TestChosenPermanentNameParserDoesNotGuess(t *testing.T) {
	for _, prefix := range []string{
		"",
		"Choose a creature you control. ",
		"Choose two lands you control. ",
		"Choose a land you control. Choose a land you control. ",
		"Choose a land you control. Bolster 1. ",
		"Choose a land you control. Populate. ",
		"Choose a land you control. Choose a creature type. ",
	} {
		t.Run(prefix, func(t *testing.T) {
			document, diagnostics := Parse(prefix+chosenNameSearchText, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			for _, reference := range document.Abilities[0].SemanticReferences {
				if reference.Text == "the chosen land" && reference.ChoiceClauseID != 0 {
					t.Fatalf("ambiguous/missing choice received identity: %+v", reference)
				}
			}
		})
	}
}
