package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CerodonYearling is the card definition for Cerodon Yearling.
//
// Type: Creature — Beast
// Cost: {R}{W}
//
// Oracle text:
//
//	Vigilance, haste
var CerodonYearling = newCerodonYearling

func newCerodonYearling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Cerodon Yearling",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Vigilance, haste
		`,
		},
	}
}
