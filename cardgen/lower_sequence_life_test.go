package cardgen

import (
	"strings"
	"testing"
)

// TestLowerSequenceGainLifeEqualToToughness proves the toughness sibling of the
// "gain life equal to its power" source/event amount lowers inside an ordered
// effect sequence: Angelic Chorus's "Whenever a creature you control enters, you
// gain life equal to its toughness." binds the amount to the entering creature's
// toughness through an event-permanent reference.
func TestLowerSequenceGainLifeEqualToToughness(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Angelic Chorus",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{3}{W}{W}",
		OracleText: "Whenever a creature you control enters, you gain life equal to its toughness.",
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.GainLife{",
		"game.DynamicAmountObjectToughness",
		"game.EventPermanentReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerSequenceThatPlayerLosesLife proves a "That player loses N life."
// clause whose subject is the controller of the preceding clause's permanent
// target lowers to a LoseLife on that target's controller. Dalek Drone's
// "destroy target creature an opponent controls. That player loses 3 life."
// loses three life from the destroyed creature's controller.
func TestLowerSequenceThatPlayerLosesLife(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Dalek Drone",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Dalek",
		ManaCost:  "{3}{B}",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "Flying, menace\n" +
			"Exterminate! — When this creature enters, destroy target creature an opponent controls. That player loses 3 life.",
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.LoseLife{",
		"game.Fixed(3)",
		"game.ObjectControllerReference(game.TargetPermanentReference(0))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
