package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// stockLibrary gives a player enough cards to survive several draw steps so a
// PlayForward rollout does not end early on an empty-library loss.
func stockPlayForwardLibrary(g *game.Game, playerID game.PlayerID, n int) {
	for range n {
		cardID := g.IDGen.Next()
		g.CardInstances[cardID] = &game.CardInstance{
			ID:    cardID,
			Def:   &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}},
			Owner: playerID,
		}
		g.Players[playerID].Library.Add(cardID)
	}
}

func TestSimulatorPlayForwardFinishesRoundWithoutMutatingOriginal(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for seat := range g.Players {
		stockPlayForwardLibrary(g, game.PlayerID(seat), 10)
	}
	setMainPhasePriority(g, game.Player1)

	beforeTurn := g.Turn.TurnNumber
	beforeActive := g.Turn.ActivePlayer
	beforeBattlefield := len(g.Battlefield)

	// Finish my turn and play every opponent's turn, landing back on my next turn.
	forwarded := e.Simulator().PlayForward(g, simPassPolicies(), game.NumPlayers-1)

	if g.Turn.TurnNumber != beforeTurn || g.Turn.ActivePlayer != beforeActive || len(g.Battlefield) != beforeBattlefield {
		t.Fatalf("PlayForward mutated the original game (turn %d->%d, active %v->%v, battlefield %d->%d)",
			beforeTurn, g.Turn.TurnNumber, beforeActive, g.Turn.ActivePlayer, beforeBattlefield, len(g.Battlefield))
	}
	if forwarded.IsGameOver() {
		t.Fatal("PlayForward ended the game unexpectedly with stocked libraries")
	}
	if forwarded.Turn.ActivePlayer != beforeActive {
		t.Fatalf("after a full round the active player is %v, want %v (back to the searching seat)",
			forwarded.Turn.ActivePlayer, beforeActive)
	}
	if forwarded.Turn.TurnNumber <= beforeTurn {
		t.Fatalf("forwarded turn number %d did not advance past %d", forwarded.Turn.TurnNumber, beforeTurn)
	}
}
