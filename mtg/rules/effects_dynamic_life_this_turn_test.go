package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDynamicLifeLostThisTurnSumsControllerLifeLoss covers Children of Korlis's
// "You gain life equal to the life you've lost this turn.": a
// DynamicAmountLifeLostThisTurn sums only the controller's life-loss this turn,
// including damage (which the rules apply as life loss), and ignores the
// opponent's losses and the controller's gains.
func TestDynamicLifeLostThisTurnSumsControllerLifeLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	amount := game.DynamicAmount{Kind: game.DynamicAmountLifeLostThisTurn}

	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 0 {
		t.Fatalf("life lost this turn = %d, want 0 before any loss", got)
	}

	loseLife(g, game.Player1, 5)
	loseLife(g, game.Player1, 3)
	loseLife(g, game.Player2, 7)
	gainLife(g, game.Player1, 4)

	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 8 {
		t.Fatalf("life lost this turn = %d, want 8 (controller's 5 + 3)", got)
	}

	withMultiplier := game.DynamicAmount{Kind: game.DynamicAmountLifeLostThisTurn, Multiplier: 2}
	if got := dynamicAmountValue(g, obj, game.Player1, withMultiplier); got != 16 {
		t.Fatalf("twice life lost this turn = %d, want 16", got)
	}
}

// TestDynamicLifeGainedThisTurnSumsControllerLifeGain covers the life-gained
// sibling "the amount of life you gained this turn": a
// DynamicAmountLifeGainedThisTurn sums only the controller's life gains this
// turn.
func TestDynamicLifeGainedThisTurnSumsControllerLifeGain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	amount := game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn}

	gainLife(g, game.Player1, 2)
	gainLife(g, game.Player1, 6)
	gainLife(g, game.Player2, 9)
	loseLife(g, game.Player1, 3)

	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 8 {
		t.Fatalf("life gained this turn = %d, want 8 (controller's 2 + 6)", got)
	}
}

// TestGainLifeEqualToLifeLostThisTurnResolves proves the full effect path:
// after the controller has lost life this turn, resolving a GainLife whose
// amount is DynamicAmountLifeLostThisTurn restores exactly the life lost.
func TestGainLifeEqualToLifeLostThisTurnResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	loseLife(g, game.Player1, 6)
	loseLife(g, game.Player1, 4)
	if g.Players[game.Player1].Life != 30 {
		t.Fatalf("setup life = %d, want 30 after losing 10", g.Players[game.Player1].Life)
	}

	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountLifeLostThisTurn}),
		Player: game.ControllerReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("player 1 life = %d, want 40 (gained back the 10 lost)", g.Players[game.Player1].Life)
	}
}
