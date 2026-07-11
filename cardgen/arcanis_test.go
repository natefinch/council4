package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerArcanisSelfReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Arcanis the Omnipotent",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Wizard",
		ManaCost:   "{3}{U}{U}{U}",
		OracleText: "{T}: Draw three cards.\n{2}{U}{U}: Return Arcanis to its owner's hand.",
	})
	if len(face.ActivatedAbilities) != 2 {
		t.Fatalf("activated abilities = %#v, want two", face.ActivatedAbilities)
	}
	primitive := face.ActivatedAbilities[1].Content.Modes[0].Sequence[0].Primitive
	if _, ok := primitive.(game.Bounce); !ok {
		t.Fatalf("primitive = %#v, want self bounce", primitive)
	}
}
