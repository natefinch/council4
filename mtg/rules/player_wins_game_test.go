package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestPlayerWinsGameMarksOpponentsToLose proves that resolving
// PlayerWinsGame{Player: ControllerReference()} marks every other still-active
// player to lose the game (CR 104.2a) without marking the winner, so the
// following state-based-action check eliminates the opponents and leaves the
// controller as the last player standing.
func TestPlayerWinsGameMarksOpponentsToLose(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
	}

	resolveInstruction(engine, g, obj, game.PlayerWinsGame{
		Player: game.ControllerReference(),
	}, &TurnLog{})

	if g.MarkedToLoseGame[game.Player1] {
		t.Fatal("winner Player1 was marked to lose the game")
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if !g.MarkedToLoseGame[opponent] {
			t.Errorf("opponent %v was not marked to lose the game", opponent)
		}
	}

	engine.checkStateBasedActions(g)

	if !g.Players[game.Player1].IsAlive() {
		t.Fatal("winner Player1 was eliminated")
	}
	winner, ok := g.Winner()
	if !ok || winner.ID != game.Player1 {
		t.Fatalf("Winner() = %v, %v, want Player1 as sole survivor", winner, ok)
	}
}
