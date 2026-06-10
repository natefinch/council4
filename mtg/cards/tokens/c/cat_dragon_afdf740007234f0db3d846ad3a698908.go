package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cat Dragon
//
// Type: Token Creature — Cat Dragon
//
// Oracle text:
//   Flying

// CatDragonTokenafdf740007234f0db3d846ad3a698908 is the card definition for Cat Dragon.
var CatDragonTokenafdf740007234f0db3d846ad3a698908 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Cat Dragon",
		Colors:    []color.Color{color.Black, color.Green, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Cat, types.Dragon},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
