package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bird
//
// Type: Token Creature — Bird
//
// Oracle text:
//   Flying

// BirdToken5c4620401ae14c9dbb2069429ee66e27 is the card definition for Bird.
var BirdToken5c4620401ae14c9dbb2069429ee66e27 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Blue),
	CardFace: game.CardFace{
		Name:      "Bird",
		Colors:    []color.Color{color.Blue, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird},
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
