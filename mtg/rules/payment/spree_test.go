package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func spreeSpellCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{Cost: opt.Val(cost.Mana{cost.O(1)})},
					{Cost: opt.Val(cost.Mana{cost.O(2)})},
					{},
				},
			}),
		},
	}
}

func TestSpreeModeManaCost(t *testing.T) {
	t.Parallel()
	card := spreeSpellCard()
	for _, test := range []struct {
		name        string
		chosenModes []int
		want        int
	}{
		{name: "no modes chosen", chosenModes: nil, want: 0},
		{name: "first mode", chosenModes: []int{0}, want: 1},
		{name: "second mode", chosenModes: []int{1}, want: 2},
		{name: "both modes", chosenModes: []int{0, 1}, want: 3},
		{name: "mode without cost", chosenModes: []int{2}, want: 0},
		{name: "out-of-range ignored", chosenModes: []int{0, 9, -1}, want: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := spreeModeManaCost(card, test.chosenModes).ManaValue(); got != test.want {
				t.Fatalf("mana value = %d; want %d", got, test.want)
			}
		})
	}
}

func TestSpreeModeManaCostNilCard(t *testing.T) {
	t.Parallel()
	if got := spreeModeManaCost(nil, []int{0}); got != nil {
		t.Fatalf("cost = %+v; want nil", got)
	}
}

func TestAddSpreeModeCosts(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.R}
	options := []spellCostOption{{manaCost: &base}}
	addSpreeModeCosts(options, spreeSpellCard(), []int{0, 1})
	if options[0].manaCost == nil {
		t.Fatal("manaCost = nil; want combined cost")
	}
	if got := options[0].manaCost.ManaValue(); got != 4 {
		t.Fatalf("combined mana value = %d; want 4 ({R} + {1} + {2})", got)
	}
	// The base cost must not be mutated in place.
	if slices.Equal(base, *options[0].manaCost) {
		t.Fatal("base cost was mutated; want a fresh combined slice")
	}
}
