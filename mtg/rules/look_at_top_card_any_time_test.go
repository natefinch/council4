package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// lookAtTopCardAnyTimePermanent gives playerID a battlefield permanent whose
// static ability lets that player privately look at the top card of their
// library at any time (Bolas's Citadel).
func lookAtTopCardAnyTimePermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Citadel",
		StaticAbilities: []game.StaticAbility{
			game.LookAtTopCardAnyTimeStaticBody,
		},
	}})
}

func TestLookAtTopCardAnyTimeRequiresStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Buried Card"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})

	if playerCanLookAtTopCardAnyTime(g, game.Player1) {
		t.Fatal("player may look at top card without the static permission")
	}
	if _, ok := NewObservation(g, game.Player1).LibraryTopLookable(game.Player1); ok {
		t.Fatal("library top is lookable without the static permission")
	}

	lookAtTopCardAnyTimePermanent(g, game.Player1)

	if !playerCanLookAtTopCardAnyTime(g, game.Player1) {
		t.Fatal("player cannot look at top card despite the static permission")
	}
	if playerCanLookAtTopCardAnyTime(g, game.Player2) {
		t.Fatal("opponent gained the look permission they do not control")
	}
}

func TestLookAtTopCardAnyTimePrivateToOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	lookAtTopCardAnyTimePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})

	top, ok := NewObservation(g, game.Player1).LibraryTopLookable(game.Player1)
	if !ok {
		t.Fatal("owner cannot look at the top card despite the static permission")
	}
	if top.Name != "Top Card" {
		t.Fatalf("top card = %q, want %q", top.Name, "Top Card")
	}

	if _, ok := NewObservation(g, game.Player2).LibraryTopLookable(game.Player1); ok {
		t.Fatal("opponent observer may look at a private top card")
	}
}

func TestLookAtTopCardAnyTimeEmptyLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	lookAtTopCardAnyTimePermanent(g, game.Player1)

	if _, ok := NewObservation(g, game.Player1).LibraryTopLookable(game.Player1); ok {
		t.Fatal("an empty library reported a lookable top card")
	}
}
