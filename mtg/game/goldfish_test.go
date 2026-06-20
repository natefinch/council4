package game

import (
	"math/rand/v2"
	"testing"
)

func TestNewGoldfishGameHasOneActivePlayer(t *testing.T) {
	config := PlayerConfig{Name: "Goldfish"}
	g := NewGoldfishGameWithRand(config, rand.New(rand.NewPCG(1, 2)))
	if g.Mode != RunModeGoldfish {
		t.Fatalf("mode = %v", g.Mode)
	}
	if got := g.AlivePlayers(); len(got) != 1 || got[0].ID != Player1 {
		t.Fatalf("alive players = %#v", got)
	}
	if g.IsGameOver() {
		t.Fatal("fresh goldfish game is over")
	}
	if _, ok := g.Winner(); ok {
		t.Fatal("goldfish game has a multiplayer winner")
	}
	for playerID := Player2; playerID < NumPlayers; playerID++ {
		if !g.Players[playerID].Eliminated || !g.TurnOrder.IsEliminated(playerID) {
			t.Fatalf("seat %d is active", playerID)
		}
	}
}
