package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DakkonBlackblade is the card definition for Dakkon Blackblade.
//
// Type: Legendary Creature — Human Warrior
// Cost: {2}{W}{U}{U}{B}
//
// Oracle text:
//
//	Dakkon Blackblade's power and toughness are each equal to the number of lands you control.
var DakkonBlackblade = newDakkonBlackblade

func newDakkonBlackblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Dakkon Blackblade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors:           []color.Color{color.Black, color.Blue, color.White},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Warrior},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			OracleText: `
			Dakkon Blackblade's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
