package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceEntersAsCopyAddSubtype(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Synth Infiltrator",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Artifact Creature — Shapeshifter",
		OracleText: "You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact creature in addition to its other types.",
		Colors:     []string{"U"},
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersAsCopyReplacement(",
		"[]types.Sub{types.Synth}",
		"types.Artifact",
		"types.Creature",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
