package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerAnyNumberGodSearchToBattlefield proves an "any number of" library
// search lowers to a SearchSpec marked AnyNumber, with an inert Fixed(0) amount
// that runtime replaces with the candidate count. It mirrors The World Tree's
// "Search your library for any number of God cards, put them onto the
// battlefield, then shuffle."
func TestLowerAnyNumberGodSearchToBattlefield(t *testing.T) {
	t.Parallel()
	search := loweredSearch(t, "Sorcery", "Search your library for any number of God cards, put them onto the battlefield, then shuffle.")
	if !search.Spec.AnyNumber {
		t.Error("search spec AnyNumber = false, want true")
	}
	if search.Spec.SourceZone != zone.Library || search.Spec.Destination != zone.Battlefield {
		t.Errorf("search zones = %v -> %v, want library -> battlefield", search.Spec.SourceZone, search.Spec.Destination)
	}
	if !slices.Contains(search.Spec.Filter.SubtypesAny, types.God) {
		t.Errorf("search filter subtypes = %+v, want to include God", search.Spec.Filter.SubtypesAny)
	}
	if got := search.Amount.Value(); got != 0 {
		t.Errorf("search amount = %d, want 0 placeholder (runtime uses the candidate count)", got)
	}
}

// TestLowerLeadingEffectThenAnyNumberSearch proves that once an "any number of"
// search lowers as a generic instruction, a leading effect before it is
// preserved rather than dropped: "Draw a card. Search your library for any
// number of basic land cards, put them onto the battlefield, then shuffle."
// lowers to a faithful two-instruction sequence (draw, then the any-number
// search). This shape formerly failed closed only because any-number searches
// were unsupported, not because the leading effect is unmodelable.
func TestLowerLeadingEffectThenAnyNumberSearch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Leading Any Number Search",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card. Search your library for any number of basic land cards, put them onto the battlefield, then shuffle.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || !faces[0].SpellAbility.Exists {
		t.Fatalf("faces = %#v, want one spell ability", faces)
	}
	modes := faces[0].SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then search", modes)
	}
	if _, ok := modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first instruction = %T, want game.Draw", modes[0].Sequence[0].Primitive)
	}
	search, ok := modes[0].Sequence[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second instruction = %T, want game.Search", modes[0].Sequence[1].Primitive)
	}
	if !search.Spec.AnyNumber || search.Spec.Destination != zone.Battlefield {
		t.Errorf("search spec = %+v, want AnyNumber onto the battlefield", search.Spec)
	}
}
