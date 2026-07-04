package cardgen

import (
	"strings"
	"testing"
)

// TestLowerBackReferencedReanimationEscalation verifies the "return that
// permanent card to the battlefield instead" back-reference reanimation escalates
// a base return-to-hand: the base clause returns the shared target to hand, and
// the gated escalation reanimates that same target to the battlefield. Court of
// Ardenvale gates it on the monarch designation; Shepherd of the Clouds gates it
// on controlling a Mount.
func TestLowerBackReferencedReanimationEscalation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		wantGate   string
		typeLine   string
		manaCost   string
		powerTough bool
	}{
		{
			name:     "monarch escalation returns shared target to battlefield",
			oracle:   "When this enchantment enters, you become the monarch.\nAt the beginning of your upkeep, return target permanent card with mana value 3 or less from your graveyard to your hand. If you're the monarch, return that permanent card to the battlefield instead.",
			wantGate: "ControllerIsMonarch: true",
			typeLine: "Enchantment",
			manaCost: "{3}{W}",
		},
		{
			name:       "controls-a-mount escalation returns shared target to battlefield",
			oracle:     "Flying, vigilance\nWhen this creature enters, return target permanent card with mana value 3 or less from your graveyard to your hand. Return that card to the battlefield instead if you control a Mount.",
			wantGate:   "ControlsMatching",
			typeLine:   "Creature — Bird",
			manaCost:   "{3}{W}",
			powerTough: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Reanimator",
				Layout:     "normal",
				ManaCost:   tc.manaCost,
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
				Colors:     []string{"W"},
			}
			if tc.powerTough {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want none", diagnostics)
			}
			for _, want := range []string{
				"Destination: zone.Hand",
				"game.PutOnBattlefield{",
				"game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})",
				tc.wantGate,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
