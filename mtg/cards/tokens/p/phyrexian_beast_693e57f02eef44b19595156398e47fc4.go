package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Beast
//
// Type: Token Creature — Phyrexian Beast
//
// Oracle text:

// PhyrexianBeastToken693e57f02eef44b19595156398e47fc4 is the card definition for Phyrexian Beast.
var PhyrexianBeastToken693e57f02eef44b19595156398e47fc4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Beast},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
