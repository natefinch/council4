package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestDrawCardMovesTopLibraryCardToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top Card"})

	if !engine.drawCard(g, game.Player1) {
		t.Fatal("drawCard() = false, want true")
	}
	if g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("drawn card remained in library")
	}
	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("drawn card was not added to hand")
	}
	if g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag set after successful draw")
	}
}

func TestDrawCardEmptyLibrarySetsFailedDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if engine.drawCard(g, game.Player1) {
		t.Fatal("drawCard() = true, want false")
	}
	if !g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag was not set")
	}
}

func TestDrawOpeningHandsDrawsSevenCardsPerPlayer(t *testing.T) {
	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		configs[i].Deck = make([]*game.CardDef, openingHandSize)
		for j := range configs[i].Deck {
			configs[i].Deck[j] = &game.CardDef{Name: "Card"}
		}
	}
	g := game.NewGame(configs)
	engine := NewEngine(nil)

	engine.drawOpeningHands(g)

	for i, player := range g.Players {
		if player.Hand.Size() != openingHandSize {
			t.Fatalf("player %d hand size = %d, want %d", i, player.Hand.Size(), openingHandSize)
		}
		if player.Library.Size() != 0 {
			t.Fatalf("player %d library size = %d, want 0", i, player.Library.Size())
		}
	}
}

func TestBeginningPhaseDrawsOnFirstTurnInCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn Card"})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("active player did not draw on first turn")
	}
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("step = %v, want %v", g.Turn.Step, game.StepDraw)
	}
}

func addCardToLibrary(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: playerID,
	}
	g.Players[playerID].Library.Add(cardID)
	return cardID
}
