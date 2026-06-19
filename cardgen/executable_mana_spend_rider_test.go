package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourcePathOfAncestry covers the full Path of
// Ancestry card: a commander-identity mana ability whose produced mana carries a
// one-shot spend rider that scries 1 when the mana is spent to cast a creature
// spell sharing a creature type with the commander.
func TestGenerateExecutableCardSourcePathOfAncestry(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Path of Ancestry",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "This land enters tapped.\n" +
			"{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1. (Look at the top card of your library. You may put that card on the bottom of your library.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ManaAbilities: []game.ManaAbility",
		"SpendRider: opt.Val(",
		"game.ManaSpendRider{",
		"Condition: game.ManaSpendCastCommanderCreatureType,",
		"game.Scry{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceManaSpendRiderFailsClosed asserts that a
// commander-identity mana ability with a rider the parser does not recognize as
// the exact Path of Ancestry shape (here a different rider effect, "draw a
// card", in place of "scry N") fails closed: it must not lower to a spend-rider
// mana ability and must surface a diagnostic rather than silently dropping the
// rider.
func TestGenerateExecutableCardSourceManaSpendRiderFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Fake Ancestry",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "{T}: Add one mana of any color in your commander's color identity. " +
			"When that mana is spent to cast a creature spell that shares a creature type with your commander, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected fail-closed diagnostic, got source:\n%s", source)
	}
	if strings.Contains(source, "SpendRider: opt.Val(") {
		t.Fatalf("unrecognized rider wrongly lowered to a spend rider:\n%s", source)
	}
}
