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

// TestGenerateExileForPlayBatchSource covers the plural "you may exile one of
// them from your graveyard. If you do, you may cast it this turn." body over a
// "discard one or more cards" batch trigger (Conspiracy Theorist). The exile
// selects one card from the coalesced discard batch, so the ExileForPlay
// primitive sets SelectFromBatch instead of binding a single CardReferenceEvent
// card, and grants a cast permission for the rest of the turn.
func TestGenerateExileForPlayBatchSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Conspiracy Theorist",
		Layout:   "normal",
		TypeLine: "Creature — Human Shaman",
		ManaCost: "{1}{R}",
		OracleText: "Whenever you discard one or more nonland cards, you may exile one of them from your graveyard. " +
			"If you do, you may cast it this turn.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Event:  game.EventCardDiscarded,",
		"OneOrMore:     true,",
		"Primitive: game.ExileForPlay{",
		"SelectFromBatch: true,",
		"FromZone:        zone.Graveyard,",
		"Duration:        game.DurationThisTurn,",
		"Cast:            true,",
		"Optional: true,",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(spaceCollapsed(source), spaceCollapsed("Card: game.CardReference{Kind: game.CardReferenceEvent}")) {
		t.Fatalf("batch exile must not bind a single event card:\n%s", source)
	}
}

// TestGenerateExileForPlayCombinedPayDiscardSource covers Conspiracy Theorist's
// attack trigger "you may pay {1} and discard a card. If you do, draw a card.",
// the combined mana-and-discard reflexive payment. The recognized payment
// carries both the {1} mana cost and the discard additional cost, and its
// "controller-paid" result gates the draw.
func TestGenerateExileForPlayCombinedPayDiscardSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Conspiracy Theorist",
		Layout:     "normal",
		TypeLine:   "Creature — Human Shaman",
		ManaCost:   "{1}{R}",
		OracleText: "Whenever this creature attacks, you may pay {1} and discard a card. If you do, draw a card.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Pay{",
		"ManaCost: opt.Val(cost.Mana{",
		"Kind:   cost.AdditionalDiscard,",
		"PublishResult: game.ResultKey(\"controller-paid\"),",
		"Primitive: game.Draw{",
		"Key:       \"controller-paid\",",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
