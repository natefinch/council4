package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceTheSpearOfBashenga covers the "attacks the
// monarch" attack trigger paired with a "that player controls" destroy target.
// The equipped creature attacking the monarch triggers on EventAttackerDeclared
// with the defending player constrained to the monarch (Player:
// TriggerPlayerMonarch, AttackRecipient: AttackRecipientPlayer), and the destroy
// target is the tapped nonland permanent controlled by that attacked player, so
// the "that player controls" clause lowers to ControlledByDefendingPlayer rather
// than the attacker-relative ControlledByEventPlayer.
func TestGenerateExecutableCardSourceTheSpearOfBashenga(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "The Spear of Bashenga",
		Layout:   "normal",
		ManaCost: "{4}{W}",
		TypeLine: "Legendary Artifact — Equipment",
		OracleText: "When The Spear of Bashenga enters, if there is no monarch, you become the monarch.\n" +
			"Equipped creature gets +2/+2 and has vigilance.\n" +
			"Whenever equipped creature attacks the monarch, destroy target tapped nonland permanent that player controls.\n" +
			"Equip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Event:            game.EventAttackerDeclared",
		"Source:           game.TriggerSourceAttachedPermanent",
		"Player:           game.TriggerPlayerMonarch",
		"AttackRecipient:  game.AttackRecipientPlayer",
		"Primitive: game.Destroy{",
		"Object: game.TargetPermanentReference(0)",
		"ExcludedTypes: []types.Card{types.Land}",
		"Tapped: game.TriTrue",
		"ControlledByDefendingPlayer: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The destroy recipient is the attacked defending player, never the
	// attacker-relative event player.
	if strings.Contains(source, "ControlledByEventPlayer:") {
		t.Fatalf("destroy target used the attacker-relative event player:\n%s", source)
	}
}
