package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func repeatLoseLifeBody() game.AbilityContent {
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(3)},
	}}}.Ability()
}

// TestRepeatProcessFixedTimes proves a RepeatProcess with a fixed count
// re-resolves its body exactly that many times.
func TestRepeatProcessFixedTimes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)

	resolveInstruction(engine, g, obj, game.RepeatProcess{
		Times: game.Fixed(3),
		Body:  repeatLoseLifeBody(),
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 31 {
		t.Fatalf("Player2 life = %d, want 31 (lost 3 life three times)", got)
	}
}

// TestRepeatProcessVariableXTimes proves a RepeatProcess whose count is the
// spell's {X} re-resolves its body XValue times.
func TestRepeatProcessVariableXTimes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)
	obj.XValue = 4

	resolveInstruction(engine, g, obj, game.RepeatProcess{
		Times: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
		Body:  repeatLoseLifeBody(),
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 28 {
		t.Fatalf("Player2 life = %d, want 28 (lost 3 life four times)", got)
	}
}
