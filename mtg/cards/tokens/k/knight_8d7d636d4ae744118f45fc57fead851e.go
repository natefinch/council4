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
//   Vigilance

// KnightToken8d7d636d4ae744118f45fc57fead851e is the card definition for Knight.
var KnightToken8d7d636d4ae744118f45fc57fead851e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Blue),
	CardFace: game.CardFace{
		Name:      "Knight",
		Colors:    []color.Color{color.Blue, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
