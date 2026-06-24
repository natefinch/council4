package cardgen

import (
	"strings"
	"testing"
)

func theCapitolineTriadCard() *ScryfallCard {
	power, toughness := "7", "7"
	return &ScryfallCard{
		Name:     "The Capitoline Triad",
		Layout:   "normal",
		ManaCost: "{10}",
		TypeLine: "Legendary Creature — God Artificer",
		OracleText: "Those Who Came Before — This spell costs {1} less to cast for each historic card in your graveyard. (Artifacts, legendaries, and Sagas are historic.)\n" +
			"Exile any number of historic cards from your graveyard with total mana value 30 or greater: You get an emblem with \"Creatures you control have base power and toughness 9/9.\"",
		Power:     &power,
		Toughness: &toughness,
	}
}

// TestGenerateExecutableCardSourceTheCapitolineTriad asserts the activated
// ability lowers its variable-cardinality exile cost (any number of historic
// cards whose total mana value reaches a threshold) and its "You get an emblem
// with ..." effect into a CreateEmblem primitive carrying the inner static.
func TestGenerateExecutableCardSourceTheCapitolineTriad(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(theCapitolineTriadCard(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Kind:                  cost.AdditionalExile",
		"Source:                zone.Graveyard",
		"MatchHistoric:         true",
		"TotalManaValueAtLeast: 30",
		"Primitive: game.CreateEmblem{",
		"EmblemAbilities: []game.Ability{",
		"game.LayerPowerToughnessSet",
		"SetPower:     opt.Val(game.PT{Value: 9})",
		"SetToughness: opt.Val(game.PT{Value: 9})",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
