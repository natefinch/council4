package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cat Beast
//
// Type: Token Creature — Cat Beast
//
// Oracle text:

// CatBeastToken8050907b400744788ebd25b8b2bf85c3 is the card definition for Cat Beast.
var CatBeastToken8050907b400744788ebd25b8b2bf85c3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Cat Beast",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Cat, types.Beast},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
