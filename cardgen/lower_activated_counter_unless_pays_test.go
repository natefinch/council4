package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerActivatedCounterUnlessPays confirms an activated ability whose body
// is "Counter target spell unless its controller pays {N}." lowers without being
// rejected as an unsupported activation condition: the resolution-time tax stays
// in the body and lowers to a Pay instruction gating a CounterObject.
func TestLowerActivatedCounterUnlessPays(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Disruptive Student",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "{T}: Counter target spell unless its controller pays {1}.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ActivationCondition.Exists {
		t.Fatalf("activation condition = %#v, want none (tax belongs to the body)", ability.ActivationCondition)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2 (pay then counter)", len(mode.Sequence))
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.Pay", mode.Sequence[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists {
		t.Fatalf("pay mana cost missing: %#v", pay.Payment)
	}
	counter, ok := mode.Sequence[1].Primitive.(game.CounterObject)
	if !ok {
		t.Fatalf("instruction 1 = %T, want game.CounterObject", mode.Sequence[1].Primitive)
	}
	if counter.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("counter object = %#v, want target stack object 0", counter.Object)
	}
	if !mode.Sequence[1].ResultGate.Exists || mode.Sequence[1].ResultGate.Val.Succeeded != game.TriFalse {
		t.Fatalf("counter gate = %#v, want gated on unpaid tax", mode.Sequence[1].ResultGate)
	}
}

// TestLowerActivatedCounterUnlessPaysSacrificeCost confirms the same body lowers
// when the activation cost is a sacrifice rather than a tap, exercising the
// Wizard Replica / Spiketail shape.
func TestLowerActivatedCounterUnlessPaysSacrificeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Wizard Replica",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Wizard",
		OracleText: "Flying\n{U}, Sacrifice this creature: Counter target spell unless its controller pays {2}.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	if face.ActivatedAbilities[0].ActivationCondition.Exists {
		t.Fatalf("activation condition = %#v, want none", face.ActivatedAbilities[0].ActivationCondition)
	}
}
