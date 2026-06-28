package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExileAttached confirms the attached-recipient exile form "Exile
// enchanted creature." (Aura) / "Exile equipped creature." (Equipment) lowers
// its exiled object to the runtime's source attached-permanent reference, the
// same machinery the attached-recipient regenerate path uses.
func TestGenerateExileAttached(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		text     string
	}{
		{
			name:     "Cooped Up",
			typeLine: "Enchantment — Aura",
			text:     "Enchant creature\nEnchanted creature can't attack or block.\n{2}{W}: Exile enchanted creature.",
		},
		{
			name:     "Dreadful Apathy",
			typeLine: "Enchantment — Aura",
			text:     "Enchant creature\nEnchanted creature can't attack or block.\n{2}{W}: Exile enchanted creature.",
		},
		{
			name:     "Banish Blade",
			typeLine: "Artifact — Equipment",
			text:     "{2}: Exile equipped creature.\nEquip {3}",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			card := &ScryfallCard{Name: c.name, Layout: "normal", TypeLine: c.typeLine, OracleText: c.text}
			source, diags, err := GenerateExecutableCardSource(card, "z")
			if err != nil {
				t.Fatal(err)
			}
			if len(diags) != 0 {
				t.Fatalf("diags = %#v", diags)
			}
			if !strings.Contains(source, "game.Exile{") {
				t.Fatalf("missing Exile primitive:\n%s", source)
			}
			if !strings.Contains(source, "game.SourceAttachedPermanentReference()") {
				t.Fatalf("missing source attached-permanent reference:\n%s", source)
			}
		})
	}
}

// TestGenerateExileAttachedSacrificeCost confirms the "{cost}, Sacrifice this
// Aura: Exile enchanted creature." form (Choking Restraints) lowers, exercising
// the attached exile alongside an activation sacrifice cost.
func TestGenerateExileAttachedSacrificeCost(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Choking Restraints",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature can't attack or block.\n{3}{W}{W}, Sacrifice this Aura: Exile enchanted creature.",
	}
	source, diags, err := GenerateExecutableCardSource(card, "z")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("diags = %#v", diags)
	}
	if !strings.Contains(source, "game.Exile{") ||
		!strings.Contains(source, "game.SourceAttachedPermanentReference()") {
		t.Fatalf("missing attached exile lowering:\n%s", source)
	}
}

// TestExileAttachedFailsClosed documents that exile shapes which are not the
// bare attached-creature recipient keep the clause unsupported so lowering fails
// closed rather than mislowering to the attached-permanent reference.
func TestExileAttachedFailsClosed(t *testing.T) {
	for _, text := range []string{
		"Exile target creature.",
		"Exile enchanted creature and all Auras attached to it.",
		"Exile enchanted permanent.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Probe", Layout: "normal", TypeLine: "Instant", OracleText: text}
			source, _, err := GenerateExecutableCardSource(card, "z")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(source, "game.Exile{") &&
				strings.Contains(source, "game.SourceAttachedPermanentReference()") {
				t.Fatalf("clause %q mislowered to attached exile:\n%s", text, source)
			}
		})
	}
}
