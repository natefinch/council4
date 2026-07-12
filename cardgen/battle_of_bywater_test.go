package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerBattleOfBywaterPostDestructionFoodCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "The Battle of Bywater",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{1}{W}{W}",
		OracleText: "Destroy all creatures with power 3 or greater. Then create a Food token for each creature you control. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok || !destroy.Group.Valid() || !destroy.Group.Selection().Power.Exists {
		t.Fatalf("destroy = %#v", mode.Sequence[0].Primitive)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	dynamic := create.Amount.DynamicAmount()
	if !ok ||
		!dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountCountSelector ||
		!dynamic.Val.Group.Valid() {
		t.Fatalf("create = %#v", mode.Sequence[1].Primitive)
	}
}
