package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phobos
//
// Type: Token Legendary Creature — Horse
//
// Oracle text:

// PhobosTokenbbec414d1270410c8118633dde6e2a18 is the card definition for Phobos.
var PhobosTokenbbec414d1270410c8118633dde6e2a18 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "Phobos",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Horse},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	},
}
