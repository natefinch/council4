package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FreedomFighterRecruit is the card definition for Freedom Fighter Recruit.
//
// Type: Creature — Human Rebel Ally
// Cost: {1}{R}
//
// Oracle text:
//
//	Freedom Fighter Recruit's power is equal to the number of creatures you control.
var FreedomFighterRecruit = newFreedomFighterRecruit

func newFreedomFighterRecruit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Freedom Fighter Recruit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:       []color.Color{color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Rebel, types.Ally},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 2}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			OracleText: `
			Freedom Fighter Recruit's power is equal to the number of creatures you control.
		`,
		},
	}
}
