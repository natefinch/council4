package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestIsAutomaticManaAbility(t *testing.T) {
	tests := []struct {
		name string
		body game.ManaAbility
		want bool
	}{
		{
			name: "fixed tap output",
			body: game.TapManaAbility(mana.G),
			want: true,
		},
		{
			name: "mana choice",
			body: game.TapManaChoiceAbility(mana.G, mana.U),
		},
		{
			name: "additional rider",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
					{Primitive: game.GainLife{Amount: game.Fixed(1)}},
				}}.Ability(),
			},
		},
		{
			name: "entry choice output",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:          game.Fixed(1),
						EntryChoiceFrom: game.EntryColorChoiceKey,
					},
				}}}.Ability(),
			},
		},
		{
			name: "sacrifice source",
			body: game.ManaAbility{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C},
				}}}.Ability(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsAutomaticManaAbility(&test.body); got != test.want {
				t.Fatalf("IsAutomaticManaAbility() = %v, want %v", got, test.want)
			}
		})
	}
}
