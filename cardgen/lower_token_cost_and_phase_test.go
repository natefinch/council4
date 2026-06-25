package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceSacrificeTokenCost covers a token-qualified
// sacrifice activation cost ("Sacrifice an artifact token: ..." — Sophia, Dogged
// Detective). The sacrifice cost must constrain its object to artifact tokens,
// emitting both the permanent-type filter and the RequireToken token filter.
func TestGenerateExecutableCardSourceSacrificeTokenCost(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Token Sacrificer",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact",
		OracleText: "{1}, Sacrifice an artifact token: Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:               cost.AdditionalSacrifice,",
		"Text:               \"Sacrifice an artifact token\",",
		"MatchPermanentType: true,",
		"PermanentType:      types.Artifact,",
		"RequireToken:       true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourcePhasesOutSelf covers a self-source phasing
// effect introduced by a flavor ability word ("Teleport — {3}{W}: This creature
// phases out." — Blink Dog). The effect must lower to a PhaseOut primitive that
// phases out the source permanent, and the rules-free ability word must not
// block lowering.
func TestGenerateExecutableCardSourcePhasesOutSelf(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Phasing Hound",
		Layout:   "normal",
		ManaCost: "{1}{W}",
		TypeLine: "Creature — Dog",
		OracleText: "Teleport — {3}{W}: This creature phases out. " +
			"(Treat it and anything attached to it as though they don't exist until your next turn.)",
		Power:     new("1"),
		Toughness: new("1"),
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.PhaseOut{",
		"Object: game.SourcePermanentReference(),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
