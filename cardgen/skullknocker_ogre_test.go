package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceSkullknockerOgre pins the non-controller
// resolving-success gate: a trigger body whose consequence is gated on whether a
// preceding action taken by a non-controller player resolved. Skullknocker
// Ogre's damaged opponent ("that player") discards at random, and the draw runs
// only "if the player does" discard.
//
// It exercises two generalizations working together:
//
//   - The parser recognizes the affirmative non-controller resolving-success
//     gate wording "if the player does" (alongside "if they do"/"if that player
//     does") and maps it to the same prior-instruction predicate as the
//     controller "if you do", so the mandatory-if-you-do flow gates the draw on
//     the discard's published success regardless of which player discarded.
//   - The compiler binds the subject player pronoun "they" ("they draw a card")
//     to the event player, matching the explicit "that player" demonstrative, so
//     the drawer is the damaged opponent rather than the triggering permanent.
//
// The discard publishes its result and the draw carries a success gate on it, so
// the "if the player does" linkage is preserved and fails closed if the discard
// does not happen.
func TestGenerateExecutableCardSourceSkullknockerOgre(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Skullknocker Ogre",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Creature — Ogre",
		OracleText: "Whenever this creature deals damage to an opponent, that player discards a card at random. If the player does, they draw a card.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventDamageDealt",
		"DamageRecipient: game.DamageRecipientPlayer",
		"Primitive: game.Discard",
		"AtRandom: true",
		"Player:   game.EventPlayerReference()",
		"PublishResult: game.ResultKey(\"if-you-do\")",
		"Primitive: game.Draw",
		"ResultGate: opt.Val(game.InstructionResultGate{",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The discard-drawer is the damaged opponent (the event player), not the
	// controller: a stray controller draw would mean the "they" pronoun bound to
	// the wrong player.
	if strings.Contains(source, "game.ControllerReference()") {
		t.Fatalf("draw recipient bound to controller, want event player:\n%s", source)
	}
}
