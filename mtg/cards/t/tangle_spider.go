package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TangleSpider is the card definition for Tangle Spider.
//
// Type: Creature — Spider
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Reach (This creature can block creatures with flying.)
var TangleSpider = newTangleSpider

func newTangleSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tangle Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
