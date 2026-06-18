package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceOptionalHaveSelfDamage covers the controller
// "you may have this creature deal ..." causative: the trigger body lowers to a
// single Damage instruction marked Optional.
func TestGenerateExecutableCardSourceOptionalHaveSelfDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Searing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, you may have this creature deal 1 damage to each creature.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"Primitive: game.Damage",
		"Optional: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceOptionalHaveItDamage covers the controller
// "you may have it deal ..." causative whose subject is the referenced object
// (the dying/blocking source). Both the structural "have" and the real Damage
// effect compile optional in this pronoun form; the lowerer still produces a
// single Optional Damage instruction sourced from the event permanent.
func TestGenerateExecutableCardSourceOptionalHaveItDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Spiteful Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, you may have it deal 1 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Damage",
		"DamageSource: opt.Val(game.EventPermanentReference())",
		"Optional: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceRejectsUnsupportedOptionalHave keeps the
// fail-closed boundaries: non-controller "<player> may have" optionals and
// "have <subject> <unsupported action>" bodies stay unsupported rather than
// silently dropping the player gate or the unsupported inner effect.
func TestGenerateExecutableCardSourceRejectsUnsupportedOptionalHave(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		power      *string
	}{
		{
			name:       "non-controller controller-may",
			typeLine:   "Enchantment",
			oracleText: "Whenever a creature enters, that creature's controller may have it deal damage equal to its power to any target.",
		},
		{
			name:       "unsupported inner discard each opponent",
			typeLine:   "Creature — Bear",
			oracleText: "When this creature enters, you may have each opponent discard a card.",
			power:      new("2"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsupported Have",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      test.power,
				Toughness:  test.power,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported optional effect" {
				t.Fatalf("diagnostics = %#v, want unsupported optional effect", diagnostics)
			}
		})
	}
}
