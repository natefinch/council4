package rules

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCloneReplayProducesIdenticalResults clones a game after setup and runs the
// same deterministic action sequence on both the original and the clone. Because
// land-only games consume no RNG after setup, the engines (seeded identically)
// must produce byte-identical GameResults, proving the clone is a faithful deep
// copy of the live game state.
func TestCloneReplayProducesIdenticalResults(t *testing.T) {
	configs := landOnlyConfigs(8)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: firstLegalAgent{},
		game.Player2: firstLegalAgent{},
		game.Player3: firstLegalAgent{},
		game.Player4: firstLegalAgent{},
	}

	originalEngine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	original := originalEngine.NewGame(configs)
	clone := original.Clone()

	cloneEngine := NewEngine(rand.New(rand.NewPCG(1, 2)))

	originalResult := originalEngine.RunGame(original, agents)
	cloneResult := cloneEngine.RunGame(clone, agents)

	if originalResult.HasWinner != cloneResult.HasWinner {
		t.Fatalf("HasWinner differs: %v != %v", originalResult.HasWinner, cloneResult.HasWinner)
	}
	if originalResult.Winner != cloneResult.Winner {
		t.Fatalf("Winner differs: %v != %v", originalResult.Winner, cloneResult.Winner)
	}
	if originalResult.TurnCount != cloneResult.TurnCount {
		t.Fatalf("TurnCount differs: %d != %d", originalResult.TurnCount, cloneResult.TurnCount)
	}
	if len(originalResult.Events) != len(cloneResult.Events) {
		t.Fatalf("event count differs: %d != %d", len(originalResult.Events), len(cloneResult.Events))
	}
	if !reflect.DeepEqual(originalResult.EndState, cloneResult.EndState) {
		t.Fatalf("end states differ:\n original=%+v\n clone=%+v", originalResult.EndState, cloneResult.EndState)
	}
	if !reflect.DeepEqual(originalResult.Losses, cloneResult.Losses) {
		t.Fatalf("losses differ:\n original=%+v\n clone=%+v", originalResult.Losses, cloneResult.Losses)
	}
	if !reflect.DeepEqual(originalResult.Events, cloneResult.Events) {
		t.Fatal("event streams differ between original and clone")
	}
}

// TestCloneBeforeRunDoesNotShareState verifies that running the original game to
// completion leaves the pre-run clone untouched.
func TestCloneBeforeRunDoesNotShareState(t *testing.T) {
	configs := landOnlyConfigs(8)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: firstLegalAgent{},
		game.Player2: firstLegalAgent{},
		game.Player3: firstLegalAgent{},
		game.Player4: firstLegalAgent{},
	}

	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	original := engine.NewGame(configs)
	clone := original.Clone()

	cloneBattlefieldBefore := len(clone.Battlefield)
	cloneLifeBefore := clone.Players[game.Player1].Life
	cloneLibraryBefore := clone.Players[game.Player1].Library.Size()

	engine.RunGame(original, agents)

	if len(clone.Battlefield) != cloneBattlefieldBefore {
		t.Errorf("clone battlefield changed after running original: %d != %d", len(clone.Battlefield), cloneBattlefieldBefore)
	}
	if clone.Players[game.Player1].Life != cloneLifeBefore {
		t.Errorf("clone life changed after running original: %d != %d", clone.Players[game.Player1].Life, cloneLifeBefore)
	}
	if clone.Players[game.Player1].Library.Size() != cloneLibraryBefore {
		t.Errorf("clone library changed after running original: %d != %d", clone.Players[game.Player1].Library.Size(), cloneLibraryBefore)
	}
}
