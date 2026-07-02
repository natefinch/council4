package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceLevelUp verifies the Level Up slice (CR 711):
// the "Level up {cost}" line lowers to a sorcery-timed activated ability that
// puts a level counter on the source, each "LEVEL lo-hi"/"LEVEL lo+" band sets
// the source's base power and toughness while its level-counter count is in
// band, and abilities printed under a band are gated to that band's levels.
func TestGenerateExecutableCardSourceLevelUp(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Enclave Cryptologist",
		Layout:    "leveler",
		ManaCost:  "{U}",
		TypeLine:  "Creature — Merfolk Wizard",
		Power:     new("0"),
		Toughness: new("1"),
		OracleText: "Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as a sorcery.)\n" +
			"LEVEL 1-2\n" +
			"0/1\n" +
			"{T}: Draw a card, then discard a card.\n" +
			"LEVEL 3+\n" +
			"0/1\n" +
			"{T}: Draw a card.",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.AddCounter{",
		"counter.Level",
		"game.SorceryOnly",
		"SourceLevelCountersAtLeast:  1",
		"SourceLevelCountersLessThan: 3",
		"SourceLevelCountersAtLeast: 3",
		"game.LayerPowerToughnessSet",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceLevelBandKeywords(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Student of Warfare",
		Layout:    "leveler",
		ManaCost:  "{W}",
		TypeLine:  "Creature — Human Knight",
		Power:     new("1"),
		Toughness: new("1"),
		OracleText: "Level up {W} ({W}: Put a level counter on this. Level up only as a sorcery.)\n" +
			"LEVEL 2-6\n" +
			"3/3\n" +
			"First strike\n" +
			"LEVEL 7+\n" +
			"4/4\n" +
			"Double strike",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.SimpleKeyword{Kind: game.FirstStrike}",
		"game.SimpleKeyword{Kind: game.DoubleStrike}",
		"SourceLevelCountersAtLeast:  2",
		"SourceLevelCountersLessThan: 7",
		"SourceLevelCountersAtLeast: 7",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "game.FirstStrikeStaticBody") ||
		strings.Contains(source, "game.DoubleStrikeStaticBody") {
		t.Fatalf("conditioned keyword bands must render as literals:\n%s", source)
	}
}
