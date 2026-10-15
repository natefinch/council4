package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdultGoldDragon is the card definition for Adult Gold Dragon.
//
// Type: Creature — Dragon
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Flying, lifelink, haste
var AdultGoldDragon = newAdultGoldDragon

func newAdultGoldDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Adult Gold Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Flying, lifelink, haste
		`,
		},
	}
}
