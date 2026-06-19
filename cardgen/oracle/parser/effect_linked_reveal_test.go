package parser

import "testing"

func TestShuffleRevealPermanentSequenceExactness(t *testing.T) {
	t.Parallel()
	const canonical = "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."
	tests := []struct {
		name       string
		oracleText string
		want       bool
	}{
		{name: "canonical", oracleText: canonical, want: true},
		{name: "target card", oracleText: "The owner of target card shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."},
		{name: "target spell", oracleText: "The owner of target spell shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."},
		{name: "controller actor", oracleText: "The controller of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."},
		{name: "bottom card", oracleText: "The owner of target permanent shuffles it into their library, then reveals the bottom card of their library. If it's a permanent card, they put it onto the battlefield."},
		{name: "multiple cards", oracleText: "The owner of target permanent shuffles it into their library, then reveals the top two cards of their library. If they're permanent cards, they put them onto the battlefield."},
		{name: "unconditional put", oracleText: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. They put it onto the battlefield."},
		{name: "nonpermanent filter", oracleText: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a creature card, they put it onto the battlefield."},
		{name: "optional put", oracleText: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they may put it onto the battlefield."},
		{name: "extra words", oracleText: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield tapped."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.oracleText, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics = %#v", diagnostics)
			}
			got := parsedShuffleRevealPermanentSequence(document)
			if got != test.want {
				t.Fatalf("recognized = %v, want %v", got, test.want)
			}
		})
	}
}

func parsedShuffleRevealPermanentSequence(document Document) bool {
	if len(document.Abilities) != 1 ||
		len(document.Abilities[0].Sentences) != 2 ||
		len(document.Abilities[0].Sentences[0].Effects) != 2 ||
		len(document.Abilities[0].Sentences[1].Effects) != 1 {
		return false
	}
	shuffle := document.Abilities[0].Sentences[0].Effects[0]
	reveal := document.Abilities[0].Sentences[0].Effects[1]
	put := document.Abilities[0].Sentences[1].Effects[0]
	return shuffle.Exact &&
		shuffle.Player == EffectPlayerTargetOwner &&
		reveal.Exact &&
		reveal.Player == EffectPlayerTargetOwner &&
		reveal.CardSource == EffectCardSourceTopOfPlayerLibrary &&
		put.Exact &&
		put.Player == EffectPlayerTargetOwner &&
		put.CardSource == EffectCardSourcePriorInstructionResult &&
		put.RequirePermanentCard
}
