package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// loweredActivatedSearch lowers a single-activated-ability card and returns the
// lone Search primitive that ability produces, failing the test on any diagnostic
// or unexpected shape.
func loweredActivatedSearch(t *testing.T, typeLine, oracleText string) game.Search {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Search Test",
		Layout:     "normal",
		TypeLine:   typeLine,
		OracleText: oracleText,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("lowerExecutableFaces(%q) diagnostics = %#v", oracleText, diagnostics)
	}
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 1 {
		t.Fatalf("lowerExecutableFaces(%q) faces = %#v", oracleText, faces)
	}
	modes := faces[0].ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("lowerExecutableFaces(%q) modes = %#v", oracleText, modes)
	}
	search, ok := modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("lowerExecutableFaces(%q) primitive = %#v, want game.Search", oracleText, modes[0].Sequence[0].Primitive)
	}
	return search
}

// TestLowerHeterogeneousSearchKrosanVerge confirms Krosan Verge's activated
// sacrifice ability lowers to one Search whose SearchSpec carries a per-slot
// filter for each distinct basic-land subtype, both onto the battlefield tapped.
func TestLowerHeterogeneousSearchKrosanVerge(t *testing.T) {
	t.Parallel()
	search := loweredActivatedSearch(t, "Land",
		"{2}, {T}, Sacrifice this land: Search your library for a Forest card and a Plains card, put them onto the battlefield tapped, then shuffle.")
	if got := search.Amount; got != game.Fixed(2) {
		t.Errorf("Amount = %#v, want Fixed(2)", got)
	}
	spec := search.Spec
	if spec.SourceZone != zone.Library {
		t.Errorf("SourceZone = %v, want Library", spec.SourceZone)
	}
	if spec.Destination != zone.Battlefield {
		t.Errorf("Destination = %v, want Battlefield", spec.Destination)
	}
	if !spec.EntersTapped {
		t.Error("EntersTapped = false, want true")
	}
	want := []types.Sub{types.Forest, types.Plains}
	if len(spec.SlotFilters) != len(want) {
		t.Fatalf("SlotFilters = %#v, want one filter per subtype %#v", spec.SlotFilters, want)
	}
	for i, sub := range want {
		filter := spec.SlotFilters[i]
		if len(filter.SubtypesAny) != 1 || filter.SubtypesAny[0] != sub {
			t.Errorf("SlotFilters[%d].SubtypesAny = %#v, want [%q]", i, filter.SubtypesAny, sub)
		}
	}
}

// TestLowerHeterogeneousSearchFailsClosed confirms wordings outside the modeled
// search → put-battlefield-tapped → shuffle envelope do not lower to a multi-slot
// search: they either lower through another path or fail closed with a
// diagnostic, but never silently drop a rider.
func TestLowerHeterogeneousSearchFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// A different destination (hand) is outside the battlefield-tapped shape.
		"{2}, {T}, Sacrifice this land: Search your library for a Forest card and a Plains card, put them into your hand, then shuffle.",
		// An untapped battlefield destination is outside the modeled shape.
		"{2}, {T}, Sacrifice this land: Search your library for a Forest card and a Plains card, put them onto the battlefield, then shuffle.",
		// A trailing rider breaks the exact three-effect envelope.
		"{2}, {T}, Sacrifice this land: Search your library for a Forest card and a Plains card, put them onto the battlefield tapped, then shuffle, then draw a card.",
	}
	for _, oracleText := range rejected {
		faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Search Test",
			Layout:     "normal",
			TypeLine:   "Land",
			OracleText: oracleText,
		})
		if len(diagnostics) != 0 {
			continue
		}
		if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 1 {
			t.Fatalf("lowerExecutableFaces(%q) faces = %#v", oracleText, faces)
		}
		seq := faces[0].ActivatedAbilities[0].Content.Modes
		if len(seq) == 1 && len(seq[0].Sequence) == 1 {
			if search, ok := seq[0].Sequence[0].Primitive.(game.Search); ok && len(search.Spec.SlotFilters) != 0 {
				t.Errorf("lowerExecutableFaces(%q) lowered to a multi-slot search, want fail closed", oracleText)
			}
		}
	}
}
