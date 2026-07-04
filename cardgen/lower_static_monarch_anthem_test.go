package cardgen

import (
	"strings"
	"testing"
)

// TestLowerMonarchGroupAnthem verifies that a group anthem gated on a live player
// designation ("As long as you're the monarch, permanents you control have
// hexproof.", Dawnglade Regent) lowers onto a conditioned static ability whose
// Condition carries the designation predicate. The runtime re-evaluates the
// condition each time continuous effects are recomputed, so the anthem turns on
// and off as the designation changes.
func TestLowerMonarchGroupAnthem(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Monarch Anthem",
		Layout:     "normal",
		ManaCost:   "{4}{G}{W}",
		TypeLine:   "Creature — Elemental",
		OracleText: "As long as you're the monarch, permanents you control have hexproof.",
		Colors:     []string{"G", "W"},
		Power:      new("5"),
		Toughness:  new("5"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	for _, want := range []string{
		"ControllerIsMonarch: true",
		"AddKeywords: []game.Keyword{",
		"game.Hexproof,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerMonarchGroupAnthemRejectsBoardCondition proves the group-anthem
// condition relaxation is limited to the live player-designation predicates. A
// conditioned group anthem gated on an ordinary board state ("As long as you
// control a Forest, ...") that the runtime does not re-evaluate for group anthems
// stays fail-closed.
func TestLowerMonarchGroupAnthemRejectsBoardCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Board Anthem",
		Layout:     "normal",
		ManaCost:   "{4}{G}{W}",
		TypeLine:   "Creature — Elemental",
		OracleText: "As long as you control a Forest, permanents you control have hexproof.",
		Colors:     []string{"G", "W"},
		Power:      new("5"),
		Toughness:  new("5"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("conditioned board-state group anthem unexpectedly lowered")
	}
}
