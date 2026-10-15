package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StuntDouble is the card definition for Stunt Double.
//
// Type: Creature — Shapeshifter
// Cost: {3}{U}
//
// Oracle text:
//
//	Flash
//	You may have this creature enter as a copy of any creature on the battlefield.
var StuntDouble = newStuntDouble

func newStuntDouble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stunt Double",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			Flash
			You may have this creature enter as a copy of any creature on the battlefield.
		`,
		},
	}
}
