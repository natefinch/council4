package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrusaderOfOdric is the card definition for Crusader of Odric.
//
// Type: Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	Crusader of Odric's power and toughness are each equal to the number of creatures you control.
var CrusaderOfOdric = newCrusaderOfOdric

func newCrusaderOfOdric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Crusader of Odric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Soldier},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			OracleText: `
			Crusader of Odric's power and toughness are each equal to the number of creatures you control.
		`,
		},
	}
}
