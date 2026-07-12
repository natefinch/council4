package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerTemurSabertoothOptionalReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Temur Sabertooth",
		Layout:     "normal",
		TypeLine:   "Creature — Cat",
		ManaCost:   "{2}{G}{G}",
		Power:      new("4"),
		Toughness:  new("3"),
		OracleText: "{1}{G}: You may return another creature you control to its owner's hand. If you do, this creature gains indestructible until end of turn.",
	})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	bounce, ok := mode.Sequence[0].Primitive.(game.Bounce)
	if !ok ||
		!bounce.ControlledChoice ||
		bounce.Amount.Value() != 1 ||
		!mode.Sequence[0].Optional ||
		mode.Sequence[0].PublishResult == "" {
		t.Fatalf("optional bounce = %#v", mode.Sequence[0])
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok ||
		!mode.Sequence[1].ResultGate.Exists ||
		mode.Sequence[1].ResultGate.Val.Key != mode.Sequence[0].PublishResult ||
		mode.Sequence[1].ResultGate.Val.Succeeded != game.TriTrue ||
		apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("gated indestructible = %#v", mode.Sequence[1])
	}
}
