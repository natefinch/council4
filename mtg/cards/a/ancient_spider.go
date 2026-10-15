package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncientSpider is the card definition for Ancient Spider.
//
// Type: Creature — Spider
// Cost: {2}{G}{W}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	First strike
var AncientSpider = newAncientSpider

func newAncientSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Ancient Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			First strike
		`,
		},
	}
}
