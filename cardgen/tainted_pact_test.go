package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTaintedPactIterativeLibraryProcess proves Tainted Pact's four-effect
// "Exile the top card. You may put it into your hand unless it duplicates a name
// exiled this way. Repeat until you take one or exile a duplicate." body folds
// into a single IterativeLibraryProcess primitive with the duplicate-name stop,
// the optional take-to-hand knob, and no naming prelude, reveal, or pre-exile.
func TestLowerTaintedPactIterativeLibraryProcess(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Tainted Pact",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{B}",
		OracleText: "Exile the top card of your library. You may put that card into your hand " +
			"unless it has the same name as another card exiled this way. Repeat this process " +
			"until you put a card into your hand or you exile two cards with the same name, " +
			"whichever comes first.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	prim, ok := mode.Sequence[0].Primitive.(game.IterativeLibraryProcess)
	if !ok {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
	if prim.Stop != game.IterativeLibraryStopDuplicateName {
		t.Errorf("Stop = %v, want DuplicateName", prim.Stop)
	}
	if !prim.OptionalTake {
		t.Error("OptionalTake = false, want true")
	}
	if prim.ChooseName || prim.Reveal {
		t.Errorf("ChooseName=%v Reveal=%v, want both false", prim.ChooseName, prim.Reveal)
	}
	if prim.AllowAbsentName {
		t.Error("AllowAbsentName = true, want false (Tainted Pact names no card)")
	}
	if prim.PreExile.IsDynamic() || prim.PreExile.Value() != 0 {
		t.Errorf("PreExile = %#v, want fixed 0", prim.PreExile)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Errorf("Player = %#v, want controller", prim.Player)
	}
}
