package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerSpiritGuideHandManaAbility(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		typ  string
		text string
	}{
		{"Simian Spirit Guide", "Creature — Ape", "Exile this card from your hand: Add {R}."},
		{"Elvish Spirit Guide", "Creature — Elf Spirit", "Exile this creature from your hand: Add {G}."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   tc.typ,
				OracleText: tc.text,
			})
			if len(face.ManaAbilities) != 1 {
				t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
			}
			ability := face.ManaAbilities[0]
			if ability.ZoneOfFunction != zone.Hand {
				t.Fatalf("zone of function = %v, want hand", ability.ZoneOfFunction)
			}
			if len(ability.AdditionalCosts) != 1 ||
				ability.AdditionalCosts[0].Kind != cost.AdditionalExileSource ||
				ability.AdditionalCosts[0].Source != zone.Hand {
				t.Fatalf("additional costs = %#v, want a single exile-self-from-hand cost", ability.AdditionalCosts)
			}
			if ability.ManaCost.Exists && len(ability.ManaCost.Val) != 0 {
				t.Fatalf("mana cost = %#v, want none", ability.ManaCost)
			}
		})
	}
}

func TestGenerateExecutableCardSourceSimianSpiritGuide(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Simian Spirit Guide",
		Layout:     "normal",
		TypeLine:   "Creature — Ape",
		OracleText: "Exile this card from your hand: Add {R}.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"ManaAbilities: []game.ManaAbility{",
		"ZoneOfFunction: zone.Hand,",
		"Kind:   cost.AdditionalExileSource,",
		"Source: zone.Hand,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
