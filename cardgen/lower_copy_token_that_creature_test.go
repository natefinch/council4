package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCopyTokenThatTriggerCreature covers a
// single-permanent entry trigger whose body copies the triggering creature
// ("Whenever a nontoken Zombie you control enters, create a token that's a copy
// of that creature." — Necroduality). The "that creature" copy source must bind
// to the triggering permanent rather than the just-created token.
func TestGenerateExecutableCardSourceCopyTokenThatTriggerCreature(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Necroduality",
		Layout:     "normal",
		ManaCost:   "{1}{B}{B}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a nontoken Zombie you control enters, create a token that's a copy of that creature.",
		Colors:     []string{"B"},
	}, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.EventPermanentReference(),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenForEachThatCreature covers a per-each
// copy over a controlled battlefield group whose per-iteration pronoun is "that
// creature" ("For each nontoken creature you control, create a token that's a
// copy of that creature, except it isn't legendary." — Multiversal Incursion).
func TestGenerateExecutableCardSourceCopyTokenForEachThatCreature(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Multiversal Incursion",
		Layout:   "normal",
		ManaCost: "{4}{U}{U}",
		TypeLine: "Sorcery",
		OracleText: "For each nontoken creature you control, create a token that's a copy of " +
			"that creature, except it isn't legendary.",
		Colors: []string{"U"},
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source:          game.TokenCopySourceEachInGroup,",
		"SetNotLegendary: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
