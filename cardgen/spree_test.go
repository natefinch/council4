package cardgen

import (
	"strings"
	"testing"
)

const explosiveDerailmentSpreeText = "Spree (Choose one or more additional costs.)\n" +
	"+ {2} — Explosive Derailment deals 4 damage to target creature.\n" +
	"+ {2} — Destroy target artifact."

func TestGenerateSpreeSpellSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Explosive Derailment",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: explosiveDerailmentSpreeText,
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Modes: []game.Mode{",
		"Cost: opt.Val(cost.Mana{cost.O(2)}),",
		"MinModes: 1,",
		"MaxModes: 2,",
		"Primitive: game.Damage{",
		"Primitive: game.Destroy",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateSpreeRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()
	// The second Spree option goads a creature, an effect the executable backend
	// does not lower, so the whole card must fail closed even though the Spree
	// structure and the first option are recognized.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Explosive Derailment",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{R}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {2} — Explosive Derailment deals 4 damage to target creature.\n" +
			"+ {2} — Goad target creature.",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the unsupported goad mode to fail closed")
	}
}
