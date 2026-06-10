package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Voja
//
// Type: Token Legendary Creature — Wolf
//
// Oracle text:

// VojaToken9903c033990949928c07687b55ff4f1c is the card definition for Voja.
var VojaToken9903c033990949928c07687b55ff4f1c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:       "Voja",
		Colors:     []color.Color{color.Green, color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Wolf},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	},
}
