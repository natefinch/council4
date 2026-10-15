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
//   Hexproof (This token can't be the target of spells or abilities your opponents control.)

// LeviathanTokenea3afc88eb2c451eb2c4d8b1fe3dd428 is the card definition for Leviathan.
var LeviathanTokenea3afc88eb2c451eb2c4d8b1fe3dd428 = newLeviathanTokenea3afc88eb2c451eb2c4d8b1fe3dd428()

func newLeviathanTokenea3afc88eb2c451eb2c4d8b1fe3dd428() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Leviathan",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			OracleText: `
			Hexproof (This token can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
