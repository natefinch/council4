package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Voja, Friend to Elves
//
// Type: Token Legendary Creature — Wolf
//
// Oracle text:

// VojaFriendToElvesTokenfe640667a0154b4283a06522dad5f898 is the card definition for Voja, Friend to Elves.
var VojaFriendToElvesTokenfe640667a0154b4283a06522dad5f898 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:       "Voja, Friend to Elves",
		Colors:     []color.Color{color.Green, color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Wolf},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
	},
}
