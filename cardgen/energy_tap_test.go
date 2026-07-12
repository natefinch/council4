package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerEnergyTapManaValue(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Energy Tap",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{U}",
		OracleText: "Tap target untapped creature you control. If you do, add an amount of {C} equal to that creature's mana value.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 ||
		!mode.Targets[0].Selection.Exists ||
		mode.Targets[0].Selection.Val.Controller != game.ControllerYou ||
		mode.Targets[0].Selection.Val.Tapped != game.TriFalse ||
		len(mode.Targets[0].Selection.Val.RequiredTypes) != 1 ||
		mode.Targets[0].Selection.Val.RequiredTypes[0] != types.Creature {
		t.Fatalf("targets = %#v, want controlled untapped creature", mode.Targets)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddMana)
	dynamic := add.Amount.DynamicAmount()
	if !ok || add.ManaColor != mana.C || !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountObjectManaValue ||
		!mode.Sequence[1].ResultGate.Exists {
		t.Fatalf("mana instruction = %#v, want gated target mana-value colorless", mode.Sequence[1])
	}
}
