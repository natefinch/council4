package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerGeralfNontokenSacrificeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Geralf, Visionary Stitcher",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		ManaCost:   "{2}{U}",
		Power:      new("1"),
		Toughness:  new("4"),
		OracleText: "Zombies you control have flying.\n{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.",
	})
	ability := face.ActivatedAbilities[0]
	var sacrifice *cost.Additional
	for i := range ability.AdditionalCosts {
		if ability.AdditionalCosts[i].Kind == cost.AdditionalSacrifice {
			sacrifice = &ability.AdditionalCosts[i]
		}

	}
	if sacrifice == nil ||
		!sacrifice.RequireNonToken ||
		!sacrifice.MatchPermanentType ||
		sacrifice.PermanentType != types.Creature ||
		!sacrifice.ExcludeSource {
		t.Fatalf("sacrifice cost = %#v", sacrifice)
	}
}

func TestGenerateGeralfNontokenSacrificeCost(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Geralf, Visionary Stitcher",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		ManaCost:   "{2}{U}",
		Power:      new("1"),
		Toughness:  new("4"),
		OracleText: "Zombies you control have flying.\n{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.",
	}, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "RequireNonToken:") {
		t.Fatalf("source missing nontoken cost:\n%s", source)
	}
}
