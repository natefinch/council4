package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sculpture is the card definition for Sculpture.
//
// Type: Token Artifact Creature — Sculpture
//
// Oracle text:
//   This creature's power and toughness are each equal to the number of Sculptures you control.

// SculptureToken30838baad1b049aeb231b671aa704412 is the card definition for Sculpture.
var SculptureToken30838baad1b049aeb231b671aa704412 = newSculptureToken30838baad1b049aeb231b671aa704412()

func newSculptureToken30838baad1b049aeb231b671aa704412() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:             "Sculpture",
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Sculpture},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Sub("Sculpture")}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Sub("Sculpture")}),
			OracleText: `
			This creature's power and toughness are each equal to the number of Sculptures you control.
		`,
		},
	}
}
