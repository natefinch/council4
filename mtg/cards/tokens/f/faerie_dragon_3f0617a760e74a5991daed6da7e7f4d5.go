package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Faerie Dragon
//
// Type: Token Creature — Faerie Dragon
//
// Oracle text:
//   Flying

// FaerieDragonToken3f0617a760e74a5991daed6da7e7f4d5 is the card definition for Faerie Dragon.
var FaerieDragonToken3f0617a760e74a5991daed6da7e7f4d5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Faerie Dragon",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Faerie, types.Dragon},
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
