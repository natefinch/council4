package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceEzurisPredation covers the correlated
// group-token-creation-plus-one-to-one-fight sequence: "For each creature your
// opponents control, create a 4/4 green Phyrexian Beast creature token. Each of
// those tokens fights a different one of those creatures." (Ezuri's Predation).
// The create clause counts the opponent creatures into a token amount while
// publishing both the created tokens and the counted creatures under linked-group
// keys; the CorrelatedFight then pairs the two groups by shared position so each
// token fights a distinct counted creature.
func TestGenerateExecutableCardSourceEzurisPredation(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Ezuri's Predation",
		Layout:     "normal",
		ManaCost:   "{7}{G}{G}",
		TypeLine:   "Sorcery",
		OracleText: "For each creature your opponents control, create a 4/4 green Phyrexian Beast creature token. Each of those tokens fights a different one of those creatures.",
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.CreateToken{",
		"Kind:       game.DynamicAmountCountSelector",
		"Controller: game.ControllerOpponent",
		"PublishLinked:     game.LinkedKey(\"correlated-fight-tokens\")",
		"PublishCountGroup: game.LinkedKey(\"correlated-fight-creatures\")",
		"Primitive: game.CorrelatedFight{",
		"Subjects: \"correlated-fight-tokens\"",
		"Objects:  \"correlated-fight-creatures\"",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
