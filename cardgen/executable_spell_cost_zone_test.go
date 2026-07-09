package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableNonHandSpellCostReduction exercises the static cast-cost
// discount "Spells you cast from anywhere other than your hand cost {N} less to
// cast." (Sage of the Beyond): it lowers to a controller CostModifierSpell scoped
// to the non-hand cast zones (graveyard, exile, library, and command) via the
// SourceZones set, alongside the card's other supported abilities.
func TestGenerateExecutableNonHandSpellCostReduction(t *testing.T) {
	t.Parallel()
	power, toughness := "5", "6"
	card := &ScryfallCard{
		Name:     "Sage of the Beyond",
		Layout:   "normal",
		ManaCost: "{5}{U}{U}",
		TypeLine: "Creature — Spirit Giant",
		OracleText: "Flying\n" +
			"Spells you cast from anywhere other than your hand cost {2} less to cast.\n" +
			"Foretell {4}{U} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)",
		Colors:    []string{"U"},
		Power:     &power,
		Toughness: &toughness,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "SourceZones:      []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},") {
		t.Fatalf("source missing non-hand SourceZones cost modifier:\n%s", source)
	}
	if !strings.Contains(source, "GenericReduction: 2,") {
		t.Fatalf("source missing {2} generic reduction:\n%s", source)
	}
	if !strings.Contains(source, "game.ForetellKeyword{Cost:") {
		t.Fatalf("source missing Foretell keyword:\n%s", source)
	}
}
