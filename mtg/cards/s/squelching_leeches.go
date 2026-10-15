package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SquelchingLeeches is the card definition for Squelching Leeches.
//
// Type: Creature — Leech
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Squelching Leeches's power and toughness are each equal to the number of Swamps you control.
var SquelchingLeeches = newSquelchingLeeches

func newSquelchingLeeches() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Squelching Leeches",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:           []color.Color{color.Black},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Leech},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			OracleText: `
			Squelching Leeches's power and toughness are each equal to the number of Swamps you control.
		`,
		},
	}
}
