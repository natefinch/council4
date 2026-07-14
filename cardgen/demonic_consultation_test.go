package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDemonicConsultationIterativeLibraryProcess proves Demonic
// Consultation's "Choose a card name." prelude plus the five-effect "Exile the
// top six, reveal until the named card, put it into your hand, exile the rest"
// body folds into a single IterativeLibraryProcess primitive with the
// chosen-name stop, the naming and reveal knobs, and a fixed six-card pre-exile.
func TestLowerDemonicConsultationIterativeLibraryProcess(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Demonic Consultation",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{B}",
		OracleText: "Choose a card name. Exile the top six cards of your library, then reveal " +
			"cards from the top of your library until you reveal a card with the chosen name. " +
			"Put that card into your hand and exile all other cards revealed this way.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	prim, ok := mode.Sequence[0].Primitive.(game.IterativeLibraryProcess)
	if !ok {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
	if prim.Stop != game.IterativeLibraryStopChosenName {
		t.Errorf("Stop = %v, want ChosenName", prim.Stop)
	}
	if !prim.ChooseName {
		t.Error("ChooseName = false, want true")
	}
	if !prim.Reveal {
		t.Error("Reveal = false, want true")
	}
	if prim.OptionalTake {
		t.Error("OptionalTake = true, want false")
	}
	if !prim.AllowAbsentName {
		t.Error("AllowAbsentName = false, want true")
	}
	if prim.PreExile.IsDynamic() || prim.PreExile.Value() != 6 {
		t.Errorf("PreExile = %#v, want fixed 6", prim.PreExile)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Errorf("Player = %#v, want controller", prim.Player)
	}
}
