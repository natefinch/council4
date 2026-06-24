package parser

import "testing"

// TestRecognizeLevelUpAbility verifies a leveler card's "Level up {cost}" line is
// recognized as a level-up activated ability carrying its mana cost, only when
// the parser is given leveler context (CR 711.2).
func TestRecognizeLevelUpAbility(t *testing.T) {
	const oracle = "Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as a sorcery.)\n" +
		"LEVEL 1-2\n" +
		"0/1\n" +
		"{T}: Draw a card, then discard a card.\n" +
		"LEVEL 3+\n" +
		"0/1\n" +
		"{T}: Draw a card."
	document, diagnostics := Parse(oracle, Context{Leveler: true, CardName: "Enclave Cryptologist"})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %v", diagnostics)
	}
	levelUps := 0
	bands := 0
	for i := range document.Abilities {
		ability := document.Abilities[i]
		if ability.LevelUpRecognized {
			levelUps++
			if len(ability.LevelUpCost) == 0 {
				t.Fatalf("ability %d recognized level up with no mana cost", i)
			}
		}
		if ability.Kind == AbilityLevelBand {
			bands++
			if ability.LevelBand == nil {
				t.Fatalf("ability %d has AbilityLevelBand kind but nil LevelBand", i)
			}
		}
	}
	if levelUps != 1 {
		t.Fatalf("level-up abilities = %d, want 1", levelUps)
	}
	if bands != 2 {
		t.Fatalf("level bands = %d, want 2", bands)
	}
}

// TestRecognizeLevelBandBounds verifies the band headers parse their inclusive
// lower bound and, for closed bands, their inclusive upper bound, with the printed
// P/T captured per band (CR 711.4).
func TestRecognizeLevelBandBounds(t *testing.T) {
	const oracle = "Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as a sorcery.)\n" +
		"LEVEL 1-2\n" +
		"0/1\n" +
		"{T}: Draw a card, then discard a card.\n" +
		"LEVEL 3+\n" +
		"2/3\n" +
		"{T}: Draw a card."
	document, _ := Parse(oracle, Context{Leveler: true, CardName: "Enclave Cryptologist"})
	var bands []*LevelBand
	for i := range document.Abilities {
		if document.Abilities[i].Kind == AbilityLevelBand {
			bands = append(bands, document.Abilities[i].LevelBand)
		}
	}
	if len(bands) != 2 {
		t.Fatalf("level bands = %d, want 2", len(bands))
	}
	if bands[0].Low != 1 || bands[0].High != 2 {
		t.Fatalf("band 0 bounds = %d-%d, want 1-2", bands[0].Low, bands[0].High)
	}
	if !bands[0].HasPowerToughness || bands[0].Power != 0 || bands[0].Toughness != 1 {
		t.Fatalf("band 0 P/T = %d/%d (has=%v), want 0/1", bands[0].Power, bands[0].Toughness, bands[0].HasPowerToughness)
	}
	if bands[1].Low != 3 || bands[1].High != 0 {
		t.Fatalf("band 1 bounds = %d-%d, want 3+ (high 0)", bands[1].Low, bands[1].High)
	}
	if !bands[1].HasPowerToughness || bands[1].Power != 2 || bands[1].Toughness != 3 {
		t.Fatalf("band 1 P/T = %d/%d (has=%v), want 2/3", bands[1].Power, bands[1].Toughness, bands[1].HasPowerToughness)
	}
}

// TestRecognizeLevelUpRequiresLevelerContext verifies the "Level up" wording is
// recognized only for leveler cards, so the same body on a non-leveler card stays
// unrecognized.
func TestRecognizeLevelUpRequiresLevelerContext(t *testing.T) {
	const oracle = "Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as a sorcery.)"
	document, _ := Parse(oracle, Context{CardName: "Not A Leveler"})
	for i := range document.Abilities {
		if document.Abilities[i].LevelUpRecognized {
			t.Fatalf("ability %d recognized level up without leveler context", i)
		}
	}
}
