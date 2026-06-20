package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRunGoldfishCompletesExactTurnLimit(t *testing.T) {
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
	config := game.PlayerConfig{
		Name:      "Goldfish",
		Commander: commander,
		Deck:      repeatedCard(forest, 99),
	}

	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	g := engine.NewGoldfishGame(config)
	players := NewObservation(g, game.Player1).Players()
	if len(players) != 1 || players[0].ID != game.Player1 {
		t.Fatalf("observed players = %#v", players)
	}
	result := engine.RunGoldfish(g, goldfishTestAgent{}, 10)
	if result.TurnCount != 10 || len(result.Turns) != 10 {
		t.Fatalf("turns = %d/%d, want 10", result.TurnCount, len(result.Turns))
	}

	if !result.TurnLimitReached {
		t.Fatal("turn limit was not reached")
	}
	if result.HasWinner {
		t.Fatal("goldfish run has a multiplayer winner")
	}
	for index, turn := range result.Turns {
		if turn.TurnNumber != index+1 || turn.ActivePlayer != game.Player1 {
			t.Fatalf("turn %d = number %d player %d", index, turn.TurnNumber, turn.ActivePlayer)
		}
	}
	if result.EndState.Players[game.Player1].LibrarySize != 82 {
		t.Fatalf("library size = %d, want 82", result.EndState.Players[game.Player1].LibrarySize)
	}
	for playerID := game.Player2; playerID < game.NumPlayers; playerID++ {
		if !result.EndState.Players[playerID].Eliminated {
			t.Fatalf("inactive seat %d is not eliminated", playerID)
		}
	}
}

func TestRunGoldfishStopsIfSolePlayerLoses(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Card"}}
	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	g := engine.NewGoldfishGame(game.PlayerConfig{
		Name: "Goldfish",
		Deck: repeatedCard(card, 7),
	})
	result := engine.RunGoldfish(g, goldfishTestAgent{}, 10)
	if result.TurnCount != 1 {
		t.Fatalf("turn count = %d, want 1", result.TurnCount)
	}
	if result.TurnLimitReached {
		t.Fatal("loss reported as turn-limit completion")
	}
	if !result.EndState.Players[game.Player1].Eliminated {
		t.Fatal("failed draw did not eliminate the goldfish player")
	}
}

type goldfishTestAgent struct{}

func (goldfishTestAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

func repeatedCard(card *game.CardDef, count int) []*game.CardDef {
	cards := make([]*game.CardDef, count)
	for index := range cards {
		cards[index] = card
	}
	return cards
}
