package parser

import "testing"

func TestLibraryOwnerDestinationOwnsSameObservationClause(t *testing.T) {
	for _, text := range []string{
		"Look at the top card of target player's library. Draw a card. If it's a nonland card, put it into that player's graveyard.",
		"Reveal the top card of target opponent's library. If it's a land card, put it into that player's hand.",
		"Look at the top card of target player's library. Put it into that player's graveyard. Look at the top card of target opponent's library. Put it on the bottom of that player's library.",
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			checked := 0
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if !effect.LibraryOwnerDestination {
						continue
					}
					producer, owner := 0, 0
					for _, reference := range effect.References {
						if libraryCardReference(reference) {
							producer = reference.ProducerClauseID
						}
						if reference.Kind == ReferenceThatPlayer {
							owner = reference.ProducerClauseID
						}
					}
					if producer <= 0 || owner != producer {
						t.Fatalf("card producer=%d destination owner producer=%d", producer, owner)
					}
					checked++
				}
			}
			if checked == 0 {
				t.Fatal("destination did not retain exact library-owner relationship")
			}
		})
	}
}
