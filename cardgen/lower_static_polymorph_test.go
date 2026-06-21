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

// TestGenerateExecutableCardSourcePolymorphBecomesFirst confirms the
// "becomes-first" Aura order ("Enchanted creature is a <types> with base power
// and toughness N/N and has <keyword>, and it loses all other abilities, card
// types, and creature types.") lowers into the same layer-faithful continuous
// effects as the loses-first order, additionally granting the stated keyword.
// This is the Darksteel Mutation / Lignify near-vanilla family.
func TestGenerateExecutableCardSourcePolymorphBecomesFirst(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Darksteel Mutation",
		Layout:   "normal",
		ManaCost: "{1}{W}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"W"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature is an Insect artifact creature with base power and toughness 0/1 and has indestructible, and it loses all other abilities, card types, and creature types.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetTypes:    []types.Card{types.Artifact, types.Creature}",
		"SetSubtypes: []types.Sub{types.Insect}",
		"SetPower:     opt.Val(game.PT{Value: 0})",
		"SetToughness: opt.Val(game.PT{Value: 1})",
		"game.Indestructible",
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

// TestGenerateExecutableCardSourcePolymorphBecomesFirstSubtypeOnly confirms the
// becomes-first order with only a subtype and no keyword grant (Lignify) lowers
// without diagnostics, setting the subtype and base power/toughness.
func TestGenerateExecutableCardSourcePolymorphBecomesFirstSubtypeOnly(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Lignify",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"G"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature is a Treefolk with base power and toughness 0/4 and loses all abilities.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetSubtypes: []types.Sub{types.Sub(\"Treefolk\")}",
		"SetPower:     opt.Val(game.PT{Value: 0})",
		"SetToughness: opt.Val(game.PT{Value: 4})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

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
