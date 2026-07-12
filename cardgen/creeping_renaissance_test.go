package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerCreepingRenaissanceChosenPermanentType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Creeping Renaissance",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{G}{G}",
		OracleText: "Choose a permanent type. Return all cards of the chosen type from your graveyard to your hand.\nFlashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want choose then mass return", mode.Sequence)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceCardType {
		t.Fatalf("choose = %#v", mode.Sequence[0].Primitive)
	}
	mass, ok := mode.Sequence[1].Primitive.(game.MassReturnFromGraveyard)
	if !ok ||
		mass.Destination != zone.Hand ||
		mass.Selection.ChosenCardTypeFrom != choose.PublishChoice {
		t.Fatalf("mass return = %#v", mode.Sequence[1].Primitive)
	}
}
