package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerSunderingGrowthPopulate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sundering Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{G/W}{G/W}",
		OracleText: "Destroy target artifact or enchantment, then populate. (Create a token that's a copy of a creature token you control.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || !mode.Targets[0].Selection.Exists ||
		len(mode.Targets[0].Selection.Val.RequiredTypesAny) != 2 {
		t.Fatalf("targets = %#v, want artifact-or-enchantment", mode.Targets)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first instruction = %#v, want destroy", mode.Sequence[0])
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	spec, copyOK := create.Source.TokenCopy()
	if !ok || !copyOK || spec.Source != game.TokenCopySourceChosenControlledCreatureToken ||
		!containsCardType(mode.Targets[0].Selection.Val.RequiredTypesAny, types.Artifact) ||
		!containsCardType(mode.Targets[0].Selection.Val.RequiredTypesAny, types.Enchantment) {
		t.Fatalf("second instruction = %#v, want populate", mode.Sequence[1])
	}
}

func containsCardType(values []types.Card, want types.Card) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
