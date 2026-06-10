package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Avacyn
//
// Type: Token Legendary Creature — Angel
//
// Oracle text:
//   Flying, vigilance, indestructible

// AvacynToken7696b708cf5d455ea91e02854a3e3a77 is the card definition for Avacyn.
var AvacynToken7696b708cf5d455ea91e02854a3e3a77 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:       "Avacyn",
		Colors:     []color.Color{color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Angel},
		Power:      opt.Val(game.PT{Value: 8}),
		Toughness:  opt.Val(game.PT{Value: 8}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.VigilanceStaticBody,
			game.IndestructibleStaticBody,
		},
		OracleText: `
			Flying, vigilance, indestructible
		`,
	},
}
