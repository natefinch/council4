package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func escalateSpellCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			SpellAbility: opt.Val(game.AbilityContent{
				Modes:        []game.Mode{{}, {}, {}},
				MinModes:     1,
				MaxModes:     3,
				EscalateCost: opt.Val(cost.Mana{cost.G}),
			}),
		},
	}
}

func TestEscalateModeManaCost(t *testing.T) {
	t.Parallel()
	card := escalateSpellCard()
	for _, test := range []struct {
		name        string
		chosenModes []int
		want        int
	}{
		{name: "no modes chosen", chosenModes: nil, want: 0},
		{name: "one mode is free", chosenModes: []int{0}, want: 0},
		{name: "two modes pay once", chosenModes: []int{0, 1}, want: 1},
		{name: "three modes pay twice", chosenModes: []int{0, 1, 2}, want: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := escalateModeManaCost(card, test.chosenModes).ManaValue(); got != test.want {
				t.Fatalf("mana value = %d; want %d", got, test.want)
			}
		})
	}
}

func TestEscalateModeManaCostNoEscalateCost(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		SpellAbility: opt.Val(game.AbilityContent{Modes: []game.Mode{{}, {}}}),
	}}
	if got := escalateModeManaCost(card, []int{0, 1}); got != nil {
		t.Fatalf("cost = %+v; want nil", got)
	}
	if got := escalateModeManaCost(nil, []int{0, 1}); got != nil {
		t.Fatalf("cost = %+v; want nil", got)
	}
}

func TestAddEscalateModeCosts(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.R}
	options := []spellCostOption{{manaCost: &base}}
	addEscalateModeCosts(options, escalateSpellCard(), []int{0, 1, 2})
	if options[0].manaCost == nil {
		t.Fatal("manaCost = nil; want combined cost")
	}
	if got := options[0].manaCost.ManaValue(); got != 3 {
		t.Fatalf("combined mana value = %d; want 3 ({R} + {G} + {G})", got)
	}
	if slices.Equal(base, *options[0].manaCost) {
		t.Fatal("base cost was mutated; want a fresh combined slice")
	}
}
