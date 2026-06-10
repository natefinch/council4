package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sphinx
//
// Type: Token Creature — Sphinx
//
// Oracle text:
//   Flying, vigilance

// SphinxToken814c8e0b694e4b65a0b6518ac005b218 is the card definition for Sphinx.
var SphinxToken814c8e0b694e4b65a0b6518ac005b218 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Blue),
	CardFace: game.CardFace{
		Name:      "Sphinx",
		Colors:    []color.Color{color.Blue, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sphinx},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.VigilanceStaticBody,
		},
		OracleText: `
			Flying, vigilance
		`,
	},
}
