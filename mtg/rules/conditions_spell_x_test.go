package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// TestEffectConditionSpellXAtLeastGatesResolution verifies that an effect gated
// on the "if X is N or more" condition resolves only when the resolving stack
// object's chosen value of {X} meets the threshold. The gated rider (a draw)
// runs when XValue is at least the threshold and is skipped below it, including
// at the boundary just under the threshold.
func TestEffectConditionSpellXAtLeastGatesResolution(t *testing.T) {
	gatedDraw := func() *game.Instruction {
		return &game.Instruction{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: 10}}}),
			}),
		}
	}

	resolve := func(xValue int) int {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		obj := &game.StackObject{
			Kind:       game.StackSpell,
			Controller: game.Player1,
			XValue:     xValue,
		}
		engine.resolveInstructionWithChoices(g, obj, gatedDraw(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return g.Players[game.Player1].Hand.Size()
	}

	if got := resolve(10); got != 1 {
		t.Fatalf("X=10 draw: hand size = %d, want 1", got)
	}
	if got := resolve(9); got != 0 {
		t.Fatalf("X=9 draw: hand size = %d, want 0", got)
	}
	if got := resolve(15); got != 1 {
		t.Fatalf("X=15 draw: hand size = %d, want 1", got)
	}
}
