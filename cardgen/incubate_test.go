package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceExciseIncubate proves the end-to-end lowering
// of Excise the Imperfect ("Exile target nonland permanent. Its controller
// incubates X, where X is its mana value.") produces the exile-then-incubate
// ordered sequence: the target nonland permanent is exiled while publishing its
// last-known information, and its last-known controller incubates its last-known
// mana value.
func TestGenerateExecutableCardSourceExciseIncubate(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Excise the Imperfect",
		Layout:     "normal",
		ManaCost:   "{3}{W}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Exile target nonland permanent. Its controller incubates X, where X is its mana value.",
		Colors:     []string{"W"},
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Exile{",
		"Object:         game.TargetPermanentReference(0),",
		"ExileLinkedKey: game.LinkedKey(",
		"Primitive: game.Incubate{",
		"Kind:       game.DynamicAmountObjectManaValue,",
		"Object:     game.LinkedObjectReference(",
		"Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),",
		"ExcludedTypes: []types.Card{types.Land}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceStandaloneIncubate proves a standalone
// incubate keyword action ("Incubate 2.") lowers to a game.Incubate primitive
// carrying the fixed count with no recipient, so the resolving controller
// incubates.
func TestGenerateExecutableCardSourceStandaloneIncubate(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Incubator",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Incubate 2.",
		Colors:     []string{"U"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Incubate{",
		"Amount: game.Fixed(2),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "Recipient:") {
		t.Fatalf("standalone incubate should not carry a recipient:\n%s", source)
	}
}
