package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Beast
//
// Type: Token Creature — Beast
//
// Oracle text:

// BeastTokencaf35bcc02af4cefa1ae2d4d235e4a79 is the card definition for Beast.
var BeastTokencaf35bcc02af4cefa1ae2d4d235e4a79 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Beast",
		Colors:    []color.Color{color.Green, color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
	},
}
