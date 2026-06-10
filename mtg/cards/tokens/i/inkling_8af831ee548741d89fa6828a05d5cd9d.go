package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Inkling
//
// Type: Token Creature Inkling
//
// Oracle text:
//   Flying

// InklingToken8af831ee548741d89fa6828a05d5cd9d is the card definition for Inkling.
var InklingToken8af831ee548741d89fa6828a05d5cd9d = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Inkling",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
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
