package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerSourceCounterCountManaAbility verifies that "Add {C} for each charge
// counter on this artifact" (Everflowing Chalice) lowers to a mana ability whose
// AddMana amount is the number of charge counters on the source permanent.
func TestLowerSourceCounterCountManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Everflowing Chalice Mana",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add {C} for each charge counter on this artifact.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want single AddMana", sequence)
	}
	add, ok := sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.C || !add.Amount.IsDynamic() {
		t.Fatalf("mana primitive = %#v", sequence[0].Primitive)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountObjectCounters ||
		dynamic.CounterKind != counter.Charge ||
		dynamic.Multiplier != 1 {
		t.Fatalf("dynamic amount = %#v", dynamic)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}
