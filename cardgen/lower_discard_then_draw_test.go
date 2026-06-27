package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableDiscardThenDrawUpToN proves the variable-looter pipeline
// lowers "discard up to two cards, then draw that many cards" into a single
// DiscardThenDraw primitive bounded at two, modeling Kinetic Augur's enters
// trigger.
func TestGenerateExecutableDiscardThenDrawUpToN(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Probe Looter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		ManaCost:   "{3}{R}",
		Colors:     []string{"R"},
		OracleText: "When this creature enters, discard up to two cards, then draw that many cards.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.DiscardThenDraw{",
		"Player: game.ControllerReference(),",
		"Max:    2,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated looter missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "DrawOffset") {
		t.Fatalf("generated looter unexpectedly carries a DrawOffset:\n%s", source)
	}
}

// TestGenerateExecutableDiscardThenDrawAnyNumberPlusOffset proves the pipeline
// lowers "discard any number of cards, then draw that many cards plus one" into
// an unbounded DiscardThenDraw carrying the draw offset, modeling Colossus of
// the Blood Age.
func TestGenerateExecutableDiscardThenDrawAnyNumberPlusOffset(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Probe Pyre",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{R}",
		Colors:     []string{"R"},
		OracleText: "Discard any number of cards, then draw that many cards plus one.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.DiscardThenDraw{",
		"Player:     game.ControllerReference(),",
		"DrawOffset: 1,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated looter missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Max:") {
		t.Fatalf("generated any-number looter unexpectedly carries a Max bound:\n%s", source)
	}
}
