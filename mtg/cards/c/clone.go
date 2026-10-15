package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Clone is the card definition for Clone.
//
// Type: Creature — Shapeshifter
// Cost: {3}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of any creature on the battlefield.
var Clone = newClone

func newClone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Clone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this creature enter as a copy of any creature on the battlefield.
		`,
		},
	}
}
