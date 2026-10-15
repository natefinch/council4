package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpiritCleric is the card definition for Spirit Cleric.
//
// Type: Token Creature — Spirit Cleric
//
// Oracle text:
//   This creature's power and toughness are each equal to the number of Spirits you control.

// SpiritClericToken506f77aead5c42c1aacb8db5630aedd7 is the card definition for Spirit Cleric.
var SpiritClericToken506f77aead5c42c1aacb8db5630aedd7 = newSpiritClericToken506f77aead5c42c1aacb8db5630aedd7()

func newSpiritClericToken506f77aead5c42c1aacb8db5630aedd7() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:             "Spirit Cleric",
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Spirit, types.Cleric},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Sub("Spirit")}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Sub("Spirit")}),
			OracleText: `
			This creature's power and toughness are each equal to the number of Spirits you control.
		`,
		},
	}
}
