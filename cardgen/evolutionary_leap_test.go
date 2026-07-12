package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerEvolutionaryLeapRevealUntil(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Evolutionary Leap",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{1}{G}",
		OracleText: "{G}, Sacrifice a creature: Reveal cards from the top of your library until you reveal a creature card. Put that card into your hand and the rest on the bottom of your library in a random order.",
	})
	primitive := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive
	reveal, ok := primitive.(game.RevealUntil)
	if !ok || reveal.Destination != zone.Hand ||
		len(reveal.Until.RequiredTypes) != 1 || reveal.Until.RequiredTypes[0] != types.Creature {
		t.Fatalf("primitive = %#v, want creature reveal to hand", primitive)
	}
}
