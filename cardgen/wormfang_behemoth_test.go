package cardgen

import (
	"strings"
	"testing"
)

func wormfangBehemothCard() *ScryfallCard {
	power, toughness := "5", "5"
	return &ScryfallCard{
		Name:     "Wormfang Behemoth",
		Layout:   "normal",
		ManaCost: "{3}{U}{U}",
		TypeLine: "Creature — Nightmare Fish Beast",
		OracleText: "When this creature enters, exile all cards from your hand.\n" +
			"When this creature leaves the battlefield, return the exiled cards to their owner's hand.",
		Power:     &power,
		Toughness: &toughness,
	}
}

// TestGenerateExecutableCardSourceWormfangBehemoth covers the exiled-card
// back-reference: an enters-the-battlefield trigger exiles the controller's
// whole hand as a linked set, and a leaves-the-battlefield trigger returns
// exactly that set to its owners' hands. The two halves must share one linked
// key so the runtime binds them.
func TestGenerateExecutableCardSourceWormfangBehemoth(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(wormfangBehemothCard(), "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Event:  game.EventPermanentEnteredBattlefield",
		"FromZone:      zone.Hand,",
		"Destination:   zone.Exile,",
		`PublishLinked: game.LinkedKey("exiled-cards-to-hand"),`,
		"MatchFromZone: true,",
		`FromLinked:  game.LinkedKey("exiled-cards-to-hand"),`,
		"FromZone:    zone.Exile,",
		"Destination: zone.Hand,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
