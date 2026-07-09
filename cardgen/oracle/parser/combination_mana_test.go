package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

func combinationManaEffect(t *testing.T, text string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(text, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	return effects[0]
}

func TestParseFixedCombinationMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		text   string
		count  int
		colors []mana.Color
	}{
		{
			name:   "two colors and/or",
			text:   "Add three mana in any combination of {R} and/or {G}.",
			count:  3,
			colors: []mana.Color{mana.R, mana.G},
		},
		{
			name:   "all colors word",
			text:   "Add five mana in any combination of colors.",
			count:  5,
			colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
		},
		{
			name:   "three color list",
			text:   "Add two mana in any combination of {U}, {B}, and/or {R}.",
			count:  2,
			colors: []mana.Color{mana.U, mana.B, mana.R},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			effect := combinationManaEffect(t, tc.text)
			if effect.Kind != EffectAddMana || !effect.Mana.Combination ||
				effect.Mana.CombinationDynamic || effect.Mana.CombinationCount != tc.count {
				t.Fatalf("mana = %#v", effect.Mana)
			}
			if !slices.Equal(effect.Mana.CombinationColors, tc.colors) {
				t.Fatalf("colors = %v, want %v", effect.Mana.CombinationColors, tc.colors)
			}
			if !effect.Mana.LegacyBodyExact {
				t.Fatalf("fixed combination body must be legacy-exact: %#v", effect.Mana)
			}
		})
	}
}

func TestParseDynamicCombinationMana(t *testing.T) {
	t.Parallel()
	// Prismatic Geoscope's domain count is the canonical dynamic combination.
	effect := combinationManaEffect(t,
		"{T}: Add X mana in any combination of colors, where X is the number of basic land types among lands you control.")
	if effect.Kind != EffectAddMana || !effect.Exact ||
		!effect.Mana.Combination || !effect.Mana.CombinationDynamic {
		t.Fatalf("mana = %#v, amount = %#v", effect.Mana, effect.Amount)
	}
	want := []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G}
	if !slices.Equal(effect.Mana.CombinationColors, want) {
		t.Fatalf("colors = %v, want %v", effect.Mana.CombinationColors, want)
	}
	if effect.Amount.DynamicKind == EffectDynamicAmountNone {
		t.Fatalf("dynamic combination must carry a dynamic amount: %#v", effect.Amount)
	}
}

func TestParseCombinationManaFailsClosed(t *testing.T) {
	t.Parallel()
	// "that much" is not a recognized amount for add mana, a single color has no
	// combination, and colorless is never an offered combination color; each of
	// these must leave the combination shape unset so the card fails closed.
	variants := []string{
		"add that much mana in any combination of {R} and/or {G}.",
		"Add three mana in any combination of {R}.",
		"Add three mana in any combination of {R}, {G}, and/or {C}.",
		"Add one mana in any combination of {R} and/or {G}.",
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
