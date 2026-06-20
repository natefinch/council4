package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerThreeTreeCityChosenColorCountMana verifies that Three Tree City's
// "Choose a color. Add an amount of mana of that color equal to the number of
// creatures you control of the chosen type." lowers to a mana ability that
// chooses a color and adds that chosen color in an amount equal to the count of
// the controller's creatures of the source's entry-time chosen creature type.
func TestLowerThreeTreeCityChosenColorCountMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Three Tree City",
		Layout:   "normal",
		TypeLine: "Legendary Land",
		OracleText: "As Three Tree City enters, choose a creature type.\n" +
			"{T}: Add {C}.\n" +
			"{2}, {T}: Choose a color. Add an amount of mana of that color equal to the number of creatures you control of the chosen type.",
	})
	if len(face.ManaAbilities) != 2 {
		t.Fatalf("mana abilities = %d, want 2", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[1].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want Choose then AddMana", sequence)
	}
	choose, ok := sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("first primitive = %#v, want a color Choose", sequence[0].Primitive)
	}
	add, ok := sequence[1].Primitive.(game.AddMana)
	if !ok || !add.Amount.IsDynamic() || add.ChoiceFrom != choose.PublishChoice {
		t.Fatalf("second primitive = %#v", sequence[1].Primitive)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("dynamic amount kind = %v, want count selector", dynamic.Kind)
	}
	selection := dynamic.Group.Selection()
	if !selection.SubtypeFromSourceEntryChoice {
		t.Fatalf("count selection SubtypeFromSourceEntryChoice not set: %#v", selection)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("count selection controller = %v, want you", selection.Controller)
	}
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("count selection required types = %#v, want creature", selection.RequiredTypes)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}
