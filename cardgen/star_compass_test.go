package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerStarCompassBasicLandMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Star Compass",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{2}",
		OracleText: "This artifact enters tapped.\n{T}: Add one mana of any color that a basic land you control could produce.",
	})
	choice, ok := face.ManaAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Choose)
	if !ok || choice.Choice.Selection == nil ||
		len(choice.Choice.Selection.Supertypes) != 1 ||
		choice.Choice.Selection.Supertypes[0] != types.Basic {
		t.Fatalf("choice = %#v, want basic-land production filter", choice)
	}
}
