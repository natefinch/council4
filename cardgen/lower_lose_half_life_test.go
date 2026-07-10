package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLoseHalfLife lowers "that player loses half their life, rounded up."
// to a LoseLife whose amount reads the losing player's life halved and rounded
// up, bound to the player dealt combat damage (the event player).
func TestLowerLoseHalfLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Half Life Loss",
		Layout:     "normal",
		TypeLine:   "Creature — Assassin",
		OracleText: "Whenever this creature deals combat damage to a player, that player loses half their life, rounded up.",
	})
	instr := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	lose, ok := instr.Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", instr.Primitive)
	}
	if lose.Player != game.EventPlayerReference() {
		t.Fatalf("player = %#v, want event player", lose.Player)
	}
	dynamic := lose.Amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountPlayerLife ||
		dynamic.Val.Divisor != 2 ||
		!dynamic.Val.RoundUp ||
		dynamic.Val.Player == nil ||
		*dynamic.Val.Player != game.EventPlayerReference() {
		t.Fatalf("amount = %#v, want half the event player's life rounded up", lose.Amount)
	}
}
