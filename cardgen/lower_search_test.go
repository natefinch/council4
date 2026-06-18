package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// loweredSearch lowers a single-spell-ability card and returns the lone Search
// primitive its spell ability produces, failing the test on any diagnostic or
// unexpected shape.
func loweredSearch(t *testing.T, typeLine, oracleText string) game.Search {
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
	if len(faces) != 1 || !faces[0].SpellAbility.Exists {
		t.Fatalf("lowerExecutableFaces(%q) faces = %#v", oracleText, faces)
	}
	modes := faces[0].SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("lowerExecutableFaces(%q) modes = %#v", oracleText, modes)
	}
	search, ok := modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("lowerExecutableFaces(%q) primitive = %#v, want game.Search", oracleText, modes[0].Sequence[0].Primitive)
	}
	return search
}

func TestLowerSearchSpellSpecs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		amount     int
		spec       game.SearchSpec
	}{
		{
			name:       "basic land ramp to battlefield tapped, up to two",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				CardType:     opt.Val(types.Land),
				Supertype:    opt.Val(types.Basic),
				EntersTapped: true,
			},
		},
		{
			name:       "dual-land subtype union to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a Forest or Island card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				SubtypesAny: []types.Sub{types.Forest, types.Island},
			},
		},
		{
			name:       "creature tutor to hand with reveal",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
				Reveal:      true,
			},
		},
		{
			name:       "basic triome to battlefield tapped",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a basic Forest, Plains, or Island card, put it onto the battlefield tapped, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				Supertype:    opt.Val(types.Basic),
				SubtypesAny:  []types.Sub{types.Forest, types.Plains, types.Island},
				EntersTapped: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			search := loweredSearch(t, test.typeLine, test.oracleText)
			if got := search.Amount.Value(); got != test.amount {
				t.Errorf("amount = %d, want %d", got, test.amount)
			}
			if got := search.Spec; !searchSpecEqual(got, test.spec) {
				t.Errorf("spec = %+v, want %+v", got, test.spec)
			}
		})
	}
}

func searchSpecEqual(a, b game.SearchSpec) bool {
	return a.SourceZone == b.SourceZone &&
		a.Destination == b.Destination &&
		a.CardType == b.CardType &&
		a.Supertype == b.Supertype &&
		a.Reveal == b.Reveal &&
		a.EntersTapped == b.EntersTapped &&
		slices.Equal(a.SubtypesAny, b.SubtypesAny)
}
