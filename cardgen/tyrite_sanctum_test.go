package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerTyriteSanctumPermanentGodThenCounter(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Tyrite Sanctum",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{2}, {T}: Target legendary creature becomes a God in addition to its other types. Put a +1/+1 counter on it.\n{4}, {T}, Sacrifice this land: Put an indestructible counter on target God.",
	})
	if len(face.ActivatedAbilities) != 2 {
		t.Fatalf("activated abilities = %#v, want two nonmana abilities", face.ActivatedAbilities)
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v, want one target and two instructions", mode)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || apply.Duration != game.DurationPermanent ||
		len(apply.ContinuousEffects) != 1 ||
		len(apply.ContinuousEffects[0].AddSubtypes) != 1 ||
		apply.ContinuousEffects[0].AddSubtypes[0] != types.God {
		t.Fatalf("first instruction = %#v, want permanent God subtype", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("second instruction = %#v, want +1/+1 counter", mode.Sequence[1].Primitive)
	}
}
