package cardgen

import (
	"strings"
	"testing"
)

// TestLowerSequenceThatPlayerDiscardsPlayerTarget proves a "That player
// discards N cards." clause whose subject is the player targeted by the
// preceding clause lowers to a Discard on that target player. Ozai's Cruelty's
// "Ozai's Cruelty deals 2 damage to target player. That player discards two
// cards." discards two from the damaged player.
func TestLowerSequenceThatPlayerDiscardsPlayerTarget(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Ozai's Cruelty",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{B}",
		OracleText: "Ozai's Cruelty deals 2 damage to target player. That player discards two cards.",
	}, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Discard{",
		"game.Fixed(2)",
		"game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerSequenceThatPlayerDiscardsPermanentOwner proves a "that player
// discards N cards." clause whose subject is the controller of the preceding
// clause's permanent target lowers to a Discard on that target's controller.
// Recoil's "Return target permanent to its owner's hand. Then that player
// discards a card." discards one card from the returned permanent's controller.
func TestLowerSequenceThatPlayerDiscardsPermanentOwner(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Recoil",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}{B}",
		OracleText: "Return target permanent to its owner's hand. Then that player discards a card.",
	}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Discard{",
		"game.Fixed(1)",
		"game.ObjectControllerReference(game.TargetPermanentReference(0))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerSequenceThatPlayerMills proves the shared fixed-card-count player
// lowering also resolves a "That player mills N cards." clause whose subject is
// the player targeted by the preceding clause, emitting a Mill on that target.
func TestLowerSequenceThatPlayerMills(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mind Sear",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{U}",
		OracleText: "Mind Sear deals 2 damage to target player. That player mills two cards.",
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Mill{",
		"game.Fixed(2)",
		"game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
