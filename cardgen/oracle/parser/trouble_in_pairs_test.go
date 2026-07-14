package parser

import "testing"

func TestParseTroubleInPairsMarkers(t *testing.T) {
	t.Parallel()
	const source = "If an opponent would begin an extra turn, that player skips that turn instead.\nWhenever an opponent attacks you with two or more creatures, draws their second card each turn, or casts their second spell each turn, you draw a card."
	document, diagnostics := Parse(source, Context{CardName: "Trouble in Pairs"})
	if len(diagnostics) != 0 || len(document.Abilities) != 2 {
		t.Fatalf("document = %#v, diagnostics = %#v", document, diagnostics)
	}
	if document.Abilities[0].SkipExtraTurnsScope != TriggerPlayerSelectorOpponent {
		t.Fatalf("skip scope = %v", document.Abilities[0].SkipExtraTurnsScope)
	}
	if !document.Abilities[1].OpponentSecondActionTriplet {
		t.Fatal("second-action triplet not recognized")
	}
}
