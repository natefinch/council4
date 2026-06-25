package cardgen

import (
	"strings"
	"testing"
)

func hedgeShredderCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Hedge Shredder",
		Layout:    "normal",
		ManaCost:  "{2}{G}{G}",
		TypeLine:  "Artifact — Vehicle",
		Power:     new("5"),
		Toughness: new("5"),
		OracleText: "Whenever this Vehicle attacks, you may mill two cards.\n" +
			"Whenever one or more land cards are put into your graveyard from your library, put them onto the battlefield tapped.\n" +
			"Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)",
	}
}

// TestGenerateExecutableCardSourceHedgeShredder asserts the Vehicle target card
// lowers all three abilities: the Crew 1 activated ability, the attack-triggered
// optional mill, and the "Whenever one or more land cards are put into your
// graveyard from your library, put them onto the battlefield tapped" batch
// reanimation. The plural "them" must lower to a MassReturnFromGraveyard
// restricted to the coalesced trigger batch (FromTriggerBatch) entering tapped,
// not a whole-graveyard recursion.
func TestGenerateExecutableCardSourceHedgeShredder(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(hedgeShredderCard(), "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.CrewActivatedAbility(1)",
		"game.Mill{",
		"game.MassReturnFromGraveyard{",
		"Destination:      zone.Battlefield,",
		"EntryTapped:      true,",
		"FromTriggerBatch: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
