package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerCabarettiCourtyardSacrificeFetch verifies the New Capenna sacrifice
// fetch land cycle ("When this land enters, sacrifice it. When you do, search
// your library for a basic <T1>, <T2>, or <T3> card, put it onto the
// battlefield tapped, then shuffle and you gain 1 life.") lowers to the
// expected three-instruction triggered ability: sacrifice the entering land,
// fetch a matching basic to the battlefield tapped, then gain 1 life.
func TestLowerCabarettiCourtyardSacrificeFetch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Cabaretti Courtyard",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "When this land enters, sacrifice it. When you do, search your library for a basic Mountain, Forest, or Plains card, put it onto the battlefield tapped, then shuffle and you gain 1 life.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || len(faces[0].TriggeredAbilities) != 1 {
		t.Fatalf("faces = %#v, want one triggered ability", faces)
	}
	trigger := faces[0].TriggeredAbilities[0].Trigger
	if trigger.Type != game.TriggerWhen ||
		trigger.Pattern.Event != game.EventPermanentEnteredBattlefield ||
		trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger = %#v, want self enters-battlefield", trigger)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %#v, want sacrifice, search, gain life", seq)
	}

	sacrifice, ok := seq[0].Primitive.(game.Sacrifice)
	if !ok || sacrifice.Object != game.EventPermanentReference() {
		t.Fatalf("first primitive = %#v, want sacrifice of the entering land", seq[0].Primitive)
	}

	search, ok := seq[1].Primitive.(game.Search)
	if !ok || search.Player != game.ControllerReference() || search.Amount.Value() != 1 {
		t.Fatalf("second primitive = %#v, want controller search for one card", seq[1].Primitive)
	}
	wantSpec := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		EntersTapped: true,
		Filter: game.Selection{
			Supertypes: []types.Super{types.Basic},
			SubtypesAny: []types.Sub{
				types.Sub("Mountain"),
				types.Sub("Forest"),
				types.Sub("Plains"),
			},
		},
	}
	if !searchSpecEqual(search.Spec, wantSpec) {
		t.Errorf("search spec = %+v, want %+v", search.Spec, wantSpec)
	}

	gain, ok := seq[2].Primitive.(game.GainLife)
	if !ok || gain.Amount.Value() != 1 || gain.Player != game.ControllerReference() {
		t.Fatalf("third primitive = %#v, want controller gain 1 life", seq[2].Primitive)
	}
}

// TestLowerRiveteersOverlookSacrificeFetch verifies a second member of the
// cycle (different basic land subtypes) lowers identically apart from the fetch
// filter, confirming the recognizer is not pinned to one card's wording.
func TestLowerRiveteersOverlookSacrificeFetch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Riveteers Overlook",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "When this land enters, sacrifice it. When you do, search your library for a basic Mountain, Swamp, or Forest card, put it onto the battlefield tapped, then shuffle and you gain 1 life.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %#v, want sacrifice, search, gain life", seq)
	}
	search, ok := seq[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want game.Search", seq[1].Primitive)
	}
	wantSpec := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		EntersTapped: true,
		Filter: game.Selection{
			Supertypes: []types.Super{types.Basic},
			SubtypesAny: []types.Sub{
				types.Sub("Mountain"),
				types.Sub("Swamp"),
				types.Sub("Forest"),
			},
		},
	}
	if !searchSpecEqual(search.Spec, wantSpec) {
		t.Errorf("search spec = %+v, want %+v", search.Spec, wantSpec)
	}
}
