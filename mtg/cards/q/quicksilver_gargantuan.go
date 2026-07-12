package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuicksilverGargantuan is the card definition for Quicksilver Gargantuan.
//
// Type: Creature — Shapeshifter
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of any creature on the battlefield, except it's 7/7.
var QuicksilverGargantuan = newQuicksilverGargantuan

func newQuicksilverGargantuan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Quicksilver Gargantuan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyWithBasePowerToughness(game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield, except it's 7/7.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, nil, nil), 7, 7),
			},
			OracleText: `
			You may have this creature enter as a copy of any creature on the battlefield, except it's 7/7.
		`,
		},
	}
}
