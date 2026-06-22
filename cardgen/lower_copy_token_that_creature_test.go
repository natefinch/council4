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

// TestGenerateExecutableCardSourceCopyTargetToken covers the bare "target
// token" target noun ("Create a token that's a copy of target token you
// control." — Caretaker's Talent's level-2 ability). The target must lower to a
// permanent target restricted to tokens (TokenOnly), not an unrestricted
// permanent.
func TestGenerateExecutableCardSourceCopyTargetToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Token Copier",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Instant",
		OracleText: "Create a token that's a copy of target token you control.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Allow:      game.TargetAllowPermanent,",
		"TokenOnly:  true,",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenOneOfThem covers a "one or more
// other creatures you control enter" trigger whose body copies one of the
// triggering creatures chosen by the controller ("create a token that's a copy
// of one of them." — Twilight Diviner). The copy source must lower to a
// controller-chosen member of the triggering event batch.
func TestGenerateExecutableCardSourceCopyTokenOneOfThem(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Twilight Diviner",
		Layout:   "normal",
		ManaCost: "{2}{B}",
		TypeLine: "Creature — Elf Cleric",
		OracleText: "When this creature enters, surveil 2.\n" +
			"Whenever one or more other creatures you control enter, if they entered or " +
			"were cast from a graveyard, create a token that's a copy of one of them. " +
			"This ability triggers only once each turn.",
		Colors: []string{"B"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceChosenFromTriggerBatch,",
		"MaxTriggersPerTurn: 1,",
		"Primitive: game.Surveil{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
