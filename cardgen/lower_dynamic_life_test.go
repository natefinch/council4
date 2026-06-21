package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerChildrenOfKorlis proves "Sacrifice this creature: You gain life equal
// to the life you've lost this turn." lowers to an activated ability whose body
// gains the controller life by DynamicAmountLifeLostThisTurn.
func TestLowerChildrenOfKorlis(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Children of Korlis",
		Layout:     "normal",
		TypeLine:   "Creature — Human Rebel Cleric",
		ManaCost:   "{W}",
		OracleText: "Sacrifice this creature: You gain life equal to the life you've lost this turn. (Damage causes loss of life.)",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	prim := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive
	gain, ok := prim.(game.GainLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.GainLife", prim)
	}
	dyn := gain.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountLifeLostThisTurn {
		t.Fatalf("amount = %#v, want DynamicAmountLifeLostThisTurn", gain.Amount)
	}
	if gain.Player != game.ControllerReference() {
		t.Fatalf("player = %#v, want controller", gain.Player)
	}
}
