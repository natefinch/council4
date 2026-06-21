package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateRegenerateRecipients(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		text     string
		wantRef  string
	}{
		{"Self This Creature", "Creature — Beast", "{G}: Regenerate this creature.", "game.SourcePermanentReference()"},
		{"Vorthos Troll", "Legendary Creature — Troll", "{G}: Regenerate Vorthos Troll.", "game.SourcePermanentReference()"},
		{"Mantle Cloak", "Enchantment — Aura", "Enchant creature\n{1}: Regenerate enchanted creature.", "game.SourceAttachedPermanentReference()"},
		{"Plate Guard", "Artifact — Equipment", "{2}: Regenerate equipped creature.\nEquip {3}", "game.SourceAttachedPermanentReference()"},
		{"Ward Spell", "Instant", "Regenerate target creature.", "game.TargetPermanentReference(0)"},
		{"Counter Troll", "Creature — Troll", "{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.", "game.SourcePermanentReference()"},
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
			if !strings.Contains(source, "game.Regenerate{") {
				t.Fatalf("missing Regenerate primitive:\n%s", source)
			}
			if !strings.Contains(source, c.wantRef) {
				t.Fatalf("missing object ref %q:\n%s", c.wantRef, source)
			}
		})
	}
}
