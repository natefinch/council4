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

func deterministicConfigs() [NumPlayers]PlayerConfig {
	var configs [NumPlayers]PlayerConfig
	for player := range configs {
		for card := 0; card < 10; card++ {
			configs[player].Deck = append(configs[player].Deck, &CardDef{Name: "Card"})
		}
	}
	return configs
}
