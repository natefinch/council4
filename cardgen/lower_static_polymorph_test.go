package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourcePolymorphAura confirms the polymorph static
// shape on an Aura lowers into the four layer-faithful continuous effects:
// remove-all-abilities on the ability layer, set-colors on the color layer,
// set-types/subtypes on the type layer, and base power/toughness on the
// power/toughness-set layer, all targeting the attached object. The Aura's
// trailing reminder text is consumed so the card round-trips without diagnostics.
func TestGenerateExecutableCardSourcePolymorphAura(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Frogify",
		Layout:   "normal",
		ManaCost: "{1}{U}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"U"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature loses all abilities and is a blue Frog creature with base power and toughness 1/1. (It loses all other card types and creature types.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetColors: []color.Color{color.Blue}",
		"SetTypes:    []types.Card{types.Creature}",
		"SetSubtypes: []types.Sub{types.Frog}",
		"SetPower:     opt.Val(game.PT{Value: 1})",
		"SetToughness: opt.Val(game.PT{Value: 1})",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

// TestGenerateExecutableCardSourcePolymorphBaseOnly confirms the bare "has base
// power and toughness N/N" tail lowers into remove-all-abilities plus a base
// power/toughness set, with no color or type effect.
func TestGenerateExecutableCardSourcePolymorphBaseOnly(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Kasmina's Transmutation",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"U"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature loses all abilities and has base power and toughness 1/1.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetPower:     opt.Val(game.PT{Value: 1})",
		"SetToughness: opt.Val(game.PT{Value: 1})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "SetColors") || strings.Contains(source, "SetTypes") {
		t.Fatalf("base-only polymorph unexpectedly set colors or types:\n%s", source)
	}
}

// TestGenerateExecutableCardSourcePolymorphNameFailsClosed confirms a name-setting
// polymorph, which the continuous machinery cannot model, fails closed with the
// unsupported-static diagnostic rather than emitting a partial implementation.
func TestGenerateExecutableCardSourcePolymorphNameFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Witness Protection",
		Layout:   "normal",
		ManaCost: "{W}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"W"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature loses all abilities and is a green and white Citizen creature with base power and toughness 1/1 named Legitimate Businessperson. (It loses all other colors, card types, creature types, and names.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for name-setting polymorph")
	}
	if strings.Contains(source, "RemoveAllAbilities") {
		t.Fatalf("name-setting polymorph unexpectedly emitted RemoveAllAbilities:\n%s", source)
	}
}
