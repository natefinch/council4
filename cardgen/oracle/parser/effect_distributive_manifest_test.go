package parser

import "testing"

const unexplainedAbsenceText = "For each player, exile up to one target nonland permanent that player controls. For each permanent exiled this way, its controller cloaks the top card of their library."

func TestParseDistributiveExileCloak(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(unexplainedAbsenceText, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("document = %#v", document)
	}
	exile := document.Abilities[0].Sentences[0].Effects[0]
	if !exile.Exact || !exile.ExileForEachPlayer {
		t.Fatalf("exile = %#v", exile)
	}
	cloak := document.Abilities[0].Sentences[1].Effects[0]
	if !cloak.Exact || !cloak.CloakForEachExiledThisWay {
		t.Fatalf("cloak = %#v", cloak)
	}
}
