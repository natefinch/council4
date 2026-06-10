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

// BeastToken9dba1bc2a76d412cbc0ef7b8eac800f4 is the card definition for Beast.
var BeastToken9dba1bc2a76d412cbc0ef7b8eac800f4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
