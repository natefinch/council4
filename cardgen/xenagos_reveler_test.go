package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerXenagosTheReveler(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Xenagos, the Reveler",
		Layout:   "normal",
		TypeLine: "Legendary Planeswalker — Xenagos",
		ManaCost: "{2}{R}{G}",
		Loyalty:  new("3"),
		Colors:   []string{"R", "G"},
		OracleText: "+1: Add X mana in any combination of {R} and/or {G}, where X is the number of creatures you control.\n" +
			"0: Create a 2/2 red and green Satyr creature token with haste.\n" +
			"−6: Exile the top seven cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield.",
	})
	if len(face.LoyaltyAbilities) != 3 {
		t.Fatalf("loyalty abilities = %d, want 3", len(face.LoyaltyAbilities))
	}
	wantCosts := []int{1, 0, -6}
	for i, want := range wantCosts {
		if face.LoyaltyAbilities[i].LoyaltyCost != want {
			t.Fatalf("loyalty ability %d cost = %d, want %d", i, face.LoyaltyAbilities[i].LoyaltyCost, want)
		}
	}

	ultimate := face.LoyaltyAbilities[2].Content.Modes[0].Sequence
	if len(ultimate) != 2 {
		t.Fatalf("ultimate sequence = %#v, want two instructions", ultimate)
	}
	exile, ok := ultimate[0].Primitive.(game.ExileTopOfLibrary)
	if !ok ||
		exile.Amount.Value() != 7 ||
		exile.Player != game.ControllerReference() ||
		exile.PublishLinked != exiledTopCardsLinkKey {
		t.Fatalf("ultimate exile = %#v", ultimate[0].Primitive)
	}
	put, ok := ultimate[1].Primitive.(game.ChooseFromZone)
	if !ok ||
		put.Player != game.ControllerReference() ||
		put.SourceZone != zone.Exile ||
		put.Count != game.ChooseAnyNumber ||
		put.Destination.Zone != zone.Battlefield ||
		put.Riders.FromLinked != exiledTopCardsLinkKey ||
		!slices.Equal(put.Filter.RequiredTypesAny, []types.Card{types.Creature, types.Land}) {
		t.Fatalf("ultimate put = %#v", ultimate[1].Primitive)
	}
}

func TestLowerExileTopThenPutAnyAmongToBattlefieldSingleType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Exile Dig",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Exile the top three cards of your library. You may put any number of artifact cards from among them onto the battlefield.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	put, ok := sequence[1].Primitive.(game.ChooseFromZone)
	if !ok || !slices.Equal(put.Filter.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("put = %#v, want artifact filter", sequence[1].Primitive)
	}
}
