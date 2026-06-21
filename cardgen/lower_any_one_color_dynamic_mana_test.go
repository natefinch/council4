package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerKamiAnyOneColorDynamicMana verifies that Kami of Whispered Hopes'
// "{T}: Add X mana of any one color, where X is this creature's power." lowers
// to a mana ability that chooses a color and adds that chosen color in an amount
// equal to the source creature's power.
func TestLowerKamiAnyOneColorDynamicMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Kami of Whispered Hopes",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Snake",
		ManaCost: "{G}{G}",
		OracleText: "If one or more +1/+1 counters would be put on a permanent you control, that many plus one +1/+1 counters are put on that permanent instead.\n" +
			"{T}: Add X mana of any one color, where X is this creature's power.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
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
	if dynamic.Kind != game.DynamicAmountObjectPower {
		t.Fatalf("dynamic amount kind = %v, want object power", dynamic.Kind)
	}
	if len(dynamic.Object.Validate()) != 0 {
		t.Fatalf("dynamic amount object invalid: %#v", dynamic.Object)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerAnyOneColorDynamicManaDevotion verifies the same any-one-color
// dynamic-mana path also lowers a devotion amount ("where X is your devotion to
// green"), so the category generalizes over the dynamic amount rather than
// binding to source power alone (Karametra's Acolyte).
func TestLowerAnyOneColorDynamicManaDevotion(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Karametra's Acolyte",
		Layout:     "normal",
		TypeLine:   "Creature — Human Druid",
		ManaCost:   "{3}{G}",
		OracleText: "{T}: Add X mana of any one color, where X is your devotion to green.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	add, ok := sequence[1].Primitive.(game.AddMana)
	if !ok || !add.Amount.IsDynamic() {
		t.Fatalf("second primitive = %#v", sequence[1].Primitive)
	}
	if dynamic := add.Amount.DynamicAmount().Val; dynamic.Kind != game.DynamicAmountDevotion {
		t.Fatalf("dynamic amount kind = %v, want devotion", dynamic.Kind)
	}
}
