package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExileForPlaySource covers the discard-trigger reflexive
// "you may exile that card from your graveyard. If you do, you may play that
// card this turn." body (Containment Construct), which exiles the just-discarded
// card and lets its controller play it for the rest of the turn. The combined
// ExileForPlay primitive performs the exile and the play-permission grant
// atomically, binding the play permission to the discarded card by identity.
func TestGenerateExileForPlaySource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Containment Construct",
		Layout:   "normal",
		TypeLine: "Artifact Creature — Construct",
		ManaCost: "{2}",
		OracleText: "Whenever you discard a card, you may exile that card from your graveyard. " +
			"If you do, you may play that card this turn.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Event:  game.EventCardDiscarded,",
		"Primitive: game.ExileForPlay{",
		"Card:     game.CardReference{Kind: game.CardReferenceEvent},",
		"FromZone: zone.Graveyard,",
		"Duration: game.DurationThisTurn,",
		"Optional: true,",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(spaceCollapsed(source), spaceCollapsed("Cast: true,")) {
		t.Fatalf("play permission must not set Cast:\n%s", source)
	}
}
