package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerKickedAdditiveRider verifies the "If this spell was kicked, <effect>"
// per-effect gate lowers to a SpellWasKicked condition on the kicked rider only,
// leaving the base effect ungated (Blink of an Eye). The kicker keyword supplies
// the optional additional cost; the rider draw resolves only when the resolving
// spell was kicked.
func TestLowerKickedAdditiveRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Blink of an Eye",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Instant",
		OracleText: "Kicker {1}{U}\nReturn target nonland permanent to its owner's hand. If this spell was kicked, draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}

	if _, ok := mode.Sequence[0].Primitive.(game.Bounce); !ok {
		t.Fatalf("instruction 0 = %T, want game.Bounce", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("base bounce must not be gated on kicker")
	}

	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("instruction 1 = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].Condition.Exists {
		t.Fatal("kicked draw rider is not gated")
	}
	if !mode.Sequence[1].Condition.Val.Condition.Val.SpellWasKicked {
		t.Fatalf("kicked draw gate = %#v, want SpellWasKicked", mode.Sequence[1].Condition.Val.Condition.Val)
	}
}
