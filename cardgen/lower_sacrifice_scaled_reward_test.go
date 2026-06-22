package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerOptionalSacrificeScaledReward verifies the anchor card Disciple of
// Freyalise: "When this creature enters, you may sacrifice another creature. If
// you do, you gain X life and draw X cards, where X is that creature's power."
// The optional sacrifice publishes its success and the sacrificed permanent as
// a linked object; both rewards are gated on the sacrifice and read that
// permanent's power through the linked object.
func TestLowerOptionalSacrificeScaledReward(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Disciple of Freyalise",
		Layout:   "normal",
		TypeLine: "Creature",
		OracleText: "When this creature enters, you may sacrifice another creature. " +
			"If you do, you gain X life and draw X cards, where X is that creature's power.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d instructions, want 3 (sacrifice, gain, draw)", len(mode.Sequence))
	}

	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("sacrifice instruction is not Optional")
	}
	if mode.Sequence[0].PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("sacrifice PublishResult = %q, want %q", mode.Sequence[0].PublishResult, optionalIfYouDoResultKey)
	}
	if sacrifice.PublishLinked != sacrificedCreatureLinkKey {
		t.Fatalf("sacrifice PublishLinked = %q, want %q", sacrifice.PublishLinked, sacrificedCreatureLinkKey)
	}

	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	draw, ok := mode.Sequence[2].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("third primitive = %T, want game.Draw", mode.Sequence[2].Primitive)
	}

	for label, amount := range map[string]game.Quantity{"gain": gain.Amount, "draw": draw.Amount} {
		dynamic := amount.DynamicAmount()
		if !dynamic.Exists {
			t.Fatalf("%s amount = %+v, want dynamic", label, amount)
		}
		if dynamic.Val.Kind != game.DynamicAmountObjectPower {
			t.Fatalf("%s dynamic kind = %v, want DynamicAmountObjectPower", label, dynamic.Val.Kind)
		}
		if dynamic.Val.Object != game.LinkedObjectReference(string(sacrificedCreatureLinkKey)) {
			t.Fatalf("%s dynamic object = %+v, want linked sacrificed creature", label, dynamic.Val.Object)
		}
	}

	for i := 1; i <= 2; i++ {
		gate := mode.Sequence[i].ResultGate
		if !gate.Exists {
			t.Fatalf("reward instruction %d has no ResultGate", i)
		}
		if gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
			t.Fatalf("reward instruction %d gate = %+v, want if-you-do succeeded", i, gate.Val)
		}
	}
}
