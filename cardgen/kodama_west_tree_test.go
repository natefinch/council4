package cardgen

import (
	"strings"
	"testing"
)

// TestLowerKodamaModifiedCombatDamageTrigger proves Kodama of the West Tree
// lowers fully. Its land-fetch trigger fires on "Whenever a modified creature
// you control deals combat damage to a player", so the "modified" adjective on
// the trigger subject must thread through the parser, compiler, and lowering to
// a MatchModified selection on the combat-damage trigger.
func TestLowerKodamaModifiedCombatDamageTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Kodama of the West Tree",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Spirit",
		ManaCost:  "{4}{G}",
		Power:     new("5"),
		Toughness: new("4"),
		OracleText: "Reach\n" +
			"Modified creatures you control have trample. (Equipment, Auras you control, and counters are modifications.)\n" +
			"Whenever a modified creature you control deals combat damage to a player, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "MatchModified: true") {
		t.Fatalf("source missing MatchModified selection:\n%s", source)
	}
}
