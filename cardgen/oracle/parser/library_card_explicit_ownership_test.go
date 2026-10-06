package parser

import "testing"

func TestParseLibraryCardExplicitObservationOwnership(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want []int
	}{
		{"look before reveal", "Look at the top card of your library. Reveal the top card of target opponent's library. Put the looked-at card into your hand.", []int{1}},
		{"reveal before look", "Reveal the top card of your library. Look at the top card of target opponent's library. Put the revealed card into your hand.", []int{1}},
		{"looked card and its owner", "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand.", []int{1, 1}},
		{"revealed card and its owner", "Reveal the top card of target player's library. Look at the top card of target opponent's library. Put the revealed card into that player's hand.", []int{1, 1}},
		{"generic card remains latest", "Look at the top card of your library. Reveal the top card of target opponent's library. Put that card into your hand.", []int{2}},
		{"linked reveal preserves card", "Look at the top card of your library. Reveal it. Put the revealed card into your hand.", []int{1, 1}},
		{"linked reveal of earlier look", "Look at the top card of your library. Reveal the top card of target opponent's library. Reveal the looked-at card. Put the revealed card into your hand.", []int{1, 1}},
		{"independent products keep owners", "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand. Put the revealed card into that player's graveyard.", []int{1, 1, 2, 2}},
		{"missing reveal", "Look at the top card of your library. Put the revealed card into your hand.", []int{0}},
		{"missing look", "Reveal the top card of your library. Put the looked-at card into your hand.", []int{0}},
		{"unsupported competing observation", "Look at the top card of your library. Reveal the top two cards of target opponent's library. Put the looked-at card into your hand.", []int{0}},
		{"competing target", "Look at the top card of your library. Exile target creature. Put the looked-at card into your hand.", []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 {
				t.Fatalf("parse diagnostics=%v abilities=%d", diagnostics, len(document.Abilities))
			}
			references := document.Abilities[0].SemanticReferences
			if len(references) != len(tc.want) {
				t.Fatalf("references=%#v, want producer clauses %v", references, tc.want)
			}
			for i, reference := range references {
				if reference.ProducerClauseID != tc.want[i] {
					t.Errorf("reference %d %q producer=%d, want %d", reference.NodeID, reference.Text, reference.ProducerClauseID, tc.want[i])
				}
			}
		})
	}
}
