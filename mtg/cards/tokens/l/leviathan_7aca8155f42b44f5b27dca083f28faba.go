package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Leviathan is the card definition for Leviathan.
//
// Type: Token Creature — Leviathan
//
// Oracle text:
//   Hexproof

// LeviathanToken7aca8155f42b44f5b27dca083f28faba is the card definition for Leviathan.
var LeviathanToken7aca8155f42b44f5b27dca083f28faba = newLeviathanToken7aca8155f42b44f5b27dca083f28faba()

func newLeviathanToken7aca8155f42b44f5b27dca083f28faba() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Leviathan",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			OracleText: `
			Hexproof
		`,
		},
	}
}
