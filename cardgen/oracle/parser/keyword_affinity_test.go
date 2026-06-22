package parser

import "testing"

// TestExpandAffinityKeyword confirms the printed "Affinity for <permanents>"
// keyword expands into the canonical "This spell costs {1} less to cast for each
// <permanent> you control." static so the source-spell cost-reduction pipeline
// recognizes it, including the plural-noun singularization for card types, basic
// land types, subtypes, and the "-ies" form.
func TestExpandAffinityKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{name: "artifacts", source: "Affinity for artifacts"},
		{name: "artifacts with reminder", source: "Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)"},
		{name: "creatures", source: "Affinity for creatures"},
		{name: "enchantments", source: "Affinity for enchantments"},
		{name: "basic land type", source: "Affinity for Forests"},
		{name: "subtype", source: "Affinity for Slivers"},
		{name: "ies plural", source: "Affinity for Allies"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := sourceSpellReductionEffect(t, test.source, Context{InstantOrSorcery: true})
			if effect == nil {
				t.Fatalf("source %q did not expand to a source-spell cost reduction", test.source)
			}
			if effect.SourceSpellCostReductionAmount != 1 {
				t.Fatalf("reduction amount = %d, want 1", effect.SourceSpellCostReductionAmount)
			}
		})
	}
}

// TestExpandAffinitySingularNoun verifies the head-noun singularization that
// builds the canonical Affinity wording.
func TestExpandAffinitySingularNoun(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"artifacts":          "artifact",
		"creatures":          "creature",
		"enchantments":       "enchantment",
		"planeswalkers":      "planeswalker",
		"Forests":            "Forest",
		"Islands":            "Island",
		"Plains":             "Plains",
		"Equipment":          "Equipment",
		"Allies":             "Ally",
		"artifact creatures": "artifact creature",
		"snow lands":         "snow land",
	}
	for plural, want := range tests {
		if got := affinitySingularNoun(plural); got != want {
			t.Errorf("affinitySingularNoun(%q) = %q, want %q", plural, got, want)
		}
	}
}
