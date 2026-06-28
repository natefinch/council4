package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerIndependentOptionalTapOrUntapSequence verifies that a gate-free
// sequence of independent bare optionals — each carrying its own resolving "you
// may" with no "if you do" relationship between them — lowers so the controller
// decides each effect separately. Toils of Night and Day ("You may tap or untap
// target permanent, then you may tap or untap another target permanent.") is the
// canonical optional tap/untap member: two TapOrUntap instructions, each marked
// Optional, over two distinct permanent targets.
func TestLowerIndependentOptionalTapOrUntapSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Toils Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: "You may tap or untap target permanent, then you may tap or untap another target permanent.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %#v, want two permanent targets", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	for i, want := range []game.ObjectReference{
		game.TargetPermanentReference(0),
		game.TargetPermanentReference(1),
	} {
		instr := mode.Sequence[i]
		tapOrUntap, ok := instr.Primitive.(game.TapOrUntap)
		if !ok || tapOrUntap.Object != want {
			t.Fatalf("sequence[%d] = %#v, want TapOrUntap on target %d", i, instr.Primitive, i)
		}
		if !instr.Optional {
			t.Fatalf("sequence[%d] not marked Optional; each independent optional must be controller-decided", i)
		}
		if instr.PublishResult != "" || instr.ResultGate.Exists {
			t.Fatalf("sequence[%d] carries a gate envelope %#v; independent optionals are gate-free", i, instr)
		}
	}
}
