package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnowVilliers is the card definition for Snow Villiers.
//
// Type: Legendary Creature — Human Rebel Monk
// Cost: {2}{W}
//
// Oracle text:
//
//	Vigilance
//	Snow Villiers's power is equal to the number of creatures you control.
var SnowVilliers = newSnowVilliers

func newSnowVilliers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Snow Villiers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:       []color.Color{color.White},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Rebel, types.Monk},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance
			Snow Villiers's power is equal to the number of creatures you control.
		`,
		},
	}
}
