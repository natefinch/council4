package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceRemovalAuraTypeSet confirms the removal-aura
// shape "Enchanted permanent is a colorless <type> land." lowers into a
// set-colorless color effect plus a set-types/subtypes type effect on the
// attached object, without removing the permanent's abilities.
func TestGenerateExecutableCardSourceRemovalAuraTypeSet(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Song of the Dryads",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"G"},
		OracleText: "Enchant permanent\n" +
			"Enchanted permanent is a colorless Forest land.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SetColorless: true,",
		"types.Land",
		"types.Forest",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "RemoveAllAbilities: true") {
		t.Fatalf("source unexpectedly removes abilities:\n%s", source)
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceRemovalAuraLoseAbilitiesGrantMana confirms the
// full removal-aura shape "Enchanted permanent is a colorless land with
// '{T}: Add {C}' and loses all other card types and abilities." lowers into
// remove-all-abilities, set-colorless, set-types, and the granted colorless
// mana ability, all on the attached object.
func TestGenerateExecutableCardSourceRemovalAuraLoseAbilitiesGrantMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Imprisoned in the Moon",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"U"},
		OracleText: "Enchant creature, land, or planeswalker\n" +
			"Enchanted permanent is a colorless land with \"{T}: Add {C}\" and loses all other card types and abilities.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetColorless: true,",
		"types.Land",
		"game.TapManaAbility(mana.C)",
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

// TestGenerateExecutableCardSourceRemovalAuraColorlessBasePT confirms the
// loses-first removal-aura shape "Enchanted creature loses all abilities and is
// a colorless <subtype> with base power and toughness N/N." lowers into
// remove-all-abilities, set-colorless, the creature subtype, and the layer-7b
// base power/toughness, all on the attached object.
func TestGenerateExecutableCardSourceRemovalAuraColorlessBasePT(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Noggle the Mind",
		Layout:   "normal",
		ManaCost: "{1}{U}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"U"},
		OracleText: "Enchant creature\n" +
			"Enchanted creature loses all abilities and is a colorless Noggle with base power and toughness 1/1.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"RemoveAllAbilities: true,",
		"SetColorless: true,",
		"SetSubtypes: []types.Sub{types.Sub(\"Noggle\")},",
		"SetPower:",
		"SetToughness:",
		"opt.Val(game.PT{Value: 1}),",
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

// TestGenerateExecutableCardSourceRemovalAuraIsFirstBasePT confirms the is-first
// removal-aura shape "Enchanted creature is a <subtype> with base power and
// toughness N/N and loses all abilities." lowers into the creature subtype, the
// layer-7b base power/toughness, and remove-all-abilities on the attached
// object.
func TestGenerateExecutableCardSourceRemovalAuraIsFirstBasePT(t *testing.T) {
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
		"SetSubtypes: []types.Sub{types.Sub(\"Treefolk\")},",
		"SetPower:",
		"opt.Val(game.PT{Value: 0}),",
		"SetToughness:",
		"opt.Val(game.PT{Value: 4}),",
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
