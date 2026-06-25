package cardgen

import (
	"strings"
	"testing"
)

// TestLowerMowuSelfCounterAmountReplacement proves Mowu, Loyal Companion's "If
// one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1
// counters are put on it instead." lowers to a self-scoped counter-amount
// replacement (multiplier 0, addend 1) whose recipient is the source permanent
// itself, the self sibling of Conclave Mentor's "a creature you control" group
// counter-amount replacement.
func TestLowerMowuSelfCounterAmountReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mowu, Loyal Companion",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Dog",
		ManaCost:   "{3}{G}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Vigilance, trample\nIf one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.CounterRecipientSelf {
		t.Fatalf("replacement is not self-scoped: %#v", replacement)
	}
	if replacement.CounterMultiplier != 0 || replacement.CounterAddend != 1 {
		t.Fatalf("counter amount = (mul %d, add %d), want (0, 1)", replacement.CounterMultiplier, replacement.CounterAddend)
	}
}

// TestGenerateMowuSelfCounterReplacementSource proves the self counter-amount
// replacement renders through the SelfCounterPlacementReplacement constructor so
// the runtime binds the recipient to the source permanent.
func TestGenerateMowuSelfCounterReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mowu, Loyal Companion",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Dog",
		ManaCost:   "{3}{G}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Vigilance, trample\nIf one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.SelfCounterPlacementReplacement(") {
		t.Fatalf("source missing self counter replacement:\n%s", source)
	}
}
