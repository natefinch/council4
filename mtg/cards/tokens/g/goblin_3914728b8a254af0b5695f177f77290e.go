package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin
//
// Type: Token Creature — Goblin
//
// Oracle text:

// GoblinToken3914728b8a254af0b5695f177f77290e is the card definition for Goblin.
var GoblinToken3914728b8a254af0b5695f177f77290e = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Goblin",
		Colors:    []color.Color{color.Black, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
