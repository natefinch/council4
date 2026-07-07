package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceMBakuJabariChieftain covers the "a creature
// attacks one of your opponents, if that player is the monarch" attack trigger
// paired with a "+1/+1 and gains trample until end of turn" buff on the
// attacker. The attack triggers on EventAttackerDeclared with the defending
// player constrained to an opponent (Player: TriggerPlayerOpponent,
// AttackRecipient: AttackRecipientPlayer); the "if that player is the monarch"
// clause lowers to the EventDefendingPlayerIsMonarch intervening condition; and
// the body buffs the attacker (game.EventPermanentReference) with a
// ModifyPT(+1/+1) and an ApplyContinuous granting Trample, both until end of
// turn.
func TestGenerateExecutableCardSourceMBakuJabariChieftain(t *testing.T) {
	t.Parallel()
	power, toughness := "4", "3"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "M'Baku, Jabari Chieftain",
		Layout:    "normal",
		ManaCost:  "{1}{G}{G}",
		TypeLine:  "Legendary Creature — Human Noble Warrior",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "At the beginning of your end step, if there is no monarch, target opponent becomes the monarch.\n" +
			"Whenever a creature attacks one of your opponents, if that player is the monarch, that creature gets +1/+1 and gains trample until end of turn.",
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Attack trigger scoped to an opponent recipient.
		"Event:            game.EventAttackerDeclared",
		"Player:           game.TriggerPlayerOpponent",
		"AttackRecipient:  game.AttackRecipientPlayer",
		"SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}}",
		// "if that player is the monarch" intervening condition.
		`InterveningIf: "if that player is the monarch"`,
		"EventDefendingPlayerIsMonarch: true",
		// "that creature gets +1/+1 ... until end of turn" on the attacker.
		"Primitive: game.ModifyPT{",
		"Object:         game.EventPermanentReference()",
		"PowerDelta:     game.Fixed(1)",
		"ToughnessDelta: game.Fixed(1)",
		"Duration:       game.DurationUntilEndOfTurn",
		// "and gains trample until end of turn" on the attacker.
		"Primitive: game.ApplyContinuous{",
		"Object: opt.Val(game.EventPermanentReference())",
		"AddKeywords: []game.Keyword{",
		"game.Trample",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
