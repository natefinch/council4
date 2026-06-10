package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wolf
//
// Type: Token Creature — Wolf
//
// Oracle text:
//   Deathtouch

// WolfToken1aca703e70834baba8accb024a8404a1 is the card definition for Wolf.
var WolfToken1aca703e70834baba8accb024a8404a1 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Wolf",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
