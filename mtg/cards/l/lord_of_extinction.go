package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LordOfExtinction is the card definition for Lord of Extinction.
//
// Type: Creature — Elemental
// Cost: {3}{B}{G}
//
// Oracle text:
//
//	Lord of Extinction's power and toughness are each equal to the number of cards in all graveyards.
var LordOfExtinction = newLordOfExtinction

func newLordOfExtinction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Lord of Extinction",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.G,
			}),
			Colors:           []color.Color{color.Black, color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elemental},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueAllGraveyardsSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueAllGraveyardsSize}),
			OracleText: `
			Lord of Extinction's power and toughness are each equal to the number of cards in all graveyards.
		`,
		},
	}
}
