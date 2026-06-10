package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bird Illusion
//
// Type: Token Creature — Bird Illusion
//
// Oracle text:
//   Flying

// BirdIllusionToken6c29dcc1d7b64ef6b6630cc0e4be0529 is the card definition for Bird Illusion.
var BirdIllusionToken6c29dcc1d7b64ef6b6630cc0e4be0529 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Bird Illusion",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird, types.Illusion},
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
