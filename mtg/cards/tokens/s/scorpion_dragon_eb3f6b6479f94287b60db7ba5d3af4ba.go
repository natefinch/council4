package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Scorpion Dragon
//
// Type: Token Creature — Scorpion Dragon
//
// Oracle text:
//   Flying, haste

// ScorpionDragonTokeneb3f6b6479f94287b60db7ba5d3af4ba is the card definition for Scorpion Dragon.
var ScorpionDragonTokeneb3f6b6479f94287b60db7ba5d3af4ba = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Scorpion Dragon",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Scorpion, types.Dragon},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Flying, haste
		`,
	},
}
