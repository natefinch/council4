package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceShakedownHeavy covers the defending-player
// "may have" causative gate whose affirmative consequence untaps the source and
// removes it from combat: on attack, the defending player decides whether the
// controller draws a card ("defending player may have you draw a card"), and if
// they accept ("If they do, ...") the source untaps and leaves combat. The draw
// is asked of the defending player (its OptionalActor names the defending player,
// not the controller), the offer publishes whether it happened, and both the
// self-untap and the self-remove-from-combat resolve only when accepted (TriTrue)
// against the source permanent.
func TestGenerateExecutableCardSourceShakedownHeavy(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Shakedown Heavy",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Creature — Ogre Warrior",
		OracleText: "Menace\nWhenever this creature attacks, defending player may have you draw a card. If they do, untap this creature and remove it from combat.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Draw{",
		"Player: game.ControllerReference()",
		"Optional:      true",
		"OptionalActor: opt.Val(game.DefendingPlayerReference())",
		"PublishResult: game.ResultKey(\"may-have-action\")",
		"Primitive: game.Untap{",
		"Object: game.SourcePermanentReference()",
		"Primitive: game.RemoveFromCombat{",
		"Key:       \"may-have-action\"",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
