package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SylvanYeti is the card definition for Sylvan Yeti.
//
// Type: Creature — Yeti
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Sylvan Yeti's power is equal to the number of cards in your hand.
var SylvanYeti = newSylvanYeti

func newSylvanYeti() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sylvan Yeti",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:       []color.Color{color.Green},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Yeti},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			OracleText: `
			Sylvan Yeti's power is equal to the number of cards in your hand.
		`,
		},
	}
}
