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
	// Lively Dirge's second Spree option is a mass-reanimation effect the
	// executable backend does not lower, so the whole card must fail closed even
	// though the Spree structure and the first option are recognized.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Lively Dirge",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{1}{B}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {1} — Search your library for a card, put it into your graveyard, then shuffle.\n" +
			"+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.",
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the unsupported reanimation mode to fail closed")
	}
}
