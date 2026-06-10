package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Horror
//
// Type: Token Creature — Horror
//
// Oracle text:

// HorrorToken7d2c540e0cf145d986ab2f3490bd12cf is the card definition for Horror.
var HorrorToken7d2c540e0cf145d986ab2f3490bd12cf = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Horror",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Horror},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
