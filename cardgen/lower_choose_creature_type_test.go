package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerChooseCreatureTypeThenDrawForEach verifies "Choose a creature type.
// Draw a card for each permanent you control of that type." (Distant Melody)
// lowers to a Choose instruction publishing the chosen subtype followed by a
// dynamic draw whose count selection reads it back through SubtypeChoiceResolution.
func TestLowerChooseCreatureTypeThenDrawForEach(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Distant Melody",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Choose a creature type. Draw a card for each permanent you control of that type.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	if len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(face.SpellAbility.Val.Modes))
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	choose, ok := sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("first primitive = %#v, want game.Choose", sequence[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceSubtype || choose.Choice.SubtypeOfType != types.Creature {
		t.Fatalf("choice = (%v,%v), want (Subtype, Creature)", choose.Choice.Kind, choose.Choice.SubtypeOfType)
	}
	if choose.PublishChoice != game.SpellChosenTypeChoiceKey {
		t.Fatalf("PublishChoice = %q, want %q", choose.PublishChoice, game.SpellChosenTypeChoiceKey)
	}
	draw, ok := sequence[1].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("second primitive = %#v, want game.Draw", sequence[1].Primitive)
	}
	dynamic := draw.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector || dynamic.Val.Multiplier != 1 {
		t.Fatalf("draw.Amount dynamic = %+v, want count selector x1", dynamic)
	}
	selection := dynamic.Val.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if selection.SubtypeChoice != game.SubtypeChoiceResolution {
		t.Fatal("SubtypeChoice != SubtypeChoiceResolution, want Resolution")
	}
}
