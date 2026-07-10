package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestEventStackObjectControllerReference(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
	}
	g.Stack.Push(target)
	resolving := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{StackObjectID: target.ID},
	}
	got, ok := resolvePlayerReference(
		g,
		resolving,
		game.ObjectControllerReference(game.EventStackObjectReference()),
	)
	if !ok || got != game.Player2 {
		t.Fatalf("controller = (%v, %v), want player 2", got, ok)
	}
}
