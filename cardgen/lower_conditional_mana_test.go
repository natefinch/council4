package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerCabalRitualThresholdMana verifies that Cabal Ritual's base and
// "Threshold — ... instead" paragraphs fuse into a single spell whose three
// base {B} productions resolve only when the controller has fewer than seven
// graveyard cards and whose five {B} productions resolve only at threshold.
func TestLowerCabalRitualThresholdMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Cabal Ritual",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Add {B}{B}{B}.\n" +
			"Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Cabal Ritual produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %#v, want 1", modes)
	}
	seq := modes[0].Sequence
	if len(seq) != 8 {
		t.Fatalf("sequence length = %d, want 8 (3 base + 5 threshold)", len(seq))
	}
	var baseCount, thresholdCount int
	for i, instr := range seq {
		add, ok := instr.Primitive.(game.AddMana)
		if !ok {
			t.Fatalf("instruction[%d] = %#v, want AddMana", i, instr.Primitive)
		}
		if add.ManaColor != mana.B {
			t.Fatalf("instruction[%d] color = %v, want black", i, add.ManaColor)
		}
		if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists {
			t.Fatalf("instruction[%d] is ungated: %#v", i, instr)
		}
		cond := instr.Condition.Val.Condition.Val
		if cond.ControllerGraveyardCardCountAtLeast != 7 {
			t.Fatalf("instruction[%d] threshold = %d, want 7", i, cond.ControllerGraveyardCardCountAtLeast)
		}
		if cond.Negate {
			baseCount++
		} else {
			thresholdCount++
		}
	}
	if baseCount != 3 || thresholdCount != 5 {
		t.Fatalf("base=%d threshold=%d, want 3 and 5", baseCount, thresholdCount)
	}
}
