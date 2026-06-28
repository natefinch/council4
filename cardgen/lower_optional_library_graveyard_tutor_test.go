package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// tutorSearchPrimitive lowers a single-face creature whose only ability is the
// enters-the-battlefield tutor trigger and returns the resolved optional Search
// primitive, failing the test if the body did not lower to exactly one optional
// Search instruction.
func tutorSearchPrimitive(t *testing.T, oracle string) game.Search {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Tutor Companion",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: oracle,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || len(faces[0].TriggeredAbilities) != 1 {
		t.Fatalf("faces = %#v", faces)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	if !seq[0].Optional {
		t.Error("Search instruction Optional = false, want true")
	}
	search, ok := seq[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %T, want game.Search", seq[0].Primitive)
	}
	return search
}

func assertTutorSearch(t *testing.T, search game.Search) {
	t.Helper()
	if !search.Spec.AlsoGraveyard {
		t.Error("Spec.AlsoGraveyard = false, want true")
	}
	if search.Spec.SourceZone != zone.Library {
		t.Errorf("Spec.SourceZone = %v, want Library", search.Spec.SourceZone)
	}
	if search.Spec.Destination != zone.Hand {
		t.Errorf("Spec.Destination = %v, want Hand", search.Spec.Destination)
	}
	if !search.Spec.Reveal {
		t.Error("Spec.Reveal = false, want true")
	}
	if search.Spec.Name != "Teferi, Timebender" {
		t.Errorf("Spec.Name = %q, want %q", search.Spec.Name, "Teferi, Timebender")
	}
	if search.Player != game.ControllerReference() {
		t.Errorf("Search.Player = %+v, want controller", search.Player)
	}
}

// TestLowerOptionalLibraryGraveyardTutorFiveEffect verifies the "and put ... If
// you search your library this way" wording (which compiles the library half to
// its own EffectSearch, five effects) lowers to one optional AlsoGraveyard
// Search reveal-to-hand.
func TestLowerOptionalLibraryGraveyardTutorFiveEffect(t *testing.T) {
	t.Parallel()
	search := tutorSearchPrimitive(t, "When this creature enters, you may search your library and/or graveyard for a card named Teferi, Timebender, reveal it, and put it into your hand. If you search your library this way, shuffle.")
	assertTutorSearch(t, search)
}

// TestLowerOptionalLibraryGraveyardTutorFourEffect verifies the "then put ... If
// you searched your library this way" wording (which folds the library half into
// the leading search, four effects) lowers to the same optional AlsoGraveyard
// Search reveal-to-hand.
func TestLowerOptionalLibraryGraveyardTutorFourEffect(t *testing.T) {
	t.Parallel()
	search := tutorSearchPrimitive(t, "When this creature enters, you may search your library and/or graveyard for a card named Teferi, Timebender, reveal it, then put it into your hand. If you searched your library this way, shuffle.")
	assertTutorSearch(t, search)
}

// TestLowerOptionalLibraryGraveyardTutorRejectsLibraryOnly verifies the
// dual-zone lowerer fails closed for a plain library-only named tutor: with no
// graveyard search the body falls through every optional lowerer and remains
// unsupported rather than producing a spurious AlsoGraveyard Search.
func TestLowerOptionalLibraryGraveyardTutorRejectsLibraryOnly(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Library Only Tutor",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "When this creature enters, you may search your library for a card named Teferi, Timebender, reveal it, and put it into your hand, then shuffle.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none, want unsupported optional effect")
	}
}
