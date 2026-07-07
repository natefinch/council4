package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCourtOfVantress covers Court of Vantress's
// upkeep trigger: "At the beginning of your upkeep, choose up to one other
// target enchantment or artifact. If you're the monarch, you may create a token
// that's a copy of it. If you're not the monarch, you may have this enchantment
// become a copy of it, except it has this ability." The body lowers to one
// shared up-to-one "other" artifact-or-enchantment target and two optional "you
// may" instructions over it: a ControllerIsMonarch-gated CreateToken copy of the
// chosen permanent, and a not-monarch-gated BecomeCopy of that same permanent
// whose copy retains Court of Vantress's own upkeep ability. The whole ability
// lowers without diagnostics.
func TestGenerateExecutableCardSourceCourtOfVantress(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Court of Vantress",
		Layout:   "normal",
		ManaCost: "{2}{U}{U}",
		TypeLine: "Enchantment",
		OracleText: "When this enchantment enters, you become the monarch.\n" +
			"At the beginning of your upkeep, choose up to one other target enchantment or artifact. If you're the monarch, you may create a token that's a copy of it. If you're not the monarch, you may have this enchantment become a copy of it, except it has this ability.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`Constraint: "up to one other target enchantment or artifact",`,
		"game.Selection{RequiredTypesAny: []types.Card{types.Enchantment, types.Artifact}, ExcludeSource: true}",
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.TargetPermanentReference(0),",
		"Primitive: game.BecomeCopy{",
		"RetainsThisAbility: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	// The monarch branch is gated on the plain designation and the become-a-copy
	// branch on its negation; both branches are optional "you may" instructions.
	if strings.Count(source, "ControllerIsMonarch: true,") != 2 {
		t.Fatalf("expected two ControllerIsMonarch gates:\n%s", source)
	}
	if !strings.Contains(source, "Negate:              true,") {
		t.Fatalf("become-a-copy branch is not negated-monarch gated:\n%s", source)
	}
	if strings.Count(source, "Optional: true,") != 2 {
		t.Fatalf("expected two optional \"you may\" instructions:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceCourtOfVantressGatesFailClosed confirms the
// upkeep recognizer reads the typed monarch designations rather than merely two
// optional copy branches: swapping the second branch's "If you're not the
// monarch" gate for the plain "If you're the monarch" gives both branches the
// same non-negated designation, so the mutually-exclusive shape is not
// recognized, no branch is emitted, and the ability is reported unsupported.
func TestGenerateExecutableCardSourceCourtOfVantressGatesFailClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, _ := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Vantress Same Gate",
		Layout:     "normal",
		ManaCost:   "{2}{U}{U}",
		TypeLine:   "Enchantment",
		OracleText: "At the beginning of your upkeep, choose up to one other target enchantment or artifact. If you're the monarch, you may create a token that's a copy of it. If you're the monarch, you may have this enchantment become a copy of it, except it has this ability.",
	}, "t")
	if len(diagnostics) == 0 {
		t.Fatalf("expected diagnostics for non-mutually-exclusive gates, got source:\n%s", source)
	}
	if strings.Contains(source, "game.BecomeCopy{") {
		t.Fatalf("same-gate body unexpectedly lowered the become-a-copy branch:\n%s", source)
	}
}
