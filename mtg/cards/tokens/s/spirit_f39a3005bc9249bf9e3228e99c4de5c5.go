package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spirit
//
// Type: Token Creature — Spirit
//
// Oracle text:
//   Flying

// SpiritTokenf39a3005bc9249bf9e3228e99c4de5c5 is the card definition for Spirit.
var SpiritTokenf39a3005bc9249bf9e3228e99c4de5c5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Spirit",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
