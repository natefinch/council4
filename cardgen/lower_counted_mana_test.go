package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerCountedSingleManaSymbol verifies that a counted single-symbol
// add-mana body ("Add six {R}") lowers to N copies of the fixed mana primitive.
// It exercises The Flux's final Saga chapter end to end, confirming the parser
// expands the "<N> {X}" shorthand into the repeated symbol sequence the
// executable backend already supports.
func TestLowerCountedSingleManaSymbol(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "The Flux",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VI.)\n" +
			"I — This Saga deals 4 damage to target creature an opponent controls.\n" +
			"II, III, IV, V — Exile the top card of your library. You may play that card this turn.\n" +
			"VI — Add six {R}.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("got %d chapter abilities, want 3", len(face.ChapterAbilities))
	}
	final := face.ChapterAbilities[2]
	if !slices.Equal(final.Chapters, []int{6}) {
		t.Fatalf("final chapter numbers = %v, want [6]", final.Chapters)
	}
	seq := final.Content.Modes[0].Sequence
	if len(seq) != 6 {
		t.Fatalf("final chapter sequence length = %d, want 6", len(seq))
	}
	for i, instr := range seq {
		add, ok := instr.Primitive.(game.AddMana)
		if !ok {
			t.Fatalf("final chapter primitive[%d] = %T, want game.AddMana", i, instr.Primitive)
		}
		if add.ManaColor != mana.R {
			t.Errorf("final chapter primitive[%d] color = %v, want red", i, add.ManaColor)
		}
		if add.Amount != game.Fixed(1) {
			t.Errorf("final chapter primitive[%d] amount = %#v, want 1", i, add.Amount)
		}
	}
}
