package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerFixedWordCountEntersWithCounters verifies that a self
// enters-with-counters replacement whose fixed quantity above four is written as
// a word numeral ("five", "six", "seven") still lowers to an exact
// EntersWithCounters placement. The parser flags such sentences non-exact only
// because the numeral is not an integer token, but the count is concrete, so
// Pentavus and its siblings must be supported.
func TestLowerFixedWordCountEntersWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Pentavus Clone",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		ManaCost:   "{5}",
		OracleText: "This creature enters with five +1/+1 counters on it.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	placements := face.ReplacementAbilities[0].Replacement.EntersWithCounters
	if len(placements) != 1 {
		t.Fatalf("got %d counter placements, want 1", len(placements))
	}
	if placements[0].Kind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", placements[0].Kind)
	}
	if placements[0].Amount != 5 {
		t.Fatalf("counter amount = %d, want 5", placements[0].Amount)
	}
}

// TestLowerFixedWordCountEntersTappedWithCounters covers the enters-tapped
// variant with a word numeral above four ("six stun counters", Baloth Prime).
func TestLowerFixedWordCountEntersTappedWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Stunned Hulk",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{4}{G}",
		OracleText: "This creature enters tapped with six stun counters on it.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	placements := face.ReplacementAbilities[0].Replacement.EntersWithCounters
	if len(placements) != 1 {
		t.Fatalf("got %d counter placements, want 1", len(placements))
	}
	if placements[0].Kind != counter.Stun {
		t.Fatalf("counter kind = %v, want stun", placements[0].Kind)
	}
	if placements[0].Amount != 6 {
		t.Fatalf("counter amount = %d, want 6", placements[0].Amount)
	}
}
