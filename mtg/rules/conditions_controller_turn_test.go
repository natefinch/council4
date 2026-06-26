package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestConditionSourceControllerTurn exercises the "During your turn,"
// predicate: it holds only while the static's controller is the active player.
func TestConditionSourceControllerTurn(t *testing.T) {
	condition := opt.Val(game.Condition{SourceControllerTurn: true})

	t.Run("controller is active player", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition did not hold on the controller's own turn")
		}
	})

	t.Run("controller is not active player", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
			t.Fatal("condition held on an opponent's turn")
		}
	})
}
