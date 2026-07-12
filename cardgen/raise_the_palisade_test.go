package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerRaiseThePalisadeChosenTypeBounce(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Raise the Palisade",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{4}{U}",
		OracleText: "Choose a creature type. Return all creatures that aren't of the chosen type to their owners' hands.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok ||
		choose.Choice.Kind != game.ResolutionChoiceSubtype ||
		choose.Choice.SubtypeOfType != types.Creature ||
		choose.PublishChoice != game.SpellChosenTypeChoiceKey {
		t.Fatalf("choice = %#v", mode.Sequence[0].Primitive)
	}
	bounce, ok := mode.Sequence[1].Primitive.(game.Bounce)
	if !ok ||
		!bounce.Group.Valid() ||
		bounce.Group.Selection().SubtypeChoice != game.SubtypeChoiceResolutionExcluded {
		t.Fatalf("bounce = %#v", mode.Sequence[1].Primitive)
	}
}
