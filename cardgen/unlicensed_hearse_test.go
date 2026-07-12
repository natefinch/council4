package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerUnlicensedHearseLinkedExileCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Unlicensed Hearse",
		Layout:     "normal",
		TypeLine:   "Artifact — Vehicle",
		ManaCost:   "{2}",
		Power:      new("*"),
		Toughness:  new("*"),
		OracleText: "{T}: Exile up to two target cards from a single graveyard.\nUnlicensed Hearse's power and toughness are each equal to the number of cards exiled with it.\nCrew 2",
	})
	if !face.DynamicPower.Exists ||
		face.DynamicPower.Val.Kind != game.DynamicValueSourceLinkedExileCount ||
		face.DynamicPower.Val.LinkedKey != exiledWithSourceKey ||
		!face.DynamicPower.Val.LinkedObjectScoped ||
		!face.DynamicToughness.Exists ||
		face.DynamicToughness.Val.Kind != game.DynamicValueSourceLinkedExileCount ||
		face.DynamicToughness.Val.LinkedKey != exiledWithSourceKey ||
		!face.DynamicToughness.Val.LinkedObjectScoped {
		t.Fatalf("dynamic P/T = %#v / %#v", face.DynamicPower, face.DynamicToughness)
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two target-card moves", mode.Sequence)
	}
	for i := range mode.Sequence {
		move, ok := mode.Sequence[i].Primitive.(game.MoveCard)
		if !ok || move.PublishLinked != exiledWithSourceKey || !move.PublishLinkedObjectScoped {
			t.Fatalf("sequence[%d] = %#v, want linked MoveCard", i, mode.Sequence[i])
		}
	}
}
