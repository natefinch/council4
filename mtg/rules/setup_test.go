package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDrawCardMovesTopLibraryCardToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})

	got, ok := engine.drawCard(g, game.Player1, false)
	if !ok {
		t.Fatal("drawCard() ok = false, want true")
	}
	if got != cardID {
		t.Fatalf("drawCard() card ID = %v, want %v", got, cardID)
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

	if _, ok := engine.drawCard(g, game.Player1, false); ok {
		t.Fatal("drawCard() ok = true, want false")
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
			configs[i].Deck[j] = &game.CardDef{CardFace: game.CardFace{Name: "Card"}}
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
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})

	log := TurnLog{}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("active player did not draw on first turn")
	}
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("step = %v, want %v", g.Turn.Step, game.StepDraw)
	}
	if len(log.Draws) != 1 {
		t.Fatalf("draw logs = %d, want 1", len(log.Draws))
	}
	if log.Draws[0].Player != game.Player1 {
		t.Fatalf("draw log player = %v, want %v", log.Draws[0].Player, game.Player1)
	}
	if log.Draws[0].CardID != cardID {
		t.Fatalf("draw log card ID = %v, want %v", log.Draws[0].CardID, cardID)
	}
	if log.Draws[0].Failed {
		t.Fatal("draw log failed = true, want false")
	}
}

func TestBeginningPhaseUntapsAndClearsSummoningSickForActivePlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	activePermanent := addCreaturePermanent(g, game.Player1)
	activePermanent.Tapped = true
	activePermanent.SummoningSick = true
	opponentPermanent := addCreaturePermanent(g, game.Player2)
	opponentPermanent.Tapped = true
	opponentPermanent.SummoningSick = true
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if activePermanent.Tapped {
		t.Fatal("active player's permanent remained tapped")
	}
	if activePermanent.SummoningSick {
		t.Fatal("active player's permanent remained summoning sick")
	}
	if !opponentPermanent.Tapped {
		t.Fatal("opponent's permanent was untapped")
	}
	if !opponentPermanent.SummoningSick {
		t.Fatal("opponent's permanent had summoning sickness cleared")
	}
}

func TestSummoningSickPermanentClearsOnlyOnControllersUntap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCreaturePermanent(g, game.Player1)
	permanent.SummoningSick = true
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player 2 Draw"}})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !permanent.SummoningSick {
		t.Fatal("summoning sickness cleared during another player's untap")
	}
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 Draw"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanent.SummoningSick {
		t.Fatal("summoning sickness remained after controller's untap")
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

func addCreaturePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Test Creature",
			Types: []types.Card{types.Creature}},
		},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}
