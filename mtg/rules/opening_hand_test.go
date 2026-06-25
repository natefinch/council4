package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestNonFrontFaceNames(t *testing.T) {
	mdfc := nonFrontFaceNames(modalDFCSpellLand())
	if got := mdfc[game.FaceBack]; got != "Back Land" {
		t.Fatalf("back face name = %q, want %q", got, "Back Land")
	}
	if _, ok := mdfc[game.FaceFront]; ok {
		t.Fatal("front face should not be included in non-front face names")
	}

	single := &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}}
	if names := nonFrontFaceNames(single); names != nil {
		t.Fatalf("single-faced card face names = %v, want nil", names)
	}
}

func TestRunGoldfishRecordsOpeningHand(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	forest := &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
	}}
	config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: repeatedCard(forest, 99)}

	engine := NewEngine(rand.New(rand.NewPCG(3, 5)))
	g := engine.NewGoldfishGame(config)
	result := engine.RunGoldfish(g, goldfishTestAgent{}, 5)

	if len(result.OpeningHand) != openingHandSize {
		t.Fatalf("opening hand size = %d, want %d", len(result.OpeningHand), openingHandSize)
	}
	for _, cardID := range result.OpeningHand {
		if _, ok := result.Cards[cardID]; !ok {
			t.Fatalf("opening hand card %d missing from Cards", cardID)
		}
	}
}
