package cardgen

import (
	"strings"
	"testing"
)

func raiyuuCard() *ScryfallCard {
	power, toughness := "4", "4"
	return &ScryfallCard{
		Name:      "Raiyuu, Storm's Edge",
		Layout:    "normal",
		ManaCost:  "{1}{R}{R}",
		TypeLine:  "Legendary Creature — Human Samurai",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "First strike\n" +
			"Whenever a Samurai or Warrior you control attacks alone, untap it. If it's the first combat phase of the turn, there is an additional combat phase after this phase.",
	}
}

// TestGenerateExecutableCardSourceRaiyuu asserts Raiyuu, Storm's Edge generates:
// its attacks-alone trigger untaps the triggering attacker, then queues an
// additional combat phase gated by the FirstCombatPhaseOfTurn condition. This
// exercises the trailing "there is an additional combat phase after this phase"
// word order and the "if it's the first combat phase of the turn" gate.
func TestGenerateExecutableCardSourceRaiyuu(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(raiyuuCard(), "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Untap{",
		"game.AddExtraPhases{",
		"Combat: true",
		"FirstCombatPhaseOfTurn: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Main:") {
		t.Fatalf("Raiyuu should queue combat only, no extra main phase:\n%s", source)
	}
}
