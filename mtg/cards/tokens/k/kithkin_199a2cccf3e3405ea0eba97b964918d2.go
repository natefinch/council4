package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kithkin
//
// Type: Token Creature — Kithkin
//
// Oracle text:

// KithkinToken199a2cccf3e3405ea0eba97b964918d2 is the card definition for Kithkin.
var KithkinToken199a2cccf3e3405ea0eba97b964918d2 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Kithkin",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kithkin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
