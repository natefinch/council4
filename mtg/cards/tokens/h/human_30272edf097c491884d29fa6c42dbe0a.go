package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human
//
// Type: Token Creature — Human
//
// Oracle text:

// HumanToken30272edf097c491884d29fa6c42dbe0a is the card definition for Human.
var HumanToken30272edf097c491884d29fa6c42dbe0a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Human",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
