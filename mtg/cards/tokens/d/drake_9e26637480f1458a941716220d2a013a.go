package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Drake
//
// Type: Token Creature — Drake
//
// Oracle text:
//   Flying

// DrakeToken9e26637480f1458a941716220d2a013a is the card definition for Drake.
var DrakeToken9e26637480f1458a941716220d2a013a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Drake",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Drake},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
