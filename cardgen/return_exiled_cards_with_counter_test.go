package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableReturnExiledCardsWithCounter proves the mass return "Put
// all exiled cards you own with intel counters on them into your hand." lowers
// to a single ReturnExiledCardsWithCounter primitive scoped to the controller
// and filtered by the named marker counter the parser read text-blind. The card
// is synthetic (not a curated card): it isolates the front-face return mechanism
// on a normal-layout permanent, complementing TestGenerateExecutableFlamewar-
// BothFaces, which exercises the same mechanism inside the full transform card.
func TestGenerateExecutableReturnExiledCardsWithCounter(t *testing.T) {
	t.Parallel()
	power := "2"
	toughness := "2"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Intel Returner",
		Layout:     "normal",
		ManaCost:   "{B}{R}",
		TypeLine:   "Artifact Creature — Robot",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.ReturnExiledCardsWithCounter{",
		"Player:  game.ControllerReference(),",
		"Counter: counter.Intel,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
