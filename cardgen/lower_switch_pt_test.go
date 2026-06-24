package cardgen

import "testing"

// TestLowerSwitchPTSourceActivated covers the discard-cost activated source form
// (Aeromoeba, Aquamoeba), which lowers to an ApplyContinuous at the
// power/toughness switch layer applied to the source until end of turn.
func TestLowerSwitchPTSourceActivated(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Aeromoeba",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental Beast",
		ManaCost:   "{3}{U}",
		OracleText: "Flying\nDiscard a card: Switch this creature's power and toughness until end of turn.",
	}, "a")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"game.LayerPowerToughnessSwitch,",
		"Duration: game.DurationUntilEndOfTurn,",
		"cost.AdditionalDiscard,",
	)
}

// TestLowerSwitchPTSourceMana covers the bare mana-cost activated source form
// (Crag Puca, Windreaver).
func TestLowerSwitchPTSourceMana(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Windreaver",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental Spirit",
		ManaCost:   "{3}{U}",
		OracleText: "{U}: Switch this creature's power and toughness until end of turn.",
	}, "w")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"game.LayerPowerToughnessSwitch,",
		"Duration: game.DurationUntilEndOfTurn,",
	)
}
