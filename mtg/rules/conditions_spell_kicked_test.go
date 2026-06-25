package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEffectConditionSpellWasKickedGatesResolution verifies that an effect gated
// on the "if this spell was kicked" condition resolves only when the resolving
// stack object recorded a paid kicker cost. The kicked rider (a draw) runs when
// KickerPaid is true and is skipped otherwise, so the base form applies when the
// spell was not kicked.
func TestEffectConditionSpellWasKickedGatesResolution(t *testing.T) {
	gatedDraw := func() *game.Instruction {
		return &game.Instruction{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{SpellWasKicked: true}),
			}),
		}
	}

	resolve := func(kickerPaid, isCopy bool) int {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		obj := &game.StackObject{
			Kind:       game.StackSpell,
			Controller: game.Player1,
			KickerPaid: kickerPaid,
			Copy:       isCopy,
		}
		engine.resolveInstructionWithChoices(g, obj, gatedDraw(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return g.Players[game.Player1].Hand.Size()
	}

	if got := resolve(true, false); got != 1 {
		t.Fatalf("kicked draw: hand size = %d, want 1", got)
	}
	if got := resolve(false, false); got != 0 {
		t.Fatalf("unkicked draw: hand size = %d, want 0", got)
	}
	if got := resolve(true, true); got != 0 {
		t.Fatalf("copied kicked spell: hand size = %d, want 0 (copies are never kicked)", got)
	}
}
