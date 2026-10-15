package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Lhurgoyf is the card definition for Lhurgoyf.
//
// Type: Creature — Lhurgoyf
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Lhurgoyf's power is equal to the number of creature cards in all graveyards and its toughness is equal to that number plus 1.
var Lhurgoyf = newLhurgoyf

func newLhurgoyf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lhurgoyf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCreatureCardsInAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCreatureCardsInAllGraveyards, Offset: 1}),
			OracleText: `
			Lhurgoyf's power is equal to the number of creature cards in all graveyards and its toughness is equal to that number plus 1.
		`,
		},
	}
}
