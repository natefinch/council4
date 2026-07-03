package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestTargetChoiceRequestCarriesTargetReferences(t *testing.T) {
	choices := [][]game.Target{
		{game.PermanentTarget(id.ID(7))},
		{game.PermanentTarget(id.ID(9))},
	}
	request := targetChoiceRequest(game.Player1, "Choose target", choices)
	if len(request.Options) != 2 {
		t.Fatalf("got %d options, want 2", len(request.Options))
	}
	if len(request.Options[1].Targets) != 1 || request.Options[1].Targets[0].PermanentID != id.ID(9) {
		t.Fatalf("option 1 targets = %+v, want the second permanent", request.Options[1].Targets)
	}
}
