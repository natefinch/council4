package cardgen

import "testing"

// TestLowerDevourReplacement verifies that the printed Devour keyword lowers to
// a game.DevourReplacement carrying the per-sacrificed-creature multiplier (CR
// 702.81).
func TestLowerDevourReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Thunder-Thrash Elder",
		Layout:     "normal",
		TypeLine:   "Creature — Dinosaur",
		OracleText: "Devour 3 (As this creature enters, you may sacrifice any number of creatures. It enters with three times that many +1/+1 counters on it.)",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.EntryDevourMultiplier != 3 {
		t.Fatalf("EntryDevourMultiplier = %d, want 3", replacement.EntryDevourMultiplier)
	}
}
