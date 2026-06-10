package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Angelo
//
// Type: Token Legendary Creature — Dog
//
// Oracle text:

// AngeloToken83806bcfaeaf4cf095dda5301839662e is the card definition for Angelo.
var AngeloToken83806bcfaeaf4cf095dda5301839662e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:       "Angelo",
		Colors:     []color.Color{color.Green, color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dog},
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
	},
}
