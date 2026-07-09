package cardgen

import (
	"strings"
	"testing"
)

func sarythViperFangCard() *ScryfallCard {
	power, toughness := "3", "4"
	return &ScryfallCard{
		Name:      "Saryth, the Viper's Fang",
		Layout:    "normal",
		ManaCost:  "{2}{G}{G}",
		TypeLine:  "Legendary Creature — Human Warlock",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Other tapped creatures you control have deathtouch.\n" +
			"Other untapped creatures you control have hexproof.\n" +
			"{1}, {T}: Untap another target creature or land you control.",
	}
}

// TestGenerateExecutableCardSourceSarythViperFang asserts the two
// tap-state-scoped static grants lower onto source-excluding controlled-creature
// groups that carry the tapped (deathtouch) and untapped (hexproof) tap states —
// the "Other untapped creatures you control" affected group this branch unlocks
// as the untapped sibling of the existing tapped form — alongside the supported
// "{1}, {T}: Untap another target ..." activated ability.
func TestGenerateExecutableCardSourceSarythViperFang(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(sarythViperFangCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}, game.SourcePermanentReference())",
		"game.Deathtouch",
		"game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}, game.SourcePermanentReference())",
		"game.Hexproof",
		"game.Untap{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
