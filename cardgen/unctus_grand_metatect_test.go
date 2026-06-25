package cardgen

import (
	"strings"
	"testing"
)

func unctusGrandMetatectCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Unctus, Grand Metatect",
		Layout:    "normal",
		ManaCost:  "{1}{U}{U}",
		TypeLine:  "Legendary Artifact Creature — Phyrexian Vedalken",
		Power:     new("2"),
		Toughness: new("4"),
		OracleText: "Other blue creatures you control have \"Whenever this creature becomes tapped, draw a card, then discard a card.\"\n" +
			"Other artifact creatures you control get +1/+1.\n" +
			"{U/P}: Until end of turn, target creature you control becomes a blue artifact in addition to its other colors and types. Activate only as a sorcery. ({U/P} can be paid with either {U} or 2 life.)",
	}
}

// TestGenerateExecutableCardSourceUnctusGrandMetatect asserts the legendary
// artifact creature lowers all three abilities: the continuous grant of a
// becomes-tapped loot trigger to other blue creatures, the +1/+1 anthem on other
// artifact creatures, and the Phyrexian-mana ({U/P}) sorcery-speed activated
// ability that makes a target creature you control a blue artifact in addition to
// its other colors and types until end of turn. The becomes-color-and-type
// effect lowers to an ApplyContinuous with both a LayerType artifact addition and
// a LayerColor blue addition on the single target.
func TestGenerateExecutableCardSourceUnctusGrandMetatect(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(unctusGrandMetatectCard(), "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"ManaCost:       opt.Val(cost.Mana{cost.PhyrexianMana(mana.U)}),",
		"Timing:         game.SorceryOnly,",
		"Layer:    game.LayerType,",
		"AddTypes: []types.Card{types.Artifact},",
		"Layer:     game.LayerColor,",
		"AddColors: []color.Color{color.Blue},",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
