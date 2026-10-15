package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DakmorSorceress is the card definition for Dakmor Sorceress.
//
// Type: Creature — Human Wizard Sorcerer
// Cost: {5}{B}
//
// Oracle text:
//
//	Dakmor Sorceress's power is equal to the number of Swamps you control.
var DakmorSorceress = newDakmorSorceress

func newDakmorSorceress() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dakmor Sorceress",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:       []color.Color{color.Black},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Wizard, types.Sorcerer},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			OracleText: `
			Dakmor Sorceress's power is equal to the number of Swamps you control.
		`,
		},
	}
}
