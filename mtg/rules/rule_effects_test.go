package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

func TestSpellCostModifierColorDisjunctionMatches(t *testing.T) {
	t.Parallel()
	modifier := game.CostModifier{
		Kind:             game.CostModifierSpell,
		GenericReduction: 1,
		MatchColors:      []color.Color{color.Red, color.Green},
	}
	tests := map[string]struct {
		colors []color.Color
		want   bool
	}{
		"red spell":            {colors: []color.Color{color.Red}, want: true},
		"green spell":          {colors: []color.Color{color.Green}, want: true},
		"red green spell":      {colors: []color.Color{color.Red, color.Green}, want: true},
		"white red spell":      {colors: []color.Color{color.White, color.Red}, want: true},
		"blue spell":           {colors: []color.Color{color.Blue}, want: false},
		"white black spell":    {colors: []color.Color{color.White, color.Black}, want: false},
		"colorless spell fail": {colors: nil, want: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &game.CardDef{CardFace: game.CardFace{Colors: test.colors}}
			if got := spellCostModifierMatchesCard(modifier, card); got != test.want {
				t.Fatalf("match %v = %v, want %v", test.colors, got, test.want)
			}
		})
	}
	if spellCostModifierMatchesCard(modifier, nil) {
		t.Fatal("nil card matched a color-disjunction modifier, want no match")
	}
}
