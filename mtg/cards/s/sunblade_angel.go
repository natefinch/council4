package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SunbladeAngel is the card definition for Sunblade Angel.
//
// Type: Creature — Angel
// Cost: {5}{W}
//
// Oracle text:
//
//	Flying, first strike, vigilance, lifelink
var SunbladeAngel = newSunbladeAngel

func newSunbladeAngel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sunblade Angel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Flying, first strike, vigilance, lifelink
		`,
		},
	}
}
