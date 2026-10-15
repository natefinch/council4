package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tarmogoyf is the card definition for Tarmogoyf.
//
// Type: Token Creature — Lhurgoyf
// Cost: {1}{G}
//
// Oracle text:
//   Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.
//   (This token's mana cost is {1}{G}.)

// TarmogoyfToken3d7d609ae44747579a0c28f9787eab38 is the card definition for Tarmogoyf.
var TarmogoyfToken3d7d609ae44747579a0c28f9787eab38 = newTarmogoyfToken3d7d609ae44747579a0c28f9787eab38()

func newTarmogoyfToken3d7d609ae44747579a0c28f9787eab38() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tarmogoyf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards, Offset: 1}),
			OracleText: `
			Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.
			(This token's mana cost is {1}{G}.)
		`,
		},
	}
}
