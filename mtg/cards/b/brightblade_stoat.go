package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrightbladeStoat is the card definition for Brightblade Stoat.
//
// Type: Creature — Weasel Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	First strike, lifelink
var BrightbladeStoat = newBrightbladeStoat

func newBrightbladeStoat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Brightblade Stoat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Weasel, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			First strike, lifelink
		`,
		},
	}
}
