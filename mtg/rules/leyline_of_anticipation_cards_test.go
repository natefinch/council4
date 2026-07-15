package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLeylineOfAnticipationGrantsFlashTimingToControllerOnly proves the real
// card's "You may cast spells as though they had flash." grants the controller
// instant-speed timing for every spell type, and does not extend that timing to
// opponents.
func TestLeylineOfAnticipationGrantsFlashTimingToControllerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfAnticipation())

	for _, tc := range []struct {
		name string
		def  *game.CardDef
	}{
		{"sorcery", &game.CardDef{CardFace: game.CardFace{Name: "Test Sorcery", Types: []types.Card{types.Sorcery}}}},
		{"creature", &game.CardDef{CardFace: game.CardFace{Name: "Test Creature", Types: []types.Card{types.Creature}}}},
		{"artifact", &game.CardDef{CardFace: game.CardFace{Name: "Test Artifact", Types: []types.Card{types.Artifact}}}},
	} {
		if !playerCanCastAsThoughFlash(g, game.Player1, tc.def) {
			t.Fatalf("controller lacks flash timing for a %s spell", tc.name)
		}
		if playerCanCastAsThoughFlash(g, game.Player2, tc.def) {
			t.Fatalf("opponent gained flash timing for a %s spell from the controller's Leyline", tc.name)
		}
	}
}

// TestLeylineOfAnticipationLetsControllerCastSorceryAtInstantSpeed proves the
// timing permission actually lets a sorcery-speed card be cast when the
// controller would otherwise be unable to (not their main phase), while an
// opponent without the permission still cannot.
func TestLeylineOfAnticipationLetsControllerCastSorceryAtInstantSpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Not the controller's main phase with an empty stack, so sorcery speed is
	// unavailable to Player1 by default.
	g.Turn.ActivePlayer = game.Player2
	sorcery := &game.CardDef{CardFace: game.CardFace{Name: "Test Sorcery", Types: []types.Card{types.Sorcery}}}

	if canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("controller could cast a sorcery at instant speed before any permission")
	}
	addCombatPermanent(g, game.Player1, cards.LeylineOfAnticipation())
	if !canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("Leyline of Anticipation did not let the controller cast a sorcery at instant speed")
	}
	if canCastAtCurrentTiming(g, game.Player2, sorcery) {
		t.Fatal("opponent could cast a sorcery at instant speed without the permission")
	}
}
