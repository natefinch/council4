package cardgen

import (
	"slices"
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

// TestLowerUrzaTronConditionalMana verifies that an Urza tron land's "{T}: Add
// {C}. If you control an Urza's Power-Plant and an Urza's Tower, add {C}{C}
// instead." ability lowers to a single mana ability whose base {C} resolves
// only when the controller does not control the named permanents and whose
// {C}{C} bonus resolves only when they do.
func TestLowerUrzaTronConditionalMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Urza's Mine",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "{T}: Add {C}. If you control an Urza's Power-Plant and an Urza's Tower, " +
			"add {C}{C} instead.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("ManaAbilities = %d, want 1", len(face.ManaAbilities))
	}
	modes := face.ManaAbilities[0].Content.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %#v, want 1", modes)
	}
	seq := modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence length = %d, want 3 (1 base + 2 bonus)", len(seq))
	}
	wantNames := []string{"Urza's Power-Plant", "Urza's Tower"}
	var baseCount, bonusCount int
	for i, instr := range seq {
		add, ok := instr.Primitive.(game.AddMana)
		if !ok {
			t.Fatalf("instruction[%d] = %#v, want AddMana", i, instr.Primitive)
		}
		if add.ManaColor != mana.C {
			t.Fatalf("instruction[%d] color = %v, want colorless", i, add.ManaColor)
		}
		if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists {
			t.Fatalf("instruction[%d] is ungated: %#v", i, instr)
		}
		cond := instr.Condition.Val.Condition.Val
		if !slices.Equal(cond.ControllerControlsNamed, wantNames) {
			t.Fatalf("instruction[%d] names = %#v, want %#v", i, cond.ControllerControlsNamed, wantNames)
		}
		if cond.Negate {
			baseCount++
		} else {
			bonusCount++
		}
	}
	if baseCount != 1 || bonusCount != 2 {
		t.Fatalf("base=%d bonus=%d, want 1 and 2", baseCount, bonusCount)
	}
}
