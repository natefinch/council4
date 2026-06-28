package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceTemporaryCopyTokenActivated covers the
// temporary-copy-token family on an activated ability ("{7}: Create a token
// that's a copy of target artifact. That token gains haste. Exile it at the
// beginning of the next end step." — Cogwork Assembler). The folded "That token
// gains haste." rider sentence sits outside the create clause's own span, so the
// granted keyword and "that token" pronoun must be attributed back to the copy
// effect for the ability to lower; the trailing exile becomes a delayed cleanup.
func TestGenerateExecutableCardSourceTemporaryCopyTokenActivated(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Cogwork Assembler",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact Creature — Assembly-Worker",
		OracleText: "{7}: Create a token that's a copy of target artifact. That token gains haste. Exile it at the beginning of the next end step.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"AddKeywords: []game.Keyword{game.Haste},",
		"Primitive: game.Exile{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceTemporaryCopyTokenTrigger covers the same
// temporary-copy-token family on a triggered ability ("Whenever a nontoken
// creature you control of the chosen type enters, create a token that's a copy
// of that creature. That token gains haste. Exile it at the beginning of the
// next end step." — Molten Echoes), where the exile resolves through a delayed
// trigger scheduled for the next end step.
func TestGenerateExecutableCardSourceTemporaryCopyTokenTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Molten Echoes",
		Layout:   "normal",
		ManaCost: "{2}{R}{R}",
		TypeLine: "Enchantment",
		OracleText: "As this enchantment enters, choose a creature type.\n" +
			"Whenever a nontoken creature you control of the chosen type enters, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.",
		Colors: []string{"R"},
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
		"AddKeywords: []game.Keyword{game.Haste},",
		"Primitive: game.CreateDelayedTrigger{",
		"Primitive: game.Exile{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceOptionalEntersAsCopyOpponentControls covers
// the optional enters-as-copy (Clone) family routed through the optional
// replacement dispatch, copying a permanent an opponent controls with a subtype
// and keyword rider ("You may have this creature enter as a copy of a creature
// an opponent controls, except it's a Faerie Shapeshifter in addition to its
// other types and it has flying." — Malleable Impostor).
func TestGenerateExecutableCardSourceOptionalEntersAsCopyOpponentControls(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Malleable Impostor",
		Layout:   "normal",
		ManaCost: "{3}{U}",
		TypeLine: "Creature — Faerie Shapeshifter",
		OracleText: "Flash\nFlying\n" +
			"You may have this creature enter as a copy of a creature an opponent controls, except it's a Faerie Shapeshifter in addition to its other types and it has flying.",
		Colors: []string{"U"},
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersAsCopyReplacement(",
		"Controller: game.ControllerOpponent",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
