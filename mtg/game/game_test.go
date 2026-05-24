package game

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestNewGameWithRandUsesDeterministicShuffle(t *testing.T) {
	configs := deterministicConfigs()

	first := NewGameWithRand(configs, rand.New(rand.NewPCG(1, 2)))
	second := NewGameWithRand(configs, rand.New(rand.NewPCG(1, 2)))

	for i := range first.Players {
		firstLibrary := first.Players[i].Library.All()
		secondLibrary := second.Players[i].Library.All()
		if !slices.Equal(firstLibrary, secondLibrary) {
			t.Fatalf("player %d library order differs: %v != %v", i, firstLibrary, secondLibrary)
		}
	}
}

func TestNewGameWithRandPanicsOnNilRand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewGameWithRand(nil rng) did not panic")
		}
	}()

	NewGameWithRand([NumPlayers]PlayerConfig{}, nil)
}

func TestNewGameCopiesCommanderMetadata(t *testing.T) {
	configs := [NumPlayers]PlayerConfig{
		Player1: {
			Name:         "Metadata Player",
			PowerBracket: "bracket-3",
			PowerLevel:   7,
		},
		Player3: {
			Name:         "Other Metadata Player",
			PowerBracket: "bracket-4",
			PowerLevel:   9,
		},
	}

	g := NewGame(configs)

	if got := g.Players[Player1].PowerBracket; got != "bracket-3" {
		t.Fatalf("power bracket = %q, want bracket-3", got)
	}
	if got := g.Players[Player1].PowerLevel; got != 7 {
		t.Fatalf("power level = %d, want 7", got)
	}
	if got := g.Players[Player2].PowerLevel; got != 0 {
		t.Fatalf("player 2 power level = %d, want zero value", got)
	}
	if got := g.Players[Player3].PowerBracket; got != "bracket-4" {
		t.Fatalf("player 3 power bracket = %q, want bracket-4", got)
	}
	if got := g.Players[Player3].PowerLevel; got != 9 {
		t.Fatalf("player 3 power level = %d, want 9", got)
	}
}

func deterministicConfigs() [NumPlayers]PlayerConfig {
	var configs [NumPlayers]PlayerConfig
	for player := range configs {
		for card := 0; card < 10; card++ {
			configs[player].Deck = append(configs[player].Deck, &CardDef{Name: "Card"})
		}
	}
	return configs
}
