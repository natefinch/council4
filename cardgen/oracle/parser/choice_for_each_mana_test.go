package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

// TestParseChoiceForEachMana proves that an "Add <color> or <color> ... for each
// <count>" body is typed as a freely-split combination of the listed colors whose
// amount is the "for each" dynamic count (Culling Ritual's destroyed-this-way
// payoff, and the general per-count color choice).
func TestParseChoiceForEachMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		text   string
		colors []mana.Color
	}{
		{
			name:   "destroyed this way two color choice",
			text:   "Add {B} or {G} for each permanent destroyed this way.",
			colors: []mana.Color{mana.B, mana.G},
		},
		{
			name:   "controlled count three color choice",
			text:   "Add {W}, {U}, or {B} for each creature you control.",
			colors: []mana.Color{mana.W, mana.U, mana.B},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			effect := combinationManaEffect(t, tc.text)
			if effect.Kind != EffectAddMana || !effect.Exact ||
				!effect.Mana.Combination || !effect.Mana.CombinationDynamic {
				t.Fatalf("mana = %#v, amount = %#v", effect.Mana, effect.Amount)
			}
			if !slices.Equal(effect.Mana.CombinationColors, tc.colors) {
				t.Fatalf("colors = %v, want %v", effect.Mana.CombinationColors, tc.colors)
			}
			if effect.Amount.DynamicForm != EffectDynamicAmountFormForEach {
				t.Fatalf("amount must be a for-each dynamic: %#v", effect.Amount)
			}
		})
	}
}

// TestParseChoiceForEachManaFailsClosed keeps shapes the combination model does
// not cover unset: a single produced color is a plain scaled output rather than a
// choice, a colorless option is never an offered combination color, and a fixed
// choice with no "for each" count stays a one-mana choice.
func TestParseChoiceForEachManaFailsClosed(t *testing.T) {
	t.Parallel()
	variants := []string{
		"Add {G} for each creature you control.",
		"Add {B} or {C} for each permanent destroyed this way.",
		"Add {B} and {G} for each permanent destroyed this way.",
		"Add {B} or {G}.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) == 0 || len(document.Abilities[0].Sentences) == 0 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			continue
		}
		if effects[0].Mana.Combination {
			t.Fatalf("variant unexpectedly recognized a combination body:\n%s", source)
		}
	}
}
