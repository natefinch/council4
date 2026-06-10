package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Knight
//
// Type: Token Creature — Knight
//
// Oracle text:
//   First strike

// KnightToken100435feae7746448b4a4c941bb2067f is the card definition for Knight.
var KnightToken100435feae7746448b4a4c941bb2067f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Knight",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FirstStrikeStaticBody,
		},
		OracleText: `
			First strike
		`,
	},
}
