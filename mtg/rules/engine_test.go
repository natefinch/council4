package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestEngineNewGameUsesEngineRand(t *testing.T) {
	configs := engineTestConfigs()
	first := NewEngine(rand.New(rand.NewPCG(1, 2))).NewGame(configs)
	second := NewEngine(rand.New(rand.NewPCG(1, 2))).NewGame(configs)

	for i := range first.Players {
		if !slices.Equal(first.Players[i].Library.All(), second.Players[i].Library.All()) {
			t.Fatalf("player %d library order differs", i)
		}
	}
}

func TestRunGameLandOnlyDecksTerminateWithWinner(t *testing.T) {
	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	gameState := engine.NewGame(landOnlyConfigs(8))
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: firstLegalAgent{},
		game.Player2: firstLegalAgent{},
		game.Player3: firstLegalAgent{},
		game.Player4: firstLegalAgent{},
	}

	result := engine.RunGame(gameState, agents)

	if !result.HasWinner {
		t.Fatal("RunGame() did not produce a winner")
	}
	if result.Winner != game.Player4 {
		t.Fatalf("winner = %v, want %v", result.Winner, game.Player4)
	}
	if result.TurnCount == 0 {
		t.Fatal("TurnCount = 0, want > 0")
	}
	if result.TurnCount >= maxGameTurns {
		t.Fatalf("TurnCount = %d, want < %d", result.TurnCount, maxGameTurns)
	}
	if len(result.Turns) != result.TurnCount {
		t.Fatalf("turn logs = %d, want %d", len(result.Turns), result.TurnCount)
	}
	if len(gameState.Battlefield) == 0 {
		t.Fatal("battlefield is empty; expected at least one land to be played")
	}
}

func TestRunGameOpeningHandFailedDrawCanEndGame(t *testing.T) {
	configs := landOnlyConfigs(0)
	configs[game.Player4].Deck = make([]*game.CardDef, openingHandSize)
	for i := range configs[game.Player4].Deck {
		configs[game.Player4].Deck[i] = basicLand()
	}
	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	gameState := engine.NewGame(configs)

	result := engine.RunGame(gameState, [game.NumPlayers]PlayerAgent{})

	if !result.HasWinner {
		t.Fatal("RunGame() did not produce a winner")
	}
	if result.Winner != game.Player4 {
		t.Fatalf("winner = %v, want %v", result.Winner, game.Player4)
	}
	if result.TurnCount != 0 {
		t.Fatalf("TurnCount = %d, want 0", result.TurnCount)
	}
	if len(result.Losses) != 3 {
		t.Fatalf("losses = %d, want 3", len(result.Losses))
	}
	for _, loss := range result.Losses {
		if loss.Reason != LossReasonEmptyLibraryDraw {
			t.Fatalf("loss reason = %q, want %q", loss.Reason, LossReasonEmptyLibraryDraw)
		}
	}
}

func TestRunGameDeterministicWithSameSeed(t *testing.T) {
	configs := landOnlyConfigs(8)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: firstLegalAgent{},
		game.Player2: firstLegalAgent{},
		game.Player3: firstLegalAgent{},
		game.Player4: firstLegalAgent{},
	}
	firstEngine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	secondEngine := NewEngine(rand.New(rand.NewPCG(1, 2)))

	first := firstEngine.RunGame(firstEngine.NewGame(configs), agents)
	second := secondEngine.RunGame(secondEngine.NewGame(configs), agents)

	if !first.HasWinner {
		t.Fatal("first run did not produce a winner")
	}
	if first.HasWinner != second.HasWinner {
		t.Fatalf("HasWinner differs: %v != %v", first.HasWinner, second.HasWinner)
	}
	if first.Winner != second.Winner {
		t.Fatalf("Winner differs: %v != %v", first.Winner, second.Winner)
	}
	if first.TurnCount != second.TurnCount {
		t.Fatalf("TurnCount differs: %v != %v", first.TurnCount, second.TurnCount)
	}
}

func engineTestConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		for card := 0; card < 10; card++ {
			configs[player].Deck = append(configs[player].Deck, &game.CardDef{Name: "Card"})
		}
	}
	return configs
}

type firstLegalAgent struct{}

func (firstLegalAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}

func landOnlyConfigs(cardsPerPlayer int) [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		for card := 0; card < cardsPerPlayer; card++ {
			configs[player].Deck = append(configs[player].Deck, basicLand())
		}
	}
	return configs
}

func basicLand() *game.CardDef {
	return &game.CardDef{
		Name:     "Forest",
		Types:    []game.CardType{game.TypeLand},
		Subtypes: []string{"Forest"},
	}
}
