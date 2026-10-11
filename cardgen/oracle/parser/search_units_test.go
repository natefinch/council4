package parser

import "testing"

func TestSearchUnitOwnership(t *testing.T) {
	t.Parallel()
	const search = "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle."
	for _, tt := range []struct {
		text     string
		producer int
	}{
		{search + " You gain 2 life. Untap that land.", 1},
		{"Draw a card. " + search + " Scry 1. Untap that land.", 2},
		{search + " " + search + " You gain 2 life. Untap that land.", 4},
		{search + " Create a 1/1 green Insect creature token. Untap that land.", 0},
		{search + " This creature gets +1/+1 until end of turn. Untap it.", 0},
		{search + " Untap that creature.", 1},
		{"Search your library for two basic land cards, put them onto the battlefield tapped, then shuffle. Untap that land.", 1},
		{"Search your library for a basic land card, put it into your hand, then shuffle. Untap that land.", 1},
		{search + " Before you control four or more lands, untap that land.", 0},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			doc, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(doc.Abilities) != 1 {
				t.Fatal(diagnostics)
			}
			references := doc.Abilities[0].SemanticReferences
			actual := references[len(references)-1].ProducerClauseID
			if actual != tt.producer {
				t.Fatalf("producer=%d, want %d", actual, tt.producer)
			}
		})
	}
}

func TestSearchUnitTrailingClausePayload(t *testing.T) {
	t.Parallel()
	text := "If you have no cards in hand, search your library for a basic land card, reveal it, put it into your hand, then shuffle, then draw a card."
	doc, diagnostics := Parse(text, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(doc.Abilities) != 1 {
		t.Fatal(diagnostics)
	}
	ability := doc.Abilities[0]
	effects := ability.Sentences[0].Effects
	if len(effects) != 5 || !effects[0].Exact || effects[0].UnsupportedDetail != "" ||
		!effects[0].Amount.Known || effects[0].Amount.Value != 1 ||
		len(ability.ConditionSegments) != 1 || len(ability.ConditionSegments[0].Ownership.ClauseIDs) != 5 {
		t.Fatal("search group parameters or shared typed condition ownership were lost")
	}
}
