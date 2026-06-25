package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSacrificeSourceUnlessPay verifies that "sacrifice this creature
// unless you pay {U}." lowers to a Pay instruction followed by a sacrifice of
// the source permanent gated on the payment not being made.
func TestLowerSacrificeSourceUnlessPay(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Phantasmal Forces",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Illusion",
		OracleText: "At the beginning of your upkeep, sacrifice Phantasmal Forces unless you pay {U}.",
		Colors:     []string{"U"},
		Power:      new("4"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(content.Modes))
	}
	seq := content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}

	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first instruction primitive = %T, want game.Pay", seq[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists {
		t.Fatal("Pay instruction has no mana cost")
	}
	if seq[0].PublishResult == "" {
		t.Error("Pay instruction does not publish a result")
	}

	sacrifice, ok := seq[1].Primitive.(game.Sacrifice)
	if !ok {
		t.Fatalf("second instruction primitive = %T, want game.Sacrifice", seq[1].Primitive)
	}
	if sacrifice.Object != game.SourcePermanentReference() {
		t.Errorf("Sacrifice.Object = %v, want SourcePermanentReference", sacrifice.Object)
	}
	gate := seq[1].ResultGate
	if !gate.Exists {
		t.Fatal("Sacrifice instruction is not gated on the payment result")
	}
	if gate.Val.Key != seq[0].PublishResult {
		t.Errorf("gate key = %v, want %v (the Pay result)", gate.Val.Key, seq[0].PublishResult)
	}
	if gate.Val.Succeeded != game.TriFalse {
		t.Errorf("gate Succeeded = %v, want TriFalse", gate.Val.Succeeded)
	}
}
