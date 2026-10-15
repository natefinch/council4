package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SimaYiWeiFieldMarshal is the card definition for Sima Yi, Wei Field Marshal.
//
// Type: Legendary Creature — Human Soldier
// Cost: {5}{B}
//
// Oracle text:
//
//	Sima Yi's power is equal to the number of Swamps you control.
var SimaYiWeiFieldMarshal = newSimaYiWeiFieldMarshal

func newSimaYiWeiFieldMarshal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sima Yi, Wei Field Marshal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:       []color.Color{color.Black},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Soldier},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			OracleText: `
			Sima Yi's power is equal to the number of Swamps you control.
		`,
		},
	}
}
