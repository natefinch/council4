package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdditionalCardsDrawStepTrigger verifies "you may draw two additional
// cards" on a draw-step trigger (the supportable half of Sylvan Library) lowers
// to an optional fixed draw of two cards, treating the "additional" qualifier as
// a plain draw.
func TestLowerAdditionalCardsDrawStepTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Extra Draw Engine",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "At the beginning of your draw step, you may draw two additional cards.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	draw, ok := seq[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 2 {
		t.Fatalf("instruction = %#v, want draw of two", seq[0].Primitive)
	}
	if !ability.Optional {
		t.Fatalf("draw-step trigger should be optional (you may): %#v", ability)
	}
}

// TestLowerMandatoryAdditionalCardsDraw verifies the mandatory "draw two
// additional cards" form lowers identically to a plain fixed draw of two.
func TestLowerMandatoryAdditionalCardsDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bonus Draw Rite",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw two additional cards.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 2 {
		t.Fatalf("instruction = %#v, want draw of two", mode.Sequence[0].Primitive)
	}
}
