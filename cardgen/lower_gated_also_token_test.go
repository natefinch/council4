package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerGatedAlsoTokenCreation proves that a clause-leading "also" on a gated
// sub-effect ("If X is 10 or more, also create ...") no longer blocks token
// creation in the ordered-sequence path. Finale of Glory lowers to two
// CreateToken instructions: the soldiers run unconditionally and the angels run
// gated on the spell's X being at least 10.
func TestLowerGatedAlsoTokenCreation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Finale of Glory",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Create X 2/2 white Soldier creature tokens with vigilance. " +
			"If X is 10 or more, also create X 4/4 white Angel creature tokens with flying and vigilance.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Finale of Glory did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2", len(mode.Sequence))
	}

	soldiers, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("first primitive = %T, want game.CreateToken", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("first token creation is gated, want unconditional")
	}
	_ = soldiers

	angels, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	_ = angels

	gate := mode.Sequence[1].Condition
	if !gate.Exists || !gate.Val.Condition.Exists {
		t.Fatal("second token creation has no effect-condition gate")
	}
	if got := gate.Val.Condition.Val.Aggregates; len(got) != 1 || got[0].Aggregate != game.AggregateSpellX || got[0].Value != 10 {
		t.Fatalf("gate spell-X aggregate = %+v, want spell-X >= 10", gate.Val.Condition.Val.Aggregates)
	}
}
