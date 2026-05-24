package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCommanderFirstMulliganIsFree(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 14)
	engine.drawOpeningHands(g)

	if !engine.performCommanderMulligan(g, game.Player1) {
		t.Fatal("performCommanderMulligan() = false, want true")
	}

	player := g.Players[game.Player1]
	if player.CommanderMulligansTaken != 1 {
		t.Fatalf("mulligans taken = %d, want 1", player.CommanderMulligansTaken)
	}
	if player.Hand.Size() != 7 {
		t.Fatalf("hand size after first Commander mulligan = %d, want 7", player.Hand.Size())
	}
}

func TestCommanderSecondMulliganBottomsOneCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 21)
	engine.drawOpeningHands(g)
	if !engine.performCommanderMulligan(g, game.Player1) {
		t.Fatal("first mulligan failed")
	}

	if !engine.performCommanderMulligan(g, game.Player1) {
		t.Fatal("second mulligan failed")
	}

	player := g.Players[game.Player1]
	if player.CommanderMulligansTaken != 2 {
		t.Fatalf("mulligans taken = %d, want 2", player.CommanderMulligansTaken)
	}
	if player.Hand.Size() != 6 {
		t.Fatalf("hand size after second Commander mulligan = %d, want 6", player.Hand.Size())
	}
}

func fillLibrary(g *game.Game, playerID game.PlayerID, count int) {
	for i := range count {
		addCardToLibrary(g, playerID, &game.CardDef{Name: "Mulligan Card " + string(rune('A'+i))})
	}
}
