package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZetalpaPrimalDawn is the card definition for Zetalpa, Primal Dawn.
//
// Type: Legendary Creature — Elder Dinosaur
// Cost: {6}{W}{W}
//
// Oracle text:
//
//	Flying, double strike, vigilance, trample, indestructible
var ZetalpaPrimalDawn = newZetalpaPrimalDawn

func newZetalpaPrimalDawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Zetalpa, Primal Dawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elder, types.Dinosaur},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DoubleStrikeStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Flying, double strike, vigilance, trample, indestructible
		`,
		},
	}
}
