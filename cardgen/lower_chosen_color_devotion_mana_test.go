package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerNykthosDevotionManaAbility verifies that Nykthos, Shrine to Nyx's
// "Choose a color. Add an amount of mana of that color equal to your devotion to
// that color." lowers to a mana ability that chooses a color and adds that
// chosen color in an amount equal to devotion to the chosen color (ColorFrom
// bound to the published color choice).
func TestLowerNykthosDevotionManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Nykthos, Shrine to Nyx",
		Layout:   "normal",
		TypeLine: "Legendary Land",
		OracleText: "{T}: Add {C}.\n" +
			"{2}, {T}: Choose a color. Add an amount of mana of that color equal to your devotion to that color. " +
			"(Your devotion to a color is the number of mana symbols of that color in the mana costs of permanents you control.)",
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
	if dynamic.Kind != game.DynamicAmountDevotion || dynamic.ColorFrom != choose.PublishChoice {
		t.Fatalf("dynamic amount = %#v, want devotion bound to the chosen color", dynamic)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}
