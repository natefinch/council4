package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceHagPunisher(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Hag of Ceaseless Torment",
		Layout:     "normal",
		ManaCost:   "{2}{B}{B}",
		TypeLine:   "Enchantment Creature — Hag",
		OracleText: "At the beginning of your upkeep, each opponent loses 3 life unless that player sacrifices a nonland permanent or discards a card.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.PunisherEachLoseLife{",
		"game.OpponentsReference()",
		"game.Fixed(3)",
		"AllowSacrifice:",
		"AllowDiscard:",
		"game.Selection{ExcludedTypes: []types.Card{types.Land}}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourcePunisherDiscardOnly(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Punisher Discard Only",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Each opponent loses 2 life unless that player discards a card.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.PunisherEachLoseLife{",
		"game.Fixed(2)",
		"AllowDiscard:",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "AllowSacrifice") {
		t.Fatalf("discard-only punisher should not set AllowSacrifice:\n%s", source)
	}
}
