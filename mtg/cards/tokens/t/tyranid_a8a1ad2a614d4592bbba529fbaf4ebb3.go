package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tyranid
//
// Type: Token Creature — Tyranid
//
// Oracle text:

// TyranidTokena8a1ad2a614d4592bbba529fbaf4ebb3 is the card definition for Tyranid.
var TyranidTokena8a1ad2a614d4592bbba529fbaf4ebb3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Tyranid",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Tyranid},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
