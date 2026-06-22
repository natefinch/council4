package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerMyriadLandscapeEndToEnd verifies the full anchor card compiles to a
// faithful runtime plan: its activated ability keeps the {2} mana cost, the tap
// cost, and the sacrifice-source cost, and its effect is a correlated
// up-to-two-basic-lands search that enters the battlefield tapped with the
// shared-land-type constraint set. This guards the whole pipeline end to end.
func TestLowerMyriadLandscapeEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Myriad Landscape",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{2}, {T}, Sacrifice this land: Search your library for up to two basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || len(ability.ManaCost.Val) != 1 {
		t.Fatalf("ManaCost = %#v, want {2}", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 2 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalTap ||
		ability.AdditionalCosts[1].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("AdditionalCosts = %#v, want [tap, sacrifice source]", ability.AdditionalCosts)
	}
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("ability content modes = %#v, want a single-instruction mode", ability.Content.Modes)
	}
	search, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("ability primitive = %#v, want game.Search", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if got := search.Amount.Value(); got != 2 {
		t.Fatalf("search amount = %d, want 2", got)
	}
	want := game.SearchSpec{
		SourceZone:    zone.Library,
		Destination:   zone.Battlefield,
		EntersTapped:  true,
		SharedSubtype: true,
		Filter: game.Selection{
			RequiredTypes: []types.Card{types.Land},
			Supertypes:    []types.Super{types.Basic},
		},
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Fatalf("search spec = %#v, want %#v", search.Spec, want)
	}
}
