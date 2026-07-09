package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceNaturesWill covers Nature's Will: "Whenever one
// or more creatures you control deal combat damage to a player, tap all lands that
// player controls and untap all lands you control." The ordered two-effect
// sequence lowers to a mass Tap of a PlayerControlledGroup anchored on the
// triggering event's player (the combat-damage recipient) followed by a mass Untap
// of the controller's own lands. The "that player controls" land group is the
// event-player mass-group path eventPlayerControlledMassGroup adds.
func TestGenerateExecutableCardSourceNaturesWill(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Nature's Will",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{G}{G}",
		OracleText: "Whenever one or more creatures you control deal combat damage to a player, tap all lands that player controls and untap all lands you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Event:                 game.EventDamageDealt",
		"RequireCombatDamage:   true",
		"DamageRecipient:       game.DamageRecipientPlayer",
		"Primitive: game.Tap{",
		"Group: game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),",
		"Primitive: game.Untap{",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceEventPlayerMassTapSpellFailsClosed asserts the
// event-player mass-group path stays scoped to triggers whose event binds "that
// player". In a spell body there is no triggering event, so the compiler never
// binds "that player" to ReferenceBindingEventPlayer; eventPlayerControlledMassGroup
// rejects the surviving reference and the card fails closed with an unsupported
// diagnostic rather than resolving the group against a nonexistent event player.
func TestGenerateExecutableCardSourceEventPlayerMassTapSpellFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Event Player Mass Tap Probe",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Tap all lands that player controls.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected an unsupported diagnostic, got none:\n%s", source)
	}
}
