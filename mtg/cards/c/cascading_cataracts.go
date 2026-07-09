package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CascadingCataracts is the card definition for Cascading Cataracts.
//
// Type: Land
//
// Oracle text:
//
//	Indestructible
//	{T}: Add {C}.
//	{5}, {T}: Add five mana in any combination of colors.
var CascadingCataracts = newCascadingCataracts

func newCascadingCataracts() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Cascading Cataracts",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(5)}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(5),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Indestructible
			{T}: Add {C}.
			{5}, {T}: Add five mana in any combination of colors.
		`,
		},
	}
}
