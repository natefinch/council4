package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateDepletionTaplands verifies the five Mercadian Masques depletion
// taplands generate without diagnostics and produce the expected body-gated
// sacrifice instruction.
func TestGenerateDepletionTaplands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		mana     string
		wantMana string
	}{
		{"Hickory Woodlot", "{G}{G}", "mana.G"},
		{"Remote Farm", "{W}{W}", "mana.W"},
		{"Peat Bog", "{B}{B}", "mana.B"},
		{"Sandstone Needle", "{R}{R}", "mana.R"},
		{"Saprazzan Skerry", "{U}{U}", "mana.U"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			oracle := "This land enters tapped with two depletion counters on it.\n" +
				"{T}, Remove a depletion counter from this land: Add " + tc.mana +
				". If there are no depletion counters on this land, sacrifice it."
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracle,
			}, "x")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				for _, d := range diagnostics {
					t.Logf("diagnostic: %s — %s", d.Summary, d.Detail)
				}
				t.Fatalf("want 0 diagnostics, got %d", len(diagnostics))
			}
			for _, want := range []string{
				"counter.Depletion",
				tc.wantMana,
				"game.AddMana",
				"game.Sacrifice",
			} {
				if !strings.Contains(source, want) {
					t.Errorf("generated source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
