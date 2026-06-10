package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Darkstar
//
// Type: Token Legendary Creature — Dog
//
// Oracle text:

// DarkstarTokenff97c30b196d494298a079c53812e9af is the card definition for Darkstar.
var DarkstarTokenff97c30b196d494298a079c53812e9af = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:       "Darkstar",
		Colors:     []color.Color{color.Black, color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dog},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	},
}
