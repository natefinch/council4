package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableDescentOfTheDragons exercises the destroy form of the
// variable-target removal-token family (issue #1925): "Destroy any number of
// target creatures. For each creature destroyed this way, its controller creates
// a 4/4 red Dragon creature token with flying." The spell announces one
// variable-count creature target spec and removes every chosen target under a
// source-keyed link, paired with a per-controller token payoff that reads the
// same key.
func TestGenerateExecutableDescentOfTheDragons(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Descent of the Dragons",
		Layout:   "normal",
		ManaCost: "{4}{R}{R}",
		TypeLine: "Sorcery",
		OracleText: "Destroy any number of target creatures. For each creature destroyed this way, " +
			"its controller creates a 4/4 red Dragon creature token with flying.",
		Colors: []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"MinTargets: 0,",
		"MaxTargets: 99,",
		"Primitive: game.RemoveTargetsForToken{",
		`LinkedKey: game.LinkedKey("removed-targets-for-token"),`,
		"Primitive: game.CreateTokenForEachDestroyed{",
		"Source:    game.TokenDef(",
		"game.FlyingStaticBody,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "Exile: true,") {
		t.Fatalf("destroy form must not set Exile:\n%s", source)
	}
}

// TestGenerateExecutableCurseOfTheSwine exercises the exile + X-target form of
// the variable-target removal-token family (issue #1925): "Exile X target
// creatures. For each creature exiled this way, its controller creates a 2/2
// green Boar creature token." The X target spec binds its target count to the
// spell's chosen X via CountEqualsX, and removal uses the exile form.
func TestGenerateExecutableCurseOfTheSwine(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Curse of the Swine",
		Layout:   "normal",
		ManaCost: "{X}{U}{U}",
		TypeLine: "Sorcery",
		OracleText: "Exile X target creatures. For each creature exiled this way, " +
			"its controller creates a 2/2 green Boar creature token.",
		Colors: []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"MinTargets:   0,",
		"MaxTargets:   20,",
		"CountEqualsX: true,",
		"Primitive: game.RemoveTargetsForToken{",
		"Exile:     true,",
		`LinkedKey: game.LinkedKey("removed-targets-for-token"),`,
		"Primitive: game.CreateTokenForEachDestroyed{",
		"Source:    game.TokenDef(",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
