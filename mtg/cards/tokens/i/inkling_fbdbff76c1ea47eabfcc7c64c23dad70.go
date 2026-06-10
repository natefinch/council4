package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Inkling
//
// Type: Token Creature — Inkling
//
// Oracle text:
//   Flying

// InklingTokenfbdbff76c1ea47eabfcc7c64c23dad70 is the card definition for Inkling.
var InklingTokenfbdbff76c1ea47eabfcc7c64c23dad70 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:      "Inkling",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Inkling},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
