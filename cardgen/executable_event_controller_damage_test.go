package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceEventPermanentControllerDamage covers the
// triggered "deals N damage to that <object>'s controller" family, whose
// recipient is the controller of the permanent that fired the trigger. The
// damage subject is the ability's own source permanent and the amount is fixed.
func TestGenerateExecutableCardSourceEventPermanentControllerDamage(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		typeLine string
		oracle   string
	}{
		// "Whenever a land enters, ~ deals 2 damage to that land's controller."
		// (Ankh of Mishra) is the enters-trigger member of the family.
		{"Test Ankh", "Artifact", "Whenever a land enters, this artifact deals 2 damage to that land's controller."},
		// "Whenever a creature dies, ~ deals 2 damage to that creature's
		// controller." (Dingus Staff) fires on a death trigger.
		{"Test Death Toll", "Artifact", "Whenever a creature dies, this artifact deals 2 damage to that creature's controller."},
		// "Whenever a creature blocks, ~ deals 1 damage to that creature's
		// controller." (Heat of Battle) fires on a combat trigger.
		{"Test Block Sting", "Enchantment", "Whenever a creature blocks, this enchantment deals 1 damage to that creature's controller."},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       tc.name,
			Layout:     "normal",
			ManaCost:   "{2}",
			TypeLine:   tc.typeLine,
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: game.Damage{",
			"Recipient:    game.PlayerDamageRecipient(game.ObjectControllerReference(game.EventPermanentReference()))",
			"DamageSource: opt.Val(game.SourcePermanentReference())",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

// TestGenerateExecutableCardSourceEventPermanentControllerDamageRejections keeps
// the controller-recipient path fail-closed: an unsupported trailing condition
// leaves the effect to the standard damage paths, which emit their own
// diagnostics rather than silently dropping the rider.
func TestGenerateExecutableCardSourceEventPermanentControllerDamageRejections(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// "unless that player pays" introduces an unsupported condition.
		"Whenever a creature blocks, this enchantment deals 3 damage to that creature's controller unless that player pays {3}.",
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			ManaCost:   "{2}{R}",
			TypeLine:   "Creature — Beast",
			OracleText: oracle,
			Power:      new("2"),
			Toughness:  new("2"),
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("%q: expected diagnostics, got none", oracle)
		}
	}
}
