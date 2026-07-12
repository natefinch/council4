package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerSelesnyaEulogistExileThenPopulate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Selesnya Eulogist",
		Layout:     "normal",
		TypeLine:   "Creature — Centaur Druid",
		ManaCost:   "{2}{G}",
		OracleText: "{2}{G}: Exile target creature card from a graveyard, then populate. (Create a token that's a copy of a creature token you control.)",
	})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || !mode.Targets[0].Selection.Exists ||
		mode.Targets[0].TargetZone != zone.Graveyard ||
		len(mode.Targets[0].Selection.Val.RequiredTypes) != 1 ||
		mode.Targets[0].Selection.Val.RequiredTypes[0] != types.Creature {
		t.Fatalf("targets = %#v, want graveyard creature card", mode.Targets)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("first instruction = %#v, want graveyard exile", mode.Sequence[0])
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	spec, copyOK := create.Source.TokenCopy()
	if !ok || !copyOK || spec.Source != game.TokenCopySourceChosenControlledCreatureToken {
		t.Fatalf("second instruction = %#v, want populate", mode.Sequence[1])
	}
}
