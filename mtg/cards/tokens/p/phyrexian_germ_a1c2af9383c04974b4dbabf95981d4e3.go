package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Germ
//
// Type: Token Creature — Phyrexian Germ
//
// Oracle text:

// PhyrexianGermTokena1c2af9383c04974b4dbabf95981d4e3 is the card definition for Phyrexian Germ.
var PhyrexianGermTokena1c2af9383c04974b4dbabf95981d4e3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Phyrexian Germ",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Germ},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
