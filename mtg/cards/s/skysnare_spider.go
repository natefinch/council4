package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkysnareSpider is the card definition for Skysnare Spider.
//
// Type: Creature — Spider
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Reach (This creature can block creatures with flying.)
var SkysnareSpider = newSkysnareSpider

func newSkysnareSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Skysnare Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Vigilance (Attacking doesn't cause this creature to tap.)
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
