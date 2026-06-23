package cardgen

import (
	"strings"
	"testing"
)

func scholarOfNewHorizonsCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Scholar of New Horizons",
		Layout:   "normal",
		ManaCost: "{1}{W}",
		TypeLine: "Creature — Human Scout",
		OracleText: "This creature enters with a +1/+1 counter on it.\n" +
			"{T}, Remove a counter from a permanent you control: Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.",
	}
}

// TestGenerateExecutableCardSourceScholarOfNewHorizons asserts the "search your
// library for a Plains card and reveal it; if an opponent controls more lands
// than you, you may put it onto the battlefield tapped, otherwise into your
// hand; then shuffle" activated ability lowers to a reveal-only search that
// publishes the found card and a ConditionalDestinationPlace that routes it
// under the lowered control-comparison gate, followed by the closing shuffle.
func TestGenerateExecutableCardSourceScholarOfNewHorizons(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(scholarOfNewHorizonsCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Search{",
		"RevealOnly: true",
		"Reveal:     true",
		"Filter:     game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Plains\")}}",
		"PublishLinked: game.LinkedKey(\"conditional-destination-card\")",
		"game.ConditionalDestinationPlace{",
		"Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: \"conditional-destination-card\"}",
		"FromZone: zone.Library",
		"ControlComparison: opt.Val(game.ControlCountComparison{",
		"EntryTapped: true",
		"Else:        zone.Hand",
		"game.ShuffleLibrary{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
