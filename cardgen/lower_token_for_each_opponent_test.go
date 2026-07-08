package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCreateTokenForEachOpponentUnconditional verifies that the
// unconditional per-opponent distributive "For each opponent, you create <one
// token>." scales the token count by the opponent count rather than flattening
// it to a fixed single token. It backs Endless Ranks of HYDRA, whose printed
// per-opponent count is one, so the create lowers to a CreateToken whose amount
// is the dynamic opponent count (one token per opponent).
func TestLowerCreateTokenForEachOpponentUnconditional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Endless Ranks of HYDRA",
		Layout:     "normal",
		ManaCost:   "{3}{B}",
		TypeLine:   "Sorcery",
		OracleText: "For each opponent, you create a 2/1 black Villain creature token with menace.",
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %+v, want a dynamic opponent count (not a fixed count)", create.Amount)
	}
	dynamic := create.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountOpponentCount {
		t.Fatalf("amount kind = %v, want DynamicAmountOpponentCount", dynamic.Kind)
	}
	if dynamic.Multiplier != 1 {
		t.Fatalf("amount multiplier = %d, want 1 (one token per opponent)", dynamic.Multiplier)
	}
}

// TestLowerCreateTokenForEachOpponentConditionalFailsClosed verifies that the
// conditional per-opponent distributive "For each opponent who <condition>,
// create ..." fails closed rather than silently emitting a fixed count that
// ignores the "who <condition>" restriction. It backs Faerie Slumber Party,
// whose "For each opponent who controlled a creature returned this way, you
// create two ... tokens" is not modeled: the whole spell must report an
// unsupported token creation instead of creating a fixed two tokens.
func TestLowerCreateTokenForEachOpponentConditionalFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Faerie Slumber Party",
		Layout:   "normal",
		ManaCost: "{4}{U}{U}",
		TypeLine: "Sorcery",
		OracleText: "Return all creatures to their owners' hands. For each opponent who controlled a creature " +
			"returned this way, you create two 1/1 blue Faerie creature tokens with flying and " +
			"\"This token can block only creatures with flying.\"",
	})
	if face.SpellAbility.Exists {
		t.Fatal("spell ability lowered, want none: unmodeled conditional distributive must fail closed")
	}
}

// TestLowerCreateTokenForEachOpponentPluralCountFailsClosed verifies that an
// unconditional per-opponent distributive whose printed per-opponent count is
// greater than one ("For each opponent, you create two <token>.") fails closed:
// the modeled path scales only a single printed token per opponent, so a
// multi-token count must not silently flatten to a fixed pair.
func TestLowerCreateTokenForEachOpponentPluralCountFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Plural Per Opponent",
		Layout:     "normal",
		ManaCost:   "{3}{B}",
		TypeLine:   "Sorcery",
		OracleText: "For each opponent, you create two 1/1 white Human creature tokens.",
	})
	if face.SpellAbility.Exists {
		t.Fatal("spell ability lowered, want none: unmodeled N>1 per-opponent count must fail closed")
	}
}
